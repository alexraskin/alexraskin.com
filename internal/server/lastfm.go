package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const lastFMURL = "https://lastfm.alexraskin.com/alexraskin"

// last.fm uses "+" rather than "%20" for spaces in its music URLs.
func lastFMPathSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "%20", "+")
}

func (s *Server) fetchLastFMTrack(ctx context.Context) (*LastFMTrack, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lastFMURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lastfm: unexpected status %s", resp.Status)
	}

	var lfmResponse LastFMResponse
	if err := json.NewDecoder(resp.Body).Decode(&lfmResponse); err != nil {
		return nil, fmt.Errorf("lastfm: decode response: %w", err)
	}

	// The upstream returns the most recently scrobbled track when nothing is
	// playing, so an empty name is the only "no track" case.
	if lfmResponse.Track == "" {
		return nil, nil
	}

	var artwork string
	if len(lfmResponse.Image) > 0 {
		artwork = lfmResponse.Image[len(lfmResponse.Image)-1]
	}

	artist := lastFMPathSegment(lfmResponse.Artist)

	return &LastFMTrack{
		Name:       lfmResponse.Track,
		Artist:     lfmResponse.Artist,
		Album:      lfmResponse.Album,
		URL:        "https://www.last.fm/music/" + artist + "/_/" + lastFMPathSegment(lfmResponse.Track),
		ArtistURL:  "https://www.last.fm/music/" + artist,
		Artwork:    artwork,
		NowPlaying: lfmResponse.NowPlaying,
	}, nil
}

// Cache successes and briefly back off after upstream errors. Concurrent requests
// use the last result while one caller refreshes it, avoiding a request stampede.
func (s *Server) lastFMTrack(ctx context.Context) (*LastFMTrack, error) {
	s.trackMu.Lock()
	if time.Now().Before(s.trackExpires) || s.trackRefreshing {
		track := s.track
		s.trackMu.Unlock()
		return track, nil
	}
	s.trackRefreshing = true
	s.trackMu.Unlock()
	track, err := s.fetchLastFMTrack(ctx)
	s.trackMu.Lock()
	defer s.trackMu.Unlock()
	s.trackRefreshing = false
	if err == nil {
		s.track = track
	}
	if ctx.Err() == nil {
		s.trackExpires = time.Now().Add(30 * time.Second)
	}
	return s.track, err
}
