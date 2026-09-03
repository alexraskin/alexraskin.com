package server

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"path"
	"sort"
	"strconv"
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

	// The sized copies found next to Photo, smallest first. Photo itself is a
	// logical name and need not exist: mise run add-review writes
	// "<stem>-<width>.avif" and "<stem>-<width>.jpg" beside it.
	AVIF []Variant `json:"-"`
	JPEG []Variant `json:"-"`

	// Filled in from the largest JPEG so the markup can reserve the right box
	// and the page doesn't reflow as photos load.
	Width  int `json:"-"`
	Height int `json:"-"`

	when time.Time
}

// Variant is one encoding of a photo at one width.
type Variant struct {
	URL   string
	Width int
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

// Fallback is the largest JPEG, for the <img> inside <picture>: what a browser
// without AVIF downloads, and what an old one that ignores srcset gets.
func (r Review) Fallback() Variant {
	return r.JPEG[len(r.JPEG)-1]
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

	stem := strings.TrimSuffix(r.Photo, path.Ext(r.Photo))
	if r.AVIF, err = findVariants(assetsFS, stem, ".avif"); err != nil {
		return err
	}
	if r.JPEG, err = findVariants(assetsFS, stem, ".jpg"); err != nil {
		return err
	}

	// AVIF is an optimisation; the JPEG is what every browser can read, so a
	// photo without one has nothing to show.
	if len(r.JPEG) == 0 {
		return fmt.Errorf("photo %q: no %s-<width>.jpg alongside it, run: mise run add-review <photo> %s",
			r.Photo, path.Base(stem), path.Base(stem))
	}

	file, err := assetsFS.Open(strings.TrimPrefix(r.Fallback().URL, "/"))
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

// findVariants collects "<stem>-<width><ext>" files, smallest first. Widths
// come from the filenames rather than a list in code, so adding a size is a
// matter of encoding one more file.
func findVariants(assetsFS fs.FS, stem, ext string) ([]Variant, error) {
	matches, err := fs.Glob(assetsFS, strings.TrimPrefix(stem, "/")+"-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("photo %q: %w", stem+ext, err)
	}

	variants := make([]Variant, 0, len(matches))
	for _, match := range matches {
		suffix := strings.TrimSuffix(match[strings.LastIndex(match, "-")+1:], ext)

		width, err := strconv.Atoi(suffix)
		if err != nil {
			// Not a sized variant, just a file that happens to sit alongside.
			continue
		}
		variants = append(variants, Variant{URL: "/" + match, Width: width})
	}

	sort.Slice(variants, func(i, j int) bool {
		return variants[i].Width < variants[j].Width
	})

	return variants, nil
}
