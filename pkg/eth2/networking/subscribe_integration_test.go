//go:build integration

package networking

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	t.Parallel()

	sub := SSESubscriber{}
	ch := make(chan Checkpoint)
	defer close(ch)

	raw, exists := os.LookupEnv("PM_BC_ENDPOINTS")
	if !exists {
		t.Fatal("PM_BC_ENDPOINTS not set")
	}
	endpoint := strings.Split(raw, ",")[0]

	go sub.Listen(endpoint+FinalizedCkptTopic, ch)

	for event := range ch {
		t.Logf("Checkpoint received: %+v", event)
		assert.NotEqual(t, event, Checkpoint{}, "Checkpoint object should not be empty")

		assert.IsType(t, event.Block, "", "block should be a string")
		assert.IsType(t, event.State, "", "state should be a string")
		assert.IsType(t, event.Epoch, "", "epoch should be a string")

		assert.True(t, strings.HasPrefix(event.Block, "0x"), "block should start with 0x")
		assert.True(t, strings.HasPrefix(event.State, "0x"), "state should start with 0x")

		break
	}
}
