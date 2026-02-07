package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/inovacc/omni/pkg/video/extractor"
	"github.com/inovacc/omni/pkg/video/nethttp"
	"github.com/inovacc/omni/pkg/video/types"
	"github.com/inovacc/omni/pkg/video/utils"
)

func init() {
	extractor.Register(&YoutubeSearchExtractor{})
}

// YoutubeSearchExtractor handles "ytsearch:QUERY" URLs.
type YoutubeSearchExtractor struct {
	extractor.BaseExtractor
}

// Name returns the extractor name.
func (e *YoutubeSearchExtractor) Name() string { return "YoutubeSearch" }

// Suitable returns true for ytsearch: prefixed queries.
func (e *YoutubeSearchExtractor) Suitable(rawURL string) bool {
	return strings.HasPrefix(rawURL, "ytsearch:")
}

// Extract performs a YouTube search and returns results.
func (e *YoutubeSearchExtractor) Extract(ctx context.Context, rawURL string, client *nethttp.Client) (*types.VideoInfo, error) {
	e.ExtractorName = "YoutubeSearch"

	query := strings.TrimPrefix(rawURL, "ytsearch:")
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, &types.ExtractorError{
			Extractor: "YoutubeSearch",
			Message:   "empty search query",
		}
	}

	// Use InnerTube search API.
	body := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    clientWeb.Name,
				"clientVersion": clientWeb.Version,
			},
		},
		"query": query,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := InnerTubeSearchURL(clientWeb.APIKey)

	data, err := client.PostJSON(ctx, apiURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, &types.ExtractorError{
			Extractor: "YoutubeSearch",
			Message:   "search request failed",
			Cause:     err,
		}
	}

	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	info := &types.VideoInfo{
		ID:           "ytsearch:" + query,
		Title:        "YouTube Search: " + query,
		Type:         "playlist",
		WebpageURL:   "https://www.youtube.com/results?search_query=" + query,
		Extractor:    "YoutubeSearch",
		ExtractorKey: "YoutubeSearch",
	}

	// Navigate the response to find video results.
	contents := utils.TraverseList(resp, "contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents")

	for _, section := range contents {
		sectionMap, ok := section.(map[string]any)
		if !ok {
			continue
		}

		items := utils.TraverseList(sectionMap, "itemSectionRenderer", "contents")
		for _, item := range items {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}

			renderer := utils.TraverseMap(itemMap, "videoRenderer")
			if renderer == nil {
				continue
			}

			entryVideoID := utils.TraverseString(renderer, "videoId")
			if entryVideoID == "" {
				continue
			}

			entryTitle := utils.TraverseString(renderer, "title", "runs", 0, "text")

			entry := types.VideoInfo{
				ID:           entryVideoID,
				Title:        entryTitle,
				Type:         "url",
				URL:          "https://www.youtube.com/watch?v=" + entryVideoID,
				WebpageURL:   "https://www.youtube.com/watch?v=" + entryVideoID,
				Extractor:    "YouTube",
				ExtractorKey: "YouTube",
			}

			// Extract duration text.
			durText := utils.TraverseString(renderer, "lengthText", "simpleText")
			if durText != "" {
				if d, ok := utils.ParseDuration(durText); ok {
					entry.Duration = d
				}
			}

			// Extract view count.
			viewText := utils.TraverseString(renderer, "viewCountText", "simpleText")
			if viewText != "" {
				entry.ViewCount = utils.IntOrNone(strings.ReplaceAll(strings.Split(viewText, " ")[0], ",", ""))
			}

			// Extract uploader.
			entry.Uploader = utils.TraverseString(renderer, "ownerText", "runs", 0, "text")

			info.Entries = append(info.Entries, entry)
		}
	}

	return info, nil
}
