package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"

	"github.com/inovacc/omni/pkg/video/extractor"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

func init() {
	extractor.Register(&YoutubePlaylistExtractor{})
}

var youtubePlaylistRe = regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.|m\.)?youtube\.com/playlist\?list=([a-zA-Z0-9_-]+)`)

// YoutubePlaylistExtractor extracts videos from a YouTube playlist.
type YoutubePlaylistExtractor struct {
	extractor.BaseExtractor
}

// Name returns the extractor name.
func (e *YoutubePlaylistExtractor) Name() string { return "YoutubePlaylist" }

// Suitable returns true if the URL is a YouTube playlist.
func (e *YoutubePlaylistExtractor) Suitable(rawURL string) bool {
	return youtubePlaylistRe.MatchString(rawURL)
}

// Extract fetches playlist metadata and video entries.
func (e *YoutubePlaylistExtractor) Extract(ctx context.Context, rawURL string, client *nethttp.Client) (*types.VideoInfo, error) {
	e.ExtractorName = "YoutubePlaylist"

	m := youtubePlaylistRe.FindStringSubmatch(rawURL)
	if m == nil {
		return nil, &types.ExtractorError{
			Extractor: "YoutubePlaylist",
			Message:   "could not extract playlist ID",
		}
	}

	playlistID := m[1]

	// Use InnerTube browse API.
	body := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    clientWeb.Name,
				"clientVersion": clientWeb.Version,
			},
		},
		"browseId": "VL" + playlistID,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := InnerTubeBrowseURL(clientWeb.APIKey)

	data, err := client.PostJSON(ctx, apiURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, &types.ExtractorError{
			Extractor: "YoutubePlaylist",
			Message:   "InnerTube browse request failed",
			Cause:     err,
		}
	}

	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	info := &types.VideoInfo{
		ID:           playlistID,
		Type:         "playlist",
		WebpageURL:   "https://www.youtube.com/playlist?list=" + playlistID,
		Extractor:    "YoutubePlaylist",
		ExtractorKey: "YoutubePlaylist",
	}

	// Extract title.
	info.Title = utils.TraverseString(resp, "metadata", "playlistMetadataRenderer", "title")
	if info.Title == "" {
		info.PlaylistTitle = playlistID
	} else {
		info.PlaylistTitle = info.Title
	}

	// Extract entries from content.
	contents := utils.TraverseList(resp, "contents", "twoColumnBrowseResultsRenderer", "tabs", 0, "tabRenderer", "content", "sectionListRenderer", "contents", 0, "itemSectionRenderer", "contents", 0, "playlistVideoListRenderer", "contents")

	for idx, item := range contents {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		renderer := utils.TraverseMap(itemMap, "playlistVideoRenderer")
		if renderer == nil {
			continue
		}

		entryVideoID := utils.TraverseString(renderer, "videoId")
		if entryVideoID == "" {
			continue
		}

		entryTitle := utils.TraverseString(renderer, "title", "runs", 0, "text")
		if entryTitle == "" {
			entryTitle = utils.TraverseString(renderer, "title", "simpleText")
		}

		pIdx := idx + 1
		entry := types.VideoInfo{
			ID:            entryVideoID,
			Title:         entryTitle,
			Type:          "url",
			URL:           "https://www.youtube.com/watch?v=" + entryVideoID,
			WebpageURL:    "https://www.youtube.com/watch?v=" + entryVideoID,
			PlaylistTitle: info.PlaylistTitle,
			PlaylistIndex: &pIdx,
			Extractor:     "YouTube",
			ExtractorKey:  "YouTube",
		}

		info.Entries = append(info.Entries, entry)
	}

	return info, nil
}
