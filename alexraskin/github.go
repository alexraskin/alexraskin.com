package alexraskin

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

type cachedContent struct {
	content   string
	timestamp time.Time
}

var (
	cache    *cachedContent
	cacheMux sync.RWMutex
)

func (s *Server) fetchGitHubProfile() (string, error) {
	cacheMux.RLock()
	if cache != nil && time.Since(cache.timestamp) < 5*time.Minute {
		content := cache.content
		cacheMux.RUnlock()
		return content, nil
	}
	cacheMux.RUnlock()

	client := github.NewClient(s.httpClient)

	username := "alexraskin"
	repo := "alexraskin"

	content, _, _, err := client.Repositories.GetContents(s.ctx, username, repo, "README.md", nil)
	if err != nil {
		cacheMux.RLock()
		if cache != nil {
			content := cache.content
			cacheMux.RUnlock()
			return content, nil
		}
		cacheMux.RUnlock()
		return "", fmt.Errorf("failed to fetch README: %w", err)
	}

	if content.GetEncoding() != "base64" {
		return "", fmt.Errorf("unexpected content encoding: %s", content.GetEncoding())
	}

	decoded, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	cacheMux.Lock()
	cache = &cachedContent{
		content:   string(decoded),
		timestamp: time.Now(),
	}
	cacheMux.Unlock()

	return string(decoded), nil
}
