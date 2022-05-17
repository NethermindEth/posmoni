//go:build integration

package networking

import (
	"os"
	"testing"
	"time"
)

func TestETH1SyncStatusIntegration(t *testing.T) {
	endpoint, exists := os.LookupEnv("PM_EC_ENDPOINT")
	if !exists {
		t.Fatal("PM_EC_ENDPOINT not set")
	}

	client := ExecutionClient{
		Endpoint:      endpoint,
		RetryDuration: time.Millisecond * 100,
	}

	got := client.SyncStatus()

	if got.Error != nil {
		t.Errorf("Got unexpected error calling SyncStatus(). Error: %v", got.Error)
	}
}
