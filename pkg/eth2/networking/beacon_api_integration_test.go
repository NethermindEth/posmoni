//go:build integration

package networking

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/posmoni/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorBalances(t *testing.T) {
	t.Parallel()

	type callArgs struct {
		stateID       string
		validatorIdxs []string
	}

	raw, exists := os.LookupEnv("PM_BC_ENDPOINTS")
	if !exists {
		t.Fatal("PM_BC_ENDPOINTS not set")
	}
	endpoint := strings.Split(raw, ",")[0]

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
			true,
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
			if err = utils.CheckErr(descr, tc.isError, err); err != nil {
				t.Error(err)
			}

			if tc.checkResponse {
				assert.Len(t, response, len(tc.args.validatorIdxs), descr+" returned wrong number of validators")
				for _, vb := range response {
					assert.NotEmpty(t, vb, descr+" returned empty validator")
				}
			}
		})
	}
}

func TestHealthIntegration(t *testing.T) {
	t.Parallel()

	raw, exists := os.LookupEnv("PM_BC_ENDPOINTS")
	if !exists {
		t.Fatal("PM_BC_ENDPOINTS not set")
	}
	bceps := strings.Split(raw, ",")

	// Validator health checks are for local validators
	// raw, exists = os.LookupEnv("VL_ENDPOINTS")
	// if !exists {
	// 	t.Fatal("VL_ENDPOINTS not set")
	// }
	// vleps := strings.Split(raw, ",")
	vleps := []string{}

	tcs := []struct {
		name string
		urls []string
		want []HealthResponse
	}{
		{
			"Test Case 1, empty endpoints",
			[]string{},
			[]HealthResponse{},
		},
		{
			"Test Case 2, bad endpoint",
			[]string{"http://localhost:8080"},
			[]HealthResponse{
				{endpoint: "http://localhost:8080", healthy: false, err: errors.New("")},
			},
		},
		{
			"Test Case 3, good endpoint",
			bceps[0:1],
			[]HealthResponse{
				{endpoint: bceps[0], healthy: true, err: nil},
			},
		},
		{
			"Test Case 4, good endpoints",
			append(bceps, vleps...),
			[]HealthResponse{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			client := BeaconClient{
				RetryDuration: time.Second,
			}

			if len(tc.want) == 0 {
				// If want is empty then create a want list with good responses with the same length as the urls list
				for i := 0; i < len(tc.urls); i++ {
					tc.want = append(tc.want, HealthResponse{endpoint: tc.urls[i], healthy: true, err: nil})
				}
			}

			got := client.Health(tc.urls)
			descr := fmt.Sprintf("Health(%v)", tc.urls)

			require.Equal(t, len(tc.want), len(got), descr+" returned wrong number of endpoints")
			for i := 0; i < len(tc.want); i++ {
				if !got[i].healthy && got[i].err == nil {
					t.Error("Unhealthy endpoint returned nil error")
					continue
				} else if got[i].healthy && got[i].err != nil {
					t.Error("Healthy endpoint returned non-nil error")
					continue
				}

				if got[i].endpoint != tc.want[i].endpoint && got[i].healthy != tc.want[i].healthy && (got[i].err != nil && tc.want[i].err != nil) {
					t.Error(descr + " returned bad response for " + tc.want[i].endpoint + " endpoint")
				}
			}
		})
	}
}
