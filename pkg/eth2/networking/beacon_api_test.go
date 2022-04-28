package networking

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type handler func(rw http.ResponseWriter, req *http.Request)

func setupServer(handler handler) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestHealth(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		want     []HealthResponse
		handlers []handler
	}{
		{
			"Test Case 1, no endpoints, empty",
			nil,
			nil,
		},
		{
			"Test Case 2, empty endpoints, request failed",
			[]HealthResponse{
				{healthy: false, err: errors.New("")},
			},
			[]handler{nil},
		},
		{
			"Test Case 3, bad endpoint, 400 response",
			[]HealthResponse{
				{healthy: false, err: errors.New("")},
			},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusBadRequest)
				},
			},
		},
		{
			"Test Case 4, good endpoint, 200 response",
			[]HealthResponse{
				{healthy: true, err: nil},
			},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
				},
			},
		},
		{
			"Test Case 5, Node is syncing but can serve incomplete data, 206 response",
			[]HealthResponse{
				{healthy: false, err: errors.New("")},
			},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusPartialContent)
				},
			},
		},
		{
			"Test Case 6, Node not initialized or having issues, 503 response",
			[]HealthResponse{
				{healthy: false, err: errors.New("")},
			},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusServiceUnavailable)
				},
			},
		},
		{
			"Test Case 7, good endpoints, all healthy",
			[]HealthResponse{
				{healthy: true, err: nil},
				{healthy: true, err: nil},
				{healthy: true, err: nil},
			},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
				},
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
				},
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			endpoints := make([]string, 0)

			for _, handler := range tc.handlers {
				srv := setupServer(handler)
				defer srv.Close()
				endpoints = append(endpoints, srv.URL)
			}

			client := BeaconClient{
				RetryDuration: time.Millisecond * 100,
			}

			got := client.Health(endpoints)

			mask := make(map[int]bool)
			for _, g := range got {
				if !g.healthy && g.err == nil {
					t.Error("Unhealthy endpoint returned nil error")
					continue
				} else if g.healthy && g.err != nil {
					t.Error("Healthy endpoint returned non-nil error")
					continue
				}

				matched := false
				for i, w := range tc.want {
					if mask[i] {
						continue
					}
					if g.healthy == w.healthy {
						mask[i] = true
						matched = true
						break
					}
				}

				if !matched {
					t.Errorf("Got %v, want %v", got, tc.want)
				}
			}
		})
	}
}
