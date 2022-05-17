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

