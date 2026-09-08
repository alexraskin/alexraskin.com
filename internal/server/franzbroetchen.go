package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
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

var displayWidths = []int{672, 1320}

type ReviewsFunc func() ([]Review, error)

func (r Review) Stars() string {
	return strings.Repeat("★", r.Rating) + strings.Repeat("☆", maxRating-r.Rating)
}

func (r Review) Day() string {
	return r.when.Format("2 January 2006")
}

func (p Photo) transform(width int) string {
	return fmt.Sprintf("%s/cdn-cgi/image/width=%d,format=auto,quality=82/%s", p.base, width, p.Key)
}

// Src is the widest rendered size: the URL for browsers that ignore srcset.
func (p Photo) Src() string {
	return p.transform(p.widths()[len(p.widths())-1])
}

// Srcset offers each width the stored photo can actually fill. Asking for more
// pixels than the object holds would bill a transformation to upscale it.
func (p Photo) Srcset() string {
	widths := p.widths()
	candidates := make([]string, len(widths))
	for i, width := range widths {
		candidates[i] = p.transform(width) + " " + strconv.Itoa(width) + "w"
	}
	return strings.Join(candidates, ", ")
}

func (p Photo) widths() []int {
	widths := make([]int, 0, len(displayWidths))
	for _, width := range displayWidths {
		if width <= p.Width {
			widths = append(widths, width)
		}
	}
	if len(widths) == 0 {
		return []int{p.Width}
	}
	return widths
}

// DisplayHeight is the height the widest rendered width implies, so the page can
// reserve the right box before the photo arrives and not shift the layout.
func (p Photo) DisplayHeight() int {
	widths := p.widths()
	return p.Height * widths[len(widths)-1] / p.Width
}

func (p Photo) DisplayWidth() int {
	widths := p.widths()
	return widths[len(widths)-1]
}

func LoadReviews(dataFS fs.FS, cdnBase string) ([]Review, error) {
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

	base := strings.TrimSuffix(cdnBase, "/")

	for i := range reviews {
		if err := reviews[i].load(base); err != nil {
			return nil, fmt.Errorf("reviews: entry %d: %w", i, err)
		}
	}

	sort.SliceStable(reviews, func(i, j int) bool {
		return reviews[i].when.After(reviews[j].when)
	})

	return reviews, nil
}

func (r *Review) load(base string) error {
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
		return fmt.Errorf("photos is required, with at least one entry")
	}
	for i := range r.Photos {
		if err := r.Photos[i].load(base); err != nil {
			return fmt.Errorf("photo %d: %w", i, err)
		}
	}

	return nil
}

func (p *Photo) load(base string) error {
	if p.Key == "" || strings.HasPrefix(p.Key, "/") || strings.Contains(p.Key, "://") {
		return fmt.Errorf("key %q must be a bucket key, not a path or URL", p.Key)
	}
	if p.Width <= 0 || p.Height <= 0 {
		return fmt.Errorf("key %q: width and height are required, run: mise run add-review <photo>", p.Key)
	}

	p.base = base

	return nil
}
