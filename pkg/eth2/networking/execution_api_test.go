package networking

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/NethermindEth/posmoni/internal/utils"
	"github.com/stretchr/testify/assert"
)

func validateReq(req *http.Request, method string) error {
	if req.Method != "POST" {
		return fmt.Errorf("Unexpected HTTP method, expected POST, got %s", req.Method)
	}

	if req.Header["Content-Type"][0] != "application/json" {
		return fmt.Errorf("Wrong Content-Type header, got %s", req.Header["Content-Type"][0])
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("Got error reading request body. Error: %v", err)
	}
	defer req.Body.Close()

	var ethReq eth1Request
	if err = json.Unmarshal(data, &ethReq); err != nil {
		return err
	}

	if ethReq.Method != method {
		return fmt.Errorf("Wrong requested method. Expected %s, got %s", method, ethReq.Method)
	}
	return nil
}

func TestCall(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		want    json.RawMessage
		method  string
		params  []any
		handler handler
		isError bool
	}{
		{
			"Test case 1, eth_syncing, good call",
			[]byte(`false`),
			"eth_syncing",
			nil,
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}

				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"jsonrpc":"2.0","result":false,"id":1}`))
			},
			false,
		},
		{
			"Test case 2, eth_syncing, bad call, bad json",
			nil,
			"eth_syncing",
			nil,
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("{"))
			},
			true,
		},
		{
			"Test case 3, eth_syncing, bad call, bad params",
			nil,
			"eth_syncing",
			[]any{map[ExecutionClient]ExecutionClient{{}: {}}},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"jsonrpc":"2.0","result":false,"id":1}`))
			},
			true,
		},
		{
			"Test case 4, eth_syncing, good call, response error",
			nil,
			"eth_syncing",
			nil,
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"Internal error"},"id":1}`))
			},
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			srv := setupServer(tc.handler)
			defer srv.Close()

			client := ExecutionClient{
				Endpoint:      srv.URL,
				RetryDuration: time.Millisecond * 100,
			}

			got, err := client.Call("eth_syncing", tc.params...)

			descr := "Call(\"eth_syncing\")"
			if err = utils.CheckErr(descr, tc.isError, err); err != nil {
				t.Error(err)
			}

			assert.Equalf(t, tc.want, got, "%s failed", descr)
		})
	}
}

func TestETH1SyncStatus(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		want    ExecutionSyncingStatus
		handler handler
	}{
		{
			"Test case 1, nil endpoint, request failed",
			ExecutionSyncingStatus{Error: errors.New("")},
			nil,
		},
		{
			"Test case 2, bad endpoint, 400 response",
			ExecutionSyncingStatus{Error: errors.New("")},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusBadRequest)
			},
		},
		{
			"Test case 3, good endpoint, empty response body",
			ExecutionSyncingStatus{Error: errors.New("")},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(""))
			},
		},
		{
			"Test case 4, good endpoint, bad response body",
			ExecutionSyncingStatus{Error: errors.New("")},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("312312"))
			},
		},
		{
			"Test case 5, good endpoint, bad response body, bad json",
			ExecutionSyncingStatus{Error: errors.New("")},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("{"))
			},
		},
		{
			"Test case 6, good endpoint, not synced",
			ExecutionSyncingStatus{IsSyncing: true},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{
						"id":1,
						"jsonrpc": "2.0",
						"result": {
							"startingBlock": "0x384",
							"currentBlock": "0x386",
							"highestBlock": "0x454"
						}
					}`))
			},
		},
		{
			"Test case 7, good endpoint, synced",
			ExecutionSyncingStatus{IsSyncing: false},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}

				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{
						"id":1,
						"jsonrpc": "2.0",
						"result": false
					}`))
			},
		},
		{
			"Test case 8, good endpoint, good response body, incorrect json",
			ExecutionSyncingStatus{Error: errors.New("")},
			func(rw http.ResponseWriter, req *http.Request) {
				if err := validateReq(req, "eth_syncing"); err != nil {
					t.Fatalf("Request validation failed. Error: %v", err)
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"result": 666}`))
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			srv := setupServer(tc.handler)
			defer srv.Close()

			client := ExecutionClient{
				Endpoint:      srv.URL,
				RetryDuration: time.Millisecond * 100,
			}

			got := client.SyncStatus()

			if got.IsSyncing != tc.want.IsSyncing && !(got.Error != nil && tc.want.Error != nil || got.Error == nil && tc.want.Error == nil) {
				t.Errorf("Got %+v, want %+v", got, tc.want)
			}
		})
	}
}
