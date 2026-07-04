package ui

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	channelIDRe    = regexp.MustCompile(`youtube\.com/channel/(UC[a-zA-Z0-9_-]{22})`)
	channelHandleRe = regexp.MustCompile(`youtube\.com/@([a-zA-Z0-9_-]+)`)
	videoIDRe      = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([a-zA-Z0-9_-]{11})`)
)

func extractChannelID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if m := channelIDRe.FindStringSubmatch(u.String()); len(m) >= 2 {
		return m[1], nil
	}

	if m := videoIDRe.FindStringSubmatch(u.String()); len(m) >= 2 {
		return m[1], nil
	}

	if m := channelHandleRe.FindStringSubmatch(u.String()); len(m) >= 2 {
		return m[1], nil
	}

	if strings.HasPrefix(rawURL, "UC") && len(rawURL) >= 24 {
		return rawURL, nil
	}

	return "", fmt.Errorf("could not extract channel ID from URL")
}
