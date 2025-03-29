package alexraskin

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) fetchLastFMTrack() (*LastFMTrack, error) {
	lastFMURL := "https://lastfm.alexraskin.com/twizycat"

	resp, err := http.Get(lastFMURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lfmResponse LastFMResponse
	if err := json.Unmarshal(body, &lfmResponse); err != nil {
		return nil, err
	}

	if !lfmResponse.NowPlaying {
		return nil, nil
	}

	var artwork string
	if len(lfmResponse.Image) > 0 {
		artwork = lfmResponse.Image[len(lfmResponse.Image)-1]
	}

	return &LastFMTrack{
		Name:      lfmResponse.Track,
		Artist:    lfmResponse.Artist,
		Album:     lfmResponse.Album,
		URL:       "https://www.last.fm/music/" + lfmResponse.Artist + "/_/" + lfmResponse.Track,
		ArtistURL: "https://www.last.fm/music/" + lfmResponse.Artist,
		Artwork:   artwork,
	}, nil
}
