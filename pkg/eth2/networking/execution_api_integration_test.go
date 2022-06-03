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
		RetryDuration: time.Millisecond * 100,
	}

	got := client.SyncStatus([]string{endpoint})

	if len(got) != 1 {
		t.Errorf("Wrong len(got) value, expected %d, got %d", 1, len(got))
	}
	if got[0].Error != nil {
		t.Errorf("Got unexpected error calling SyncStatus(). Error: %v", got[0].Error)
	}
}
