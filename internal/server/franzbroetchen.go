package server

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"sort"
	"strings"
	"time"
)

const (
	reviewsPath   = "data/franzbroetchen.json"
	maxRating     = 5
	reviewDateFmt = "2006-01-02"
)

// Review is one Franzbrötchen, as written by hand in data/franzbroetchen.json.
// Adding one means appending an object there; nothing else has to change.
type Review struct {
	Place    string `json:"place"`
	Location string `json:"location"`
	Date     string `json:"date"`
	Rating   int    `json:"rating"`
	Photo    string `json:"photo"`
	Note     string `json:"note,omitempty"`
	URL      string `json:"url,omitempty"`

	// Filled in from the image itself so the markup can reserve the right box
	// and the page doesn't reflow as photos load.
	Width  int `json:"-"`
	Height int `json:"-"`

	when time.Time
}

// ReviewsFunc returns the reviews to render. Dev mode re-reads the file per
// request so a new entry shows up on refresh; the embedded build parses once.
type ReviewsFunc func() ([]Review, error)

// Stars renders the rating as filled and empty stars. The numeric rating is
// what screen readers get, from the label in the template.
func (r Review) Stars() string {
	return strings.Repeat("★", r.Rating) + strings.Repeat("☆", maxRating-r.Rating)
}

// Day is the date as prose, e.g. "24 August 2025".
func (r Review) Day() string {
	return r.when.Format("2 January 2006")
}

// LoadReviews reads the review file and measures each photo. Anything wrong
// with an entry is an error rather than a half-rendered page: in the embedded
// build that fails the process at startup, where it's obvious.
func LoadReviews(dataFS, assetsFS fs.FS) ([]Review, error) {
	body, err := fs.ReadFile(dataFS, reviewsPath)
	if err != nil {
		return nil, fmt.Errorf("reviews: read %s: %w", reviewsPath, err)
	}

	var reviews []Review
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, fmt.Errorf("reviews: parse %s: %w", reviewsPath, err)
	}

	for i := range reviews {
		if err := reviews[i].load(assetsFS); err != nil {
			return nil, fmt.Errorf("reviews: entry %d: %w", i, err)
		}
	}

	// Newest first.
	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].when.After(reviews[j].when)
	})

	return reviews, nil
}

func (r *Review) load(assetsFS fs.FS) error {
	if r.Place == "" {
		return fmt.Errorf("place is required")
	}
	if r.Rating < 1 || r.Rating > maxRating {
		return fmt.Errorf("rating %d is outside 1-%d", r.Rating, maxRating)
	}

	when, err := time.Parse(reviewDateFmt, r.Date)
	if err != nil {
		return fmt.Errorf("date %q is not %s", r.Date, reviewDateFmt)
	}
	r.when = when

	if !strings.HasPrefix(r.Photo, assetPrefix) {
		return fmt.Errorf("photo %q must live under %s", r.Photo, assetPrefix)
	}

	file, err := assetsFS.Open(strings.TrimPrefix(r.Photo, "/"))
	if err != nil {
		return fmt.Errorf("photo %q: %w", r.Photo, err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("photo %q: %w", r.Photo, err)
	}
	r.Width, r.Height = config.Width, config.Height

	return nil
}
