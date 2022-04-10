//go:build integration

package eth2

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	t.Parallel()

	sub := sseSubscriber{}
	ch := make(chan checkpoint)
	defer close(ch)

	endpoint, exists := os.LookupEnv("SSE_ENDPOINT")
	if !exists {
		t.Fatal("SSE_ENDPOINT not set")
	}

	go sub.listen(endpoint+finalizedCkptTopic, ch)

	for event := range ch {
		t.Logf("Checkpoint received: %+v", event)
		assert.NotEqual(t, event, checkpoint{}, "checkpoint object should not be empty")

		assert.IsType(t, event.Block, "", "block should be a string")
		assert.IsType(t, event.State, "", "state should be a string")
		assert.IsType(t, event.Epoch, "", "epoch should be a string")

		assert.True(t, strings.HasPrefix(event.Block, "0x"), "block should start with 0x")
		assert.True(t, strings.HasPrefix(event.State, "0x"), "state should start with 0x")

		break
	}
}
