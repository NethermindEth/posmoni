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

	beaconClient := BeaconClient{
		Endpoint:      endpoint,
		RetryDuration: time.Second,
	}
	syncingStatus := beaconClient.SyncStatus([]string{endpoint})
	headSlot := syncingStatus[0].HeadSlot

	tcs := []struct {
		name          string
		url           string
		args          callArgs
		checkResponse bool
		isError       bool
	}{
		{
			"Test Case 1, empty endpoint, head slot",
			"",
			callArgs{
				"head",
				[]string{},
			},
			true,
			true,
		},
		{
			"Test Case 2, bad endpoint, head slot",
			"http://localhost:8080",
			callArgs{
				"head",
				[]string{},
			},
			false,
			true,
		},
		{
			"Test Case 3, good endpoint, head slot, empty validatorIdxs",
			endpoint,
			callArgs{
				"head",
				[]string{},
			},
			false,
			false,
		},
		{
			fmt.Sprintf("Test Case 4, good endpoint, slot %s", headSlot),
			endpoint,
			callArgs{
				headSlot,
				[]string{"1", "2", "3"},
			},
			true,
			false,
		},
		{
			fmt.Sprintf("Test Case 5, good endpoint, slot %s, bad validators", headSlot),
			endpoint,
			callArgs{
				headSlot,
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
				{Endpoint: "http://localhost:8080", Healthy: false, Error: errors.New("")},
			},
		},
		{
			"Test Case 3, good endpoint",
			bceps[0:1],
			[]HealthResponse{
				{Endpoint: bceps[0], Healthy: true, Error: nil},
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
					tc.want = append(tc.want, HealthResponse{Endpoint: tc.urls[i], Healthy: true, Error: nil})
				}
			}

			got := client.Health(tc.urls)
			descr := fmt.Sprintf("Health(%v)", tc.urls)

			require.Equal(t, len(tc.want), len(got), descr+" returned wrong number of endpoints")
			for i := 0; i < len(tc.want); i++ {
				if !got[i].Healthy && got[i].Error == nil {
					t.Error("Unhealthy endpoint returned nil error")
					continue
				} else if got[i].Healthy && got[i].Error != nil {
					t.Error("Healthy endpoint returned non-nil error")
					continue
				}

				if got[i].Endpoint != tc.want[i].Endpoint && got[i].Healthy != tc.want[i].Healthy && (got[i].Error == nil && tc.want[i].Error == nil || got[i].Error != nil && tc.want[i].Error != nil) {
					t.Error(descr + " returned bad response for " + tc.want[i].Endpoint + " endpoint")
				}
			}
		})
	}
}

func TestSyncStatusIntegration(t *testing.T) {
	t.Parallel()

	raw, exists := os.LookupEnv("PM_BC_ENDPOINTS")
	if !exists {
		t.Fatal("PM_BC_ENDPOINTS not set")
	}
	bceps := strings.Split(raw, ",")

	tcs := []struct {
		name string
		urls []string
		want []BeaconSyncingStatus
	}{
		{
			"Test Case 1, empty endpoints",
			[]string{},
			nil,
		},
		{
			"Test Case 2, bad endpoint",
			[]string{"http://localhost:8080"},
			[]BeaconSyncingStatus{{Error: errors.New("")}},
		},
		{
			"Test Case 3, good endpoint",
			bceps[0:1],
			[]BeaconSyncingStatus{{IsSyncing: false}},
		},
		{
			"Test Case 4, good endpoints",
			bceps,
			[]BeaconSyncingStatus{},
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
					tc.want = append(tc.want, BeaconSyncingStatus{IsSyncing: false})
				}
			}

			got := client.SyncStatus(tc.urls)
			descr := fmt.Sprintf("SyncStatus(%v)", tc.urls)

			require.Equal(t, len(tc.want), len(got), descr+" returned wrong number of endpoints")
			for i := 0; i < len(tc.want); i++ {
				if got[i].IsSyncing != tc.want[i].IsSyncing && (got[i].Error == nil && tc.want[i].Error == nil || got[i].Error != nil && tc.want[i].Error != nil) {
					t.Errorf(descr+" got bad response %+v", got[i])
				}
			}
		})
	}
}
