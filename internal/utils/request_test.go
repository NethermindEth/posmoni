package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			query := req.URL.Query().Get("test")
			if query == "ERROR" {
				rw.WriteHeader(http.StatusBadRequest)
				return
			} else if query == "OK" {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("OK"))
			}
		}))
	defer server.Close()

	tcs := []struct {
		name          string
		url           string
		want          string
		retryDuration time.Duration
		isError       bool
	}{
		{
			"Good request",
			server.URL + "/?test=OK",
			"OK",
			time.Second,
			false,
		},
		{
			"Bad request",
			server.URL + "/?test=ERROR",
			"",
			time.Second,
			false,
		},
		{
			"No response",
			"http://127.0.0.1" + "/",
			"",
			time.Second,
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := GetRequest(tc.url, tc.retryDuration)
			descr := fmt.Sprintf("DoRequest(%s)", tc.url)
			CheckErr(t, descr, tc.isError, err)

			if resp != nil {
				defer resp.Body.Close()
				contents, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Reading response body failed. Error: %v", err)
				}

				if tc.want != string(contents) {
					t.Errorf("DoRequest(%s) response body is %s (got) != %s (want)", tc.url, string(contents), tc.want)
				}
			}
		})
	}
}
