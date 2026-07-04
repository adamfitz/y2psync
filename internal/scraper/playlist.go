package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	playlistIDRe = regexp.MustCompile(`[&?]list=([a-zA-Z0-9_-]+)`)
)

type VideoInfo struct {
	VideoID string
	Title   string
}

type PlaylistResult struct {
	PlaylistID string
	Videos     []VideoInfo
}

func ExtractPlaylistID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	matches := playlistIDRe.FindStringSubmatch(u.String())
	if len(matches) < 2 {
		return "", fmt.Errorf("no playlist ID found in URL")
	}
	return matches[1], nil
}

func FetchPlaylistVideos(playlistURL string) (*PlaylistResult, error) {
	playlistID, err := ExtractPlaylistID(playlistURL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	pageURL := fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID)
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch playlist page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	html := string(body)

	jsonStr := extractInitialData(html)
	if jsonStr == "" {
		return nil, fmt.Errorf("could not find playlist data on page")
	}

	var root map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		return nil, fmt.Errorf("parse playlist data: %w", err)
	}

	videos := walkPlaylistData(root)
	if len(videos) == 0 {
		return nil, fmt.Errorf("no videos found in playlist")
	}

	result := &PlaylistResult{
		PlaylistID: playlistID,
		Videos:     videos,
	}
	return result, nil
}

func extractInitialData(html string) string {
	markers := []string{
		`window.ytInitialData = `,
		`var ytInitialData = `,
		`ytInitialData = `,
	}
	for _, marker := range markers {
		idx := strings.Index(html, marker)
		if idx == -1 {
			continue
		}
		start := idx + len(marker)
		braceIdx := strings.IndexByte(html[start:], '{')
		if braceIdx == -1 {
			continue
		}
		start += braceIdx

		depth := 0
		end := start
		inStr := false
		escaped := false
		for i := start; i < len(html); i++ {
			ch := html[i]
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inStr = !inStr
				continue
			}
			if inStr {
				continue
			}
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 {
					end = i + 1
					return html[start:end]
				}
			}
		}
	}
	return ""
}

func walkPlaylistData(root map[string]any) []VideoInfo {
	var videos []VideoInfo
	contents, _ := getNested(root, "contents", "twoColumnBrowseResultsRenderer", "tabs")
	tabs, ok := contents.([]any)
	if !ok {
		return nil
	}
	for _, tab := range tabs {
		tabMap, ok := tab.(map[string]any)
		if !ok {
			continue
		}
		content, _ := getNested(tabMap, "tabRenderer", "content", "sectionListRenderer", "contents")
		sections, ok := content.([]any)
		if !ok {
			continue
		}
		for _, section := range sections {
			sectionMap, ok := section.(map[string]any)
			if !ok {
				continue
			}
			itemsRaw, _ := getNested(sectionMap, "itemSectionRenderer", "contents")
			items, ok := itemsRaw.([]any)
			if !ok {
				continue
			}
			for _, item := range items {
				itemMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				lockup, ok := itemMap["lockupViewModel"].(map[string]any)
				if !ok {
					continue
				}
				contentType, _ := lockup["contentType"].(string)
				if contentType != "LOCKUP_CONTENT_TYPE_VIDEO" {
					continue
				}
				videoID, _ := lockup["contentId"].(string)
				if videoID == "" {
					continue
				}
				title := extractTitleFromLockup(lockup)

				videos = append(videos, VideoInfo{
					VideoID: videoID,
					Title:   title,
				})
			}
		}
	}
	return videos
}

func extractTitleFromLockup(lockup map[string]any) string {
	meta, ok := lockup["metadata"].(map[string]any)
	if !ok {
		return ""
	}
	metaVM, ok := meta["lockupMetadataViewModel"].(map[string]any)
	if !ok {
		return ""
	}
	title, ok := metaVM["title"].(map[string]any)
	if !ok {
		return ""
	}
	content, _ := title["content"].(string)
	return content
}

func getNested(data map[string]any, keys ...string) (any, bool) {
	current := data
	for i, key := range keys {
		val, ok := current[key]
		if !ok {
			return nil, false
		}
		if i == len(keys)-1 {
			return val, true
		}
		current, ok = val.(map[string]any)
		if !ok {
			return nil, false
		}
	}
	return nil, false
}
