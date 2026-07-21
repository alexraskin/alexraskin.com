package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const lastFMURL = "https://lastfm.alexraskin.com/alexraskin"

// last.fm uses "+" rather than "%20" for spaces in its music URLs.
func lastFMPathSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "%20", "+")
}

func (s *Server) fetchLastFMTrack() (*LastFMTrack, error) {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, lastFMURL, nil)
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
