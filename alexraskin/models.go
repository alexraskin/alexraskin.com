package alexraskin

import "html/template"

type PageData struct {
	Home      HomeData
	Error     string
	Status    int
	Path      string
	RequestID string
}

type HomeData struct {
	Content template.HTML
}

type LastFMTrack struct {
	Name      string
	Artist    string
	Album     string
	URL       string
	ArtistURL string
	Artwork   string
}

type LastFMResponse struct {
	Track      string   `json:"track"`
	Artist     string   `json:"artist"`
	Album      string   `json:"album"`
	NowPlaying bool     `json:"nowPlaying"`
	Image      []string `json:"image"`
}