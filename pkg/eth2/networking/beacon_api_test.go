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
			"Empty endpoint",
			"",
			callArgs{
				"head",
				[]string{},
			},
			false,
			true,
		},
		{
			"Bad endpoint",
			"http://localhost:8080",
			callArgs{
				"head",
				[]string{},
			},
			false,
			true,
		},
		{
			"Good endpoint, head, empty validatorIdxs",
			endpoint,
			callArgs{
				"head",
				[]string{},
			},
			true,
			false,
		},
		{
			"Good endpoint, slot 100",
			endpoint,
			callArgs{
				"100",
				[]string{"1", "2", "3"},
			},
			true,
			false,
		},
		{
			"Good endpoint, head",
			endpoint,
			callArgs{
				"100",
				[]string{"1", "2", "3"},
			},
			true,
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
