package main

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/alexraskin/alexraskin.com/internal/server"
)

func TestEmbeddedPages(t *testing.T) {
	tmpl, err := template.New("").ParseFS(Templates, "templates/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	const cdn = "https://photos.example.com"
	reviews, err := server.LoadReviews(Data, cdn)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data any
	}{
		{"index.gohtml", server.PageData{Track: &server.LastFMTrack{Name: "Song", Artist: "Artist"}}},
		{"franzbroetchen.gohtml", server.ReviewsPageData{Reviews: reviews, CDNBase: cdn}},
		{"error.gohtml", server.PageData{Error: "Failed to render template", Status: 500}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var page bytes.Buffer
			if err := tmpl.ExecuteTemplate(&page, tc.name, tc.data); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(page.String(), "</html>") {
				t.Fatal("incomplete page")
			}
			if tc.name == "franzbroetchen.gohtml" && len(reviews) > 0 && !strings.Contains(page.String(), cdn+"/cdn-cgi/image/") {
				t.Fatal("photo CDN override missing")
			}
		})
	}
}
