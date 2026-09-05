package server

import (
	"bytes"
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
	Name   string
	AVIF   []Variant
	JPEG   []Variant
	Width  int
	Height int
}

func (p *Photo) UnmarshalJSON(body []byte) error {
	var name string
	if err := json.Unmarshal(body, &name); err != nil {
		return err
	}
	p.Name = name
	return nil
}

type Variant struct {
	URL   string
	Width int
}

type ReviewsFunc func() ([]Review, error)

func (r Review) Stars() string {
	return strings.Repeat("★", r.Rating) + strings.Repeat("☆", maxRating-r.Rating)
}

func (r Review) Day() string {
	return r.when.Format("2 January 2006")
}

func (p Photo) Fallback() Variant {
	return p.JPEG[len(p.JPEG)-1]
}

func LoadReviews(dataFS, assetsFS fs.FS) ([]Review, error) {
	body, err := fs.ReadFile(dataFS, reviewsPath)
	if err != nil {
		return nil, fmt.Errorf("reviews: read %s: %w", reviewsPath, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	var reviews []Review
	if err := decoder.Decode(&reviews); err != nil {
		return nil, fmt.Errorf("reviews: parse %s: %w", reviewsPath, err)
	}

	for i := range reviews {
		if err := reviews[i].load(assetsFS); err != nil {
			return nil, fmt.Errorf("reviews: entry %d: %w", i, err)
		}
	}

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

	if len(r.Photos) == 0 {
		return fmt.Errorf("photos is required, with at least one path")
	}
	for i := range r.Photos {
		if err := r.Photos[i].load(assetsFS); err != nil {
			return fmt.Errorf("photo %d: %w", i, err)
		}
	}

	return nil
}

func (p *Photo) load(assetsFS fs.FS) error {
	if !strings.HasPrefix(p.Name, assetPrefix) {
		return fmt.Errorf("%q must live under %s", p.Name, assetPrefix)
	}

	stem := strings.TrimSuffix(p.Name, path.Ext(p.Name))

	var err error
	if p.AVIF, err = findVariants(assetsFS, stem, ".avif"); err != nil {
		return err
	}
	if p.JPEG, err = findVariants(assetsFS, stem, ".jpg"); err != nil {
		return err
	}

	if len(p.JPEG) == 0 {
		return fmt.Errorf("%q: no %s-<width>.jpg alongside it, run: mise run add-review <photo> %s",
			p.Name, path.Base(stem), path.Base(stem))
	}

	file, err := assetsFS.Open(strings.TrimPrefix(p.Fallback().URL, "/"))
	if err != nil {
		return fmt.Errorf("%q: %w", p.Name, err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("%q: %w", p.Name, err)
	}
	p.Width, p.Height = config.Width, config.Height

	return nil
}

func findVariants(assetsFS fs.FS, stem, ext string) ([]Variant, error) {
	prefix := strings.TrimPrefix(stem, "/")

	matches, err := fs.Glob(assetsFS, prefix+"-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("photo %q: %w", stem+ext, err)
	}

	variants := make([]Variant, 0, len(matches))
	for _, match := range matches {
		cut := strings.LastIndex(match, "-")
		suffix := strings.TrimSuffix(match[cut+1:], ext)

		if match[:cut] != prefix {
			continue
		}

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
