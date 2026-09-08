package server

import "time"

type PageData struct {
	Error  string
	Status int
	Track  *LastFMTrack
}

type ReviewsPageData struct {
	Reviews []Review
	CDNBase string
}

type Review struct {
	Place    string  `json:"place"`
	Location string  `json:"location"`
	Date     string  `json:"date"`
	Rating   int     `json:"rating"`
	Photos   []Photo `json:"photos"`
	Note     string  `json:"note,omitempty"`
	URL      string  `json:"url,omitempty"`

	when time.Time
}

type Photo struct {
	Key    string `json:"key"`
	Width  int    `json:"width"`
	Height int    `json:"height"`

	base string
}

type LastFMTrack struct {
	Name       string
	Artist     string
	Album      string
	URL        string
	ArtistURL  string
	Artwork    string
	NowPlaying bool
}

type LastFMResponse struct {
	Track      string   `json:"track"`
	Artist     string   `json:"artist"`
	Album      string   `json:"album"`
	NowPlaying bool     `json:"nowPlaying"`
	Image      []string `json:"image"`
}
