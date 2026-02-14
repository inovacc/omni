package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/cache"
	"github.com/inovacc/omni/pkg/video/extractor"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

func init() {
	extractor.Register(&YoutubeExtractor{})
}

var (
	youtubeValidURLRe = regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.|m\.)?(?:youtube\.com/(?:watch\?.*v=|embed/|v/|shorts/)|youtu\.be/)([a-zA-Z0-9_-]{11})`)
)

// YoutubeExtractor extracts video info from YouTube.
type YoutubeExtractor struct {
	extractor.BaseExtractor

	cache *cache.Cache
}

// Name returns the extractor name.
func (e *YoutubeExtractor) Name() string { return "YouTube" }

// Suitable returns true if the URL is a YouTube video URL.
func (e *YoutubeExtractor) Suitable(rawURL string) bool {
	return youtubeValidURLRe.MatchString(rawURL)
}

// Extract fetches video metadata from YouTube.
func (e *YoutubeExtractor) Extract(ctx context.Context, rawURL string, client *nethttp.Client) (*types.VideoInfo, error) {
	e.ExtractorName = "YouTube"

	if e.cache == nil {
		e.cache = cache.New("")
	}

	videoID := extractVideoID(rawURL)
	if videoID == "" {
		return nil, &types.ExtractorError{
			Extractor: "YouTube",
			Message:   "could not extract video ID from URL",
		}
	}

	// Try InnerTube API with different clients.
	var (
		playerResponse map[string]any
		lastReason     string
		debugInfo      []string
	)

	for _, cfg := range clientOrder {
		resp, err := e.callInnerTubePlayer(ctx, client, videoID, cfg)
		if err != nil {
			debugInfo = append(debugInfo, cfg.Name+": "+err.Error())
			continue
		}

		// Check for playability errors.
		playability := utils.TraverseMap(resp, "playabilityStatus")
		status := utils.TraverseString(playability, "status")
		lastReason = utils.TraverseString(playability, "reason")
		debugInfo = append(debugInfo, cfg.Name+": "+status+" "+lastReason)

		if status == "OK" || status == "CONTENT_CHECK_REQUIRED" {
			playerResponse = resp

			break
		}

		if status == "LOGIN_REQUIRED" {
			if strings.Contains(lastReason, "age") {
				continue // Try next client.
			}
		}
	}

	if playerResponse == nil {
		msg := "all InnerTube clients failed"
		if len(debugInfo) > 0 {
			msg += " [" + strings.Join(debugInfo, "; ") + "]"
		}

		return nil, &types.ExtractorError{
			Extractor: "YouTube",
			VideoID:   videoID,
			Message:   msg,
		}
	}

	return e.parsePlayerResponse(ctx, client, videoID, playerResponse)
}

func (e *YoutubeExtractor) callInnerTubePlayer(ctx context.Context, client *nethttp.Client, videoID string, cfg ClientConfig) (map[string]any, error) {
	body := InnerTubeRequest(videoID, cfg)

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := InnerTubePlayerURL(cfg.APIKey)

	// Compute SAPISIDHASH for authenticated requests if cookies are available.
	var sapisidHash string
	if jar := client.CookieJar(); jar != nil {
		origin := "https://www.youtube.com"
		if sapisid := nethttp.ExtractSAPISID(jar, origin); sapisid != "" {
			sapisidHash = nethttp.ComputeSAPISIDHash(sapisid, origin)
		}
	}

	headers := InnerTubeHeaders(videoID, cfg, sapisidHash)
	headers["User-Agent"] = cfg.UserAgent

	data, err := client.PostJSON(ctx, apiURL, bytes.NewReader(bodyJSON), headers)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *YoutubeExtractor) parsePlayerResponse(ctx context.Context, client *nethttp.Client, videoID string, resp map[string]any) (*types.VideoInfo, error) {
	videoDetails := utils.TraverseMap(resp, "videoDetails")
	microformat := utils.TraverseMap(resp, "microformat", "playerMicroformatRenderer")

	info := &types.VideoInfo{
		ID:           videoID,
		Title:        utils.TraverseString(videoDetails, "title"),
		Description:  utils.TraverseString(videoDetails, "shortDescription"),
		Uploader:     utils.TraverseString(videoDetails, "author"),
		ChannelID:    utils.TraverseString(videoDetails, "channelId"),
		Duration:     floatFromStr(utils.TraverseString(videoDetails, "lengthSeconds")),
		ViewCount:    utils.TraverseInt(videoDetails, "viewCount"),
		WebpageURL:   "https://www.youtube.com/watch?v=" + videoID,
		Extractor:    "YouTube",
		ExtractorKey: "YouTube",
	}

	if info.ChannelID != "" {
		info.ChannelURL = "https://www.youtube.com/channel/" + info.ChannelID
		info.UploaderURL = info.ChannelURL
		info.UploaderID = info.ChannelID
	}

	// Upload date.
	if date := utils.TraverseString(microformat, "uploadDate"); date != "" {
		info.UploadDate = strings.ReplaceAll(date, "-", "")
	}

	// Category.
	if cat := utils.TraverseString(microformat, "category"); cat != "" {
		info.Categories = []string{cat}
	}

	// Tags.
	if tags := utils.TraverseList(videoDetails, "keywords"); tags != nil {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				info.Tags = append(info.Tags, s)
			}
		}
	}

	// Is live.
	if live := utils.TraverseObj(videoDetails, "isLive"); live != nil {
		if b, ok := live.(bool); ok {
			info.IsLive = &b
		}
	}

	// Thumbnails.
	if thumbList := utils.TraverseList(videoDetails, "thumbnail", "thumbnails"); thumbList != nil {
		for _, t := range thumbList {
			if tm, ok := t.(map[string]any); ok {
				thumb := types.Thumbnail{
					URL: utils.TraverseString(tm, "url"),
				}
				if w := utils.TraverseInt(tm, "width"); w != nil {
					wi := int(*w)
					thumb.Width = &wi
				}

				if h := utils.TraverseInt(tm, "height"); h != nil {
					hi := int(*h)
					thumb.Height = &hi
				}

				if thumb.URL != "" {
					info.Thumbnails = append(info.Thumbnails, thumb)
				}
			}
		}
	}

	// Extract formats.
	formats, err := e.extractFormats(ctx, client, resp, videoID)
	if err != nil {
		return nil, err
	}

	info.Formats = formats

	// Sort formats.
	e.SortFormats(info.Formats)

	return info, nil
}

func (e *YoutubeExtractor) extractFormats(ctx context.Context, client *nethttp.Client, resp map[string]any, videoID string) ([]types.Format, error) {
	streamingData := utils.TraverseMap(resp, "streamingData")
	if streamingData == nil {
		return nil, &types.ExtractorError{
			Extractor: "YouTube",
			VideoID:   videoID,
			Message:   "no streaming data found",
		}
	}

	var formats []types.Format

	// Process regular formats.
	regularFormats := utils.TraverseList(streamingData, "formats")
	for _, f := range regularFormats {
		if fm, ok := f.(map[string]any); ok {
			if vf := e.parseFormat(fm); vf != nil {
				formats = append(formats, *vf)
			}
		}
	}

	// Process adaptive formats.
	adaptiveFormats := utils.TraverseList(streamingData, "adaptiveFormats")
	for _, f := range adaptiveFormats {
		if fm, ok := f.(map[string]any); ok {
			if vf := e.parseFormat(fm); vf != nil {
				formats = append(formats, *vf)
			}
		}
	}

	// Process HLS manifest.
	if hlsURL := utils.TraverseString(streamingData, "hlsManifestUrl"); hlsURL != "" {
		hlsFormats, err := e.ExtractM3U8Formats(ctx, client, hlsURL, videoID)
		if err == nil {
			formats = append(formats, hlsFormats...)
		}
	}

	// Handle signature decryption if needed.
	needsSigDecrypt := false

	for _, f := range formats {
		if f.URL == "" || strings.Contains(f.URL, "signature") {
			needsSigDecrypt = true
			break
		}
	}

	if needsSigDecrypt {
		decryptor := NewSignatureDecryptor(e.cache)
		// Try to get player URL from the watch page.
		pageHTML, err := client.GetString(ctx, "https://www.youtube.com/watch?v="+videoID)
		if err == nil {
			playerURL := decryptor.ExtractPlayerURL(pageHTML)
			if playerURL != "" {
				if err := decryptor.LoadPlayer(ctx, client, playerURL); err == nil {
					formats = e.decryptFormats(formats, decryptor)
				}
			}
		}
	}

	if len(formats) == 0 {
		return nil, &types.ExtractorError{
			Extractor: "YouTube",
			VideoID:   videoID,
			Message:   "no formats found",
		}
	}

	return formats, nil
}

func (e *YoutubeExtractor) parseFormat(fm map[string]any) *types.Format {
	formatURL := utils.TraverseString(fm, "url")
	signatureCipher := utils.TraverseString(fm, "signatureCipher")
	cipher := utils.TraverseString(fm, "cipher")

	// Handle signature cipher.
	if formatURL == "" && (signatureCipher != "" || cipher != "") {
		sc := signatureCipher
		if sc == "" {
			sc = cipher
		}

		params := utils.ParseQueryString(sc)
		formatURL = params["url"]
		// Signature will be decrypted later.
	}

	if formatURL == "" {
		return nil
	}

	f := &types.Format{
		URL:      formatURL,
		FormatID: utils.TraverseString(fm, "itag"),
		Protocol: "https",
	}

	// Mime type parsing.
	mimeType := utils.TraverseString(fm, "mimeType")
	if mimeType != "" {
		parseMimeType(mimeType, f)
	}

	// Quality label.
	f.FormatNote = utils.TraverseString(fm, "qualityLabel")

	// Dimensions.
	if w := utils.TraverseInt(fm, "width"); w != nil {
		wi := int(*w)
		f.Width = &wi
	}

	if h := utils.TraverseInt(fm, "height"); h != nil {
		hi := int(*h)
		f.Height = &hi
	}

	// Bitrate.
	if br := utils.TraverseFloat(fm, "bitrate"); br != nil {
		tbr := *br / 1000.0
		f.TBR = &tbr
	}

	// Filesize.
	if fs := utils.TraverseInt(fm, "contentLength"); fs != nil {
		f.Filesize = fs
	}

	// FPS.
	if fps := utils.TraverseFloat(fm, "fps"); fps != nil {
		f.FPS = fps
	}

	// Audio sample rate.
	if asr := utils.TraverseInt(fm, "audioSampleRate"); asr != nil {
		asri := int(*asr)
		f.ASR = &asri
	}

	// Audio channels.
	if ch := utils.TraverseFloat(fm, "audioChannels"); ch != nil {
		abr := *ch
		f.ABR = &abr
	}

	return f
}

func parseMimeType(mime string, f *types.Format) {
	// Example: "video/mp4; codecs=\"avc1.640028, mp4a.40.2\""
	parts := strings.SplitN(mime, ";", 2)
	typeParts := strings.SplitN(strings.TrimSpace(parts[0]), "/", 2)

	if len(typeParts) == 2 {
		mediaType := typeParts[0]
		container := typeParts[1]

		f.Ext = container
		if container == "x-flv" {
			f.Ext = "flv"
		}

		if mediaType == "audio" {
			f.VCodec = "none"
		}
	}

	if len(parts) > 1 {
		codecStr := strings.TrimSpace(parts[1])
		codecStr = strings.TrimPrefix(codecStr, "codecs=\"")

		codecStr = strings.TrimSuffix(codecStr, "\"")
		for c := range strings.SplitSeq(codecStr, ",") {
			c = strings.TrimSpace(c)
			switch {
			case strings.HasPrefix(c, "avc1"), strings.HasPrefix(c, "av01"),
				strings.HasPrefix(c, "vp9"), strings.HasPrefix(c, "vp09"),
				strings.HasPrefix(c, "hev1"), strings.HasPrefix(c, "hvc1"):
				f.VCodec = c
			case strings.HasPrefix(c, "mp4a"), strings.HasPrefix(c, "opus"),
				strings.HasPrefix(c, "vorbis"), strings.HasPrefix(c, "ac-3"),
				strings.HasPrefix(c, "ec-3"), strings.HasPrefix(c, "flac"):
				f.ACodec = c
			}
		}
	}
}

func (e *YoutubeExtractor) decryptFormats(formats []types.Format, decryptor *SignatureDecryptor) []types.Format {
	result := make([]types.Format, 0, len(formats))
	for _, f := range formats {
		// Decrypt n parameter (throttling).
		if f.URL != "" {
			u, err := url.Parse(f.URL)
			if err == nil {
				n := u.Query().Get("n")
				if n != "" {
					decrypted, err := decryptor.DecryptNsig(n)
					if err == nil && decrypted != n {
						q := u.Query()
						q.Set("n", decrypted)
						u.RawQuery = q.Encode()
						f.URL = u.String()
					}
				}
			}
		}

		result = append(result, f)
	}

	return result
}

func extractVideoID(rawURL string) string {
	m := youtubeValidURLRe.FindStringSubmatch(rawURL)
	if m != nil {
		return m[1]
	}
	// Try query parameter.
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Query().Get("v")
}

func floatFromStr(s string) float64 {
	if s == "" {
		return 0
	}

	var n float64

	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + float64(c-'0')
		} else if c == '.' {
			// Simple: ignore decimals for duration.
			break
		}
	}

	return n
}
