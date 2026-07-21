package alexraskin

type PageData struct {
	Error     string
	Status    int
	Path      string
	RequestID string
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
