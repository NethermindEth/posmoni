package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
			if req.Method != "GET" {
				t.Errorf("Unexpected HTTP method, expected GET, got %s", req.Method)
			}

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
			descr := fmt.Sprintf("GetRequest(%s)", tc.url)
			if err = CheckErr(descr, tc.isError, err); err != nil {
				t.Error(err)
			}

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

func TestPostRequest(t *testing.T) {
	t.Parallel()

	type testObj struct {
		Data int `json:"data"`
	}

	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != "POST" {
				t.Errorf("Unexpected HTTP method, expected POST, got %s", req.Method)
			}

			if req.Header["Content-Type"][0] == "application/json" {
				data, err := ioutil.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("Got error reading request body. Error: %v", err)
				}
				defer req.Body.Close()

				tobj := new(testObj)
				err = json.Unmarshal(data, &tobj)
				if err != nil {
					t.Fatalf("Got error unmarshalling request body as json. Error: %v", err)
				}

				if tobj.Data != 666 {
					t.Errorf("Expected json result with data '666', got %+v", tobj)
				}
			}

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

	type postArgs struct {
		contentType   string
		body          io.Reader
		retry         bool
		retryDuration time.Duration
	}

	tcs := []struct {
		name    string
		url     string
		want    string
		args    postArgs
		isError bool
	}{
		{
			"Test case 1, good request, json body",
			server.URL + "/?test=OK",
			"OK",
			postArgs{
				contentType: "application/json",
				body: bytes.NewBufferString(`{
					"data": 666
				}`),
			},
			false,
		},
		{
			"Test case 2, good request, text body",
			server.URL + "/?test=OK",
			"OK",
			postArgs{
				contentType: "text/plain",
				body:        bytes.NewBufferString("OK"),
			},
			false,
		},
		{
			"Test case 3, bad request, retries",
			server.URL + "/?test=ERROR",
			"",
			postArgs{
				contentType:   "text/plain",
				body:          bytes.NewBufferString("ERROR"),
				retry:         true,
				retryDuration: time.Millisecond,
			},
			false,
		},
		{
			"Test case 4, bad request, no retries",
			server.URL + "/?test=ERROR",
			"",
			postArgs{
				contentType: "text/plain",
				body:        bytes.NewBufferString("ERROR"),
			},
			false,
		},
		{
			"Test case 5, no response, no retries",
			"http://127.0.0.1" + "/",
			"",
			postArgs{
				contentType: "text/plain",
				body:        bytes.NewBufferString("ERROR"),
			},
			true,
		},
		{
			"Test case 6, no response, retries",
			"http://127.0.0.1" + "/",
			"",
			postArgs{
				contentType:   "text/plain",
				body:          bytes.NewBufferString("ERROR"),
				retry:         true,
				retryDuration: time.Millisecond,
			},
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := PostRequest(tc.url, tc.args.contentType, tc.args.body, tc.args.retry, tc.args.retryDuration)
			descr := fmt.Sprintf("PostRequest(%s)", tc.url)
			if err = CheckErr(descr, tc.isError, err); err != nil {
				t.Error(err)
			}

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
