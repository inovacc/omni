package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/extractor"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

func init() {
	extractor.Register(&YoutubeChannelExtractor{})
}

var youtubeChannelRe = regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.|m\.)?youtube\.com/(?:channel/([a-zA-Z0-9_-]+)|(@[a-zA-Z0-9_.-]+)|c/([a-zA-Z0-9_.-]+))(?:/videos?)?/?(?:\?.*)?$`)

// videosTabParams is a well-known InnerTube parameter that selects the
// Videos tab sorted by "Latest" on a channel page.
const videosTabParams = "EgZ2aWRlb3PyBgQKAjoA"

// YoutubeChannelExtractor extracts all videos from a YouTube channel.
type YoutubeChannelExtractor struct {
	extractor.BaseExtractor
}

// Name returns the extractor name.
func (e *YoutubeChannelExtractor) Name() string { return "YoutubeChannel" }

// Suitable returns true if the URL is a YouTube channel URL.
func (e *YoutubeChannelExtractor) Suitable(rawURL string) bool {
	return youtubeChannelRe.MatchString(rawURL)
}

// Extract fetches channel metadata and all video entries with pagination.
func (e *YoutubeChannelExtractor) Extract(ctx context.Context, rawURL string, client *nethttp.Client) (*types.VideoInfo, error) {
	e.ExtractorName = "YoutubeChannel"

	channelID, err := e.resolveChannelID(ctx, client, rawURL)
	if err != nil {
		return nil, err
	}

	// First request: browse the Videos tab.
	resp, err := e.browseChannel(ctx, client, channelID, "")
	if err != nil {
		return nil, err
	}

	info := &types.VideoInfo{
		ID:           channelID,
		Type:         "playlist",
		WebpageURL:   "https://www.youtube.com/channel/" + channelID,
		Extractor:    "YoutubeChannel",
		ExtractorKey: "YoutubeChannel",
		Metadata:     make(map[string]string),
	}

	// Extract channel metadata.
	e.extractChannelMetadata(resp, info)

	// Extract initial video entries + continuation token.
	entries, contToken := e.extractInitialEntries(resp)
	info.Entries = append(info.Entries, entries...)

	// Paginate until no more continuation tokens.
	for contToken != "" {
		select {
		case <-ctx.Done():
			return info, ctx.Err()
		default:
		}

		contResp, contErr := e.browseChannel(ctx, client, "", contToken)
		if contErr != nil {
			break
		}

		newEntries, nextToken := e.extractContinuationEntries(contResp)
		if len(newEntries) == 0 {
			break
		}

		info.Entries = append(info.Entries, newEntries...)
		contToken = nextToken
	}

	return info, nil
}

// resolveChannelID converts any channel URL form to a channel ID (UC...).
func (e *YoutubeChannelExtractor) resolveChannelID(ctx context.Context, client *nethttp.Client, rawURL string) (string, error) {
	m := youtubeChannelRe.FindStringSubmatch(rawURL)
	if m == nil {
		return "", &types.ExtractorError{
			Extractor: "YoutubeChannel",
			Message:   "could not parse channel URL",
		}
	}

	// Group 1: /channel/UCxxx -- use directly.
	if m[1] != "" {
		return m[1], nil
	}

	// Group 2 (@handle) or Group 3 (/c/name) -- need to resolve via page HTML.
	html, err := client.GetString(ctx, normalizeChannelURL(rawURL))
	if err != nil {
		return "", &types.ExtractorError{
			Extractor: "YoutubeChannel",
			Message:   "failed to fetch channel page",
			Cause:     err,
		}
	}

	// Try <link rel="canonical" href="...channel/UCxxx">.
	canonicalRe := regexp.MustCompile(`<link\s+rel="canonical"\s+href="https?://www\.youtube\.com/channel/([a-zA-Z0-9_-]+)"`)
	if cm := canonicalRe.FindStringSubmatch(html); cm != nil {
		return cm[1], nil
	}

	// Try ytInitialData JSON blob.
	dataRe := regexp.MustCompile(`var\s+ytInitialData\s*=\s*(\{.+?\});\s*</script>`)
	if dm := dataRe.FindStringSubmatch(html); dm != nil {
		var data map[string]any
		if err := json.Unmarshal([]byte(dm[1]), &data); err == nil {
			if chID := utils.TraverseString(data, "metadata", "channelMetadataRenderer", "externalId"); chID != "" {
				return chID, nil
			}
		}
	}

	// Fallback: browse endpoint pattern in HTML.
	browseRe := regexp.MustCompile(`"browseId"\s*:\s*"(UC[a-zA-Z0-9_-]+)"`)
	if bm := browseRe.FindStringSubmatch(html); bm != nil {
		return bm[1], nil
	}

	return "", &types.ExtractorError{
		Extractor: "YoutubeChannel",
		Message:   "could not resolve channel ID from URL",
	}
}

// browseChannel calls the InnerTube browse API.
// If continuation is non-empty, it uses continuation; otherwise browseId + params.
func (e *YoutubeChannelExtractor) browseChannel(ctx context.Context, client *nethttp.Client, channelID, continuation string) (map[string]any, error) {
	body := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    clientWeb.Name,
				"clientVersion": clientWeb.Version,
			},
		},
	}

	if continuation != "" {
		body["continuation"] = continuation
	} else {
		body["browseId"] = channelID
		body["params"] = videosTabParams
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := InnerTubeBrowseURL(clientWeb.APIKey)

	data, err := client.PostJSON(ctx, apiURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, &types.ExtractorError{
			Extractor: "YoutubeChannel",
			Message:   "InnerTube browse request failed",
			Cause:     err,
		}
	}

	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// extractChannelMetadata populates info with channel-level metadata.
func (e *YoutubeChannelExtractor) extractChannelMetadata(resp map[string]any, info *types.VideoInfo) {
	meta := utils.TraverseMap(resp, "metadata", "channelMetadataRenderer")
	if meta != nil {
		info.Title = utils.TraverseString(meta, "title")
		info.Channel = info.Title
		info.Description = utils.TraverseString(meta, "description")
		info.ChannelURL = utils.TraverseString(meta, "channelUrl")
		info.ChannelID = utils.TraverseString(meta, "externalId")
		if info.ChannelID != "" {
			info.ID = info.ChannelID
		}
	}

	header := utils.TraverseMap(resp, "header", "c4TabbedHeaderRenderer")
	if header == nil {
		// Newer YouTube layout uses pageHeaderRenderer.
		header = utils.TraverseMap(resp, "header", "pageHeaderRenderer")
	}

	if header != nil {
		if sub := utils.TraverseString(header, "subscriberCountText", "simpleText"); sub != "" {
			info.Metadata["subscriber_count"] = sub
		}

		// Avatar from c4TabbedHeaderRenderer.
		if avatar := utils.TraverseString(header, "avatar", "thumbnails", 0, "url"); avatar != "" {
			info.Metadata["avatar_url"] = avatar
		}
	}

	info.Uploader = info.Title
	info.UploaderID = info.ChannelID
	info.UploaderURL = info.ChannelURL
}

// extractInitialEntries extracts video entries from the initial browse response.
// Returns entries and a continuation token (empty string if no more pages).
func (e *YoutubeChannelExtractor) extractInitialEntries(resp map[string]any) ([]types.VideoInfo, string) {
	// Navigate to the Videos tab content.
	tabs := utils.TraverseList(resp, "contents", "twoColumnBrowseResultsRenderer", "tabs")

	var contents []any
	for _, tab := range tabs {
		tabMap, ok := tab.(map[string]any)
		if !ok {
			continue
		}

		renderer := utils.TraverseMap(tabMap, "tabRenderer")
		if renderer == nil {
			continue
		}

		// Look for the "Videos" tab (or any tab with richGridRenderer).
		richGrid := utils.TraverseList(renderer, "content", "richGridRenderer", "contents")
		if richGrid != nil {
			contents = richGrid
			break
		}
	}

	if contents == nil {
		return nil, ""
	}

	return e.parseGridContents(contents)
}

// extractContinuationEntries extracts entries from a continuation response.
func (e *YoutubeChannelExtractor) extractContinuationEntries(resp map[string]any) ([]types.VideoInfo, string) {
	actions := utils.TraverseList(resp, "onResponseReceivedActions")
	if actions == nil {
		return nil, ""
	}

	for _, action := range actions {
		actionMap, ok := action.(map[string]any)
		if !ok {
			continue
		}

		items := utils.TraverseList(actionMap, "appendContinuationItemsAction", "continuationItems")
		if items != nil {
			return e.parseGridContents(items)
		}
	}

	return nil, ""
}

// parseGridContents extracts video entries and continuation token from grid items.
func (e *YoutubeChannelExtractor) parseGridContents(contents []any) ([]types.VideoInfo, string) {
	var entries []types.VideoInfo
	var contToken string

	for _, item := range contents {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Video entry: richItemRenderer > content > videoRenderer.
		videoRenderer := utils.TraverseMap(itemMap, "richItemRenderer", "content", "videoRenderer")
		if videoRenderer != nil {
			entry := e.parseVideoRenderer(videoRenderer)
			if entry != nil {
				entries = append(entries, *entry)
			}

			continue
		}

		// Continuation token.
		token := utils.TraverseString(itemMap, "continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token")
		if token != "" {
			contToken = token
		}
	}

	return entries, contToken
}

// parseVideoRenderer extracts a single video entry from a videoRenderer object.
func (e *YoutubeChannelExtractor) parseVideoRenderer(renderer map[string]any) *types.VideoInfo {
	videoID := utils.TraverseString(renderer, "videoId")
	if videoID == "" {
		return nil
	}

	title := utils.TraverseString(renderer, "title", "runs", 0, "text")
	if title == "" {
		title = utils.TraverseString(renderer, "title", "simpleText")
	}

	entry := &types.VideoInfo{
		ID:           videoID,
		Title:        title,
		Type:         "url",
		URL:          "https://www.youtube.com/watch?v=" + videoID,
		WebpageURL:   "https://www.youtube.com/watch?v=" + videoID,
		Extractor:    "YouTube",
		ExtractorKey: "YouTube",
	}

	// Duration text (e.g., "12:34").
	if durText := utils.TraverseString(renderer, "lengthText", "simpleText"); durText != "" {
		entry.Duration = parseDurationText(durText)
	}

	// View count.
	if vc := utils.TraverseString(renderer, "viewCountText", "simpleText"); vc != "" {
		entry.ViewCount = parseViewCount(vc)
	}

	// Upload date (relative, e.g., "2 days ago").
	if dateText := utils.TraverseString(renderer, "publishedTimeText", "simpleText"); dateText != "" {
		entry.UploadDate = dateText
	}

	// Thumbnail.
	if thumbURL := utils.TraverseString(renderer, "thumbnail", "thumbnails", 0, "url"); thumbURL != "" {
		entry.Thumbnails = []types.Thumbnail{{URL: thumbURL}}
	}

	return entry
}

// parseDurationText converts "H:MM:SS" or "MM:SS" to seconds.
func parseDurationText(s string) float64 {
	parts := strings.Split(s, ":")
	var total float64

	for _, p := range parts {
		total = total*60 + floatFromStr(p)
	}

	return total
}

// parseViewCount extracts a number from strings like "1,234 views".
func parseViewCount(s string) *int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}

	if n == 0 && !strings.Contains(s, "0") {
		return nil
	}

	return &n
}

// normalizeChannelURL ensures the URL is a full HTTPS URL.
func normalizeChannelURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}

	return rawURL
}
