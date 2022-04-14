//go:build integration

package networking

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NethermindEth/posgonitor/internal/utils"
	"github.com/stretchr/testify/assert"
)

type callArgs struct {
	stateID       string
	validatorIdxs []string
}

func TestValidatorBalances(t *testing.T) {
	t.Parallel()

	endpoint, exists := os.LookupEnv("BC_ENDPOINT")
	if !exists {
		t.Fatal("BC_ENDPOINT not set")
	}

	tcs := []struct {
		name          string
		url           string
		args          callArgs
		checkResponse bool
		isError       bool
	}{
		{
			"Test Case 1, empty endpoint",
			"",
			callArgs{
				"head",
				[]string{},
			},
			true,
			true,
		},
		{
			"Test Case 2, bad endpoint",
			"http://localhost:8080",
			callArgs{
				"head",
				[]string{},
			},
			false,
			true,
		},
		{
			"Test Case 3, good endpoint, head, empty validatorIdxs",
			endpoint,
			callArgs{
				"head",
				[]string{},
			},
			true,
			false,
		},
		{
			"Test Case 4, good endpoint, slot 100",
			endpoint,
			callArgs{
				"100",
				[]string{"1", "2", "3"},
			},
			true,
			false,
		},
		{
			"Test Case 5, good endpoint, head",
			endpoint,
			callArgs{
				"100",
				[]string{"1", "2", "3"},
			},
			true,
			false,
		},
		{
			"Test Case 5, good endpoint, head, bad validators",
			endpoint,
			callArgs{
				"100",
				[]string{"1ad", "ttt", "0xwwww", "0x"},
			},
			false,
			false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			client := BeaconClient{
				Endpoint:      tc.url,
				RetryDuration: time.Second,
			}

			response, err := client.ValidatorBalances(tc.args.stateID, tc.args.validatorIdxs)
			descr := fmt.Sprintf("ValidatorBalances(%s, %s) with endpoint %s", tc.args.stateID, tc.args.validatorIdxs, tc.url)
			utils.CheckErr(t, descr, tc.isError, err)

			if tc.checkResponse {
				assert.Len(t, response, len(tc.args.validatorIdxs), descr+" returned wrong number of validators")
				for _, vb := range response {
					assert.NotEmpty(t, vb, descr+" returned empty validator")
				}
			}
		})
	}
}
