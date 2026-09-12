package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testServer(t *testing.T) *Server {
	t.Helper()
	files := fstest.MapFS{"assets/test.css": {Data: []byte("body {}")}, "assets/robots.txt": {Data: []byte("User-agent: *")}}
	return NewServer(Config{
		Assets: http.FS(files),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"track":"Song"}`))}, nil
		})},
		ExecuteTemplate: func(w io.Writer, name string, data any) error {
			_, err := io.WriteString(w, "<html>"+name+"</html>")
			return err
		},
		Reviews: func() ([]Review, error) { return nil, nil },
	})
}

func TestTemplateErrorsAreBuffered(t *testing.T) {
	for _, failError := range []bool{false, true} {
		s := testServer(t)
		s.tmplFunc = func(w io.Writer, name string, data any) error {
			if name == "error.gohtml" && !failError {
				_, err := io.WriteString(w, "<html>Error</html>")
				return err
			}
			io.WriteString(w, "PARTIAL")
			return errors.New("broken template")
		}
		for _, path := range []string{"/", "/franzbroetchen"} {
			w := httptest.NewRecorder()
			s.Routes().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code != 500 || strings.Contains(w.Body.String(), "PARTIAL") || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
			}
		}
	}
}

func TestReadRoutes(t *testing.T) {
	h := testServer(t).Routes()
	for _, path := range []string{"/", "/franzbroetchen", "/robots.txt", "/assets/test.css", "/ping", "/version"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 200 || (method == "HEAD" && w.Body.Len() != 0) {
				t.Fatalf("%s %s: status=%d body=%s", method, path, w.Code, w.Body.String())
			}
			if w.Header().Get("ETag") != "" {
				t.Fatal("unexpected ETag")
			}
		}
	}
}

func TestRoutingFallbacks(t *testing.T) {
	h := testServer(t).Routes()
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/missing", 302},
		{"GET", "/assets/missing.css", 404},
		{"POST", "/", 405},
		{"POST", "/assets/test.css", 405},
		{"POST", "/robots.txt", 405},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status {
			t.Errorf("%s %s: got %d, want %d", tc.method, tc.path, w.Code, tc.status)
		}
	}
}

func TestLastFMCacheAndStaleFallback(t *testing.T) {
	s := testServer(t)
	var calls int
	s.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls > 1 {
			return nil, errors.New("upstream down")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"track":"Song"}`))}, nil
	})
	for i := 0; i < 2; i++ {
		track, err := s.lastFMTrack(context.Background())
		if err != nil || track.Name != "Song" {
			t.Fatalf("track=%v err=%v", track, err)
		}
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
	s.trackExpires = time.Time{}
	track, err := s.lastFMTrack(context.Background())
	if err == nil || track.Name != "Song" {
		t.Fatalf("track=%v err=%v", track, err)
	}
	s.lastFMTrack(context.Background())
	if calls != 2 {
		t.Fatalf("failure was not cached: %d", calls)
	}
}

func TestLastFMCancellationAndConcurrentRefresh(t *testing.T) {
	s := testServer(t)
	entered := make(chan struct{})
	var calls atomic.Int32
	s.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := s.lastFMTrack(ctx); done <- err }()
	<-entered
	if track, err := s.lastFMTrack(context.Background()); track != nil || err != nil {
		t.Fatalf("track=%v err=%v", track, err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("request cancellation did not propagate")
	}
	if calls.Load() != 1 || !s.trackExpires.IsZero() {
		t.Fatal("canceled refresh cached or duplicated")
	}
}

func TestReviewsRejectTrailingContent(t *testing.T) {
	for _, tc := range []struct {
		body    string
		wantErr bool
	}{{"[]", false}, {"[] \n\t", false}, {"[] {}", true}, {"[] garbage", true}, {`[{"unexpected":true}]`, true}} {
		_, err := LoadReviews(fstest.MapFS{reviewsPath: {Data: []byte(tc.body)}}, "https://cdn.example.com")
		if (err != nil) != tc.wantErr {
			t.Errorf("body=%q err=%v", tc.body, err)
		}
	}
}

func TestLastFMEmptyResultReplacesStaleTrack(t *testing.T) {
	s := testServer(t)
	s.track = &LastFMTrack{Name: "Old song"}
	s.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"track":""}`))}, nil
	})
	track, err := s.lastFMTrack(context.Background())
	if err != nil || track != nil {
		t.Fatalf("track=%v err=%v", track, err)
	}
}
