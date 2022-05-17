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
				{Healthy: false, Error: errors.New("")},
			},
			[]handler{nil},
		},
		{
			"Test Case 3, bad endpoint, 400 response",
			[]HealthResponse{
				{Healthy: false, Error: errors.New("")},
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
				{Healthy: true, Error: nil},
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
				{Healthy: false, Error: errors.New("")},
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
				{Healthy: false, Error: errors.New("")},
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
				{Healthy: true, Error: nil},
				{Healthy: true, Error: nil},
				{Healthy: true, Error: nil},
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
				if !g.Healthy && g.Error == nil {
					t.Error("Unhealthy endpoint returned nil error")
					continue
				} else if g.Healthy && g.Error != nil {
					t.Error("Healthy endpoint returned non-nil error")
					continue
				}

				matched := false
				for i, w := range tc.want {
					if mask[i] {
						continue
					}
					if g.Healthy == w.Healthy {
						mask[i] = true
						matched = true
						break
					}
				}

				if !matched {
					t.Errorf("Got %+v, want %+v", got, tc.want)
				}
			}
		})
	}
}

func TestSyncStatus(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		want     []BeaconSyncingStatus
		handlers []handler
	}{
		{
			"Test Case 1, no endpoints, empty",
			nil,
			nil,
		},
		{
			"Test Case 2, empty endpoints, request failed",
			[]BeaconSyncingStatus{{Error: errors.New("")}},
			[]handler{nil},
		},
		{
			"Test Case 3, bad endpoint, 400 response",
			[]BeaconSyncingStatus{{Error: errors.New("")}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusBadRequest)
				},
			},
		},
		{
			"Test Case 4, good endpoint, empty response body",
			[]BeaconSyncingStatus{{Error: errors.New("")}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(""))
				},
			},
		},
		{
			"Test Case 4, good endpoint, bad response body",
			[]BeaconSyncingStatus{{Error: errors.New("")}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("312312"))
				},
			},
		},
		{
			"Test Case 5, good endpoint, bad response body, bad json",
			[]BeaconSyncingStatus{{Error: errors.New("")}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("{"))
				},
			},
		},
		{
			"Test Case 6, good endpoint, not synced",
			[]BeaconSyncingStatus{{IsSyncing: true}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(`{"data":{
						"head_slot": "1",
						"sync_distance": "1",
						"is_syncing": true
					}}`))
				},
			},
		},
		{
			"Test Case 7, good endpoint, synced",
			[]BeaconSyncingStatus{{IsSyncing: false}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(`{"data":{
						"head_slot": "1",
						"sync_distance": "1",
						"is_syncing": false
					}}`))
				},
			},
		},
		{
			"Test Case 8, good endpoints, mixed sync status",
			[]BeaconSyncingStatus{{IsSyncing: false}, {IsSyncing: true}, {IsSyncing: false}},
			[]handler{
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(`{"data":{
						"head_slot": "1",
						"sync_distance": "1",
						"is_syncing": false
					}}`))
				},
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(`{"data":{
						"head_slot": "1",
						"sync_distance": "1",
						"is_syncing": true
					}}`))
				},
				func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(`{"data":{
						"head_slot": "1",
						"sync_distance": "1",
						"is_syncing": false
					}}`))
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

			got := client.SyncStatus(endpoints)

			mask := make(map[int]bool)
			for _, g := range got {
				matched := false
				for i, w := range tc.want {
					if mask[i] {
						continue
					}
					if g.IsSyncing == w.IsSyncing && (g.Error != nil && w.Error != nil || g.Error == nil && w.Error == nil) {
						mask[i] = true
						matched = true
						break
					}
				}

				if !matched {
					t.Errorf("Got %+v, want %+v", got, tc.want)
				}
			}
		})
	}
}
