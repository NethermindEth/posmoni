package networking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testSubscriber struct {
	data map[string][]Checkpoint
}

func (s testSubscriber) Listen(url string, ch chan<- Checkpoint) {
	for _, data := range s.data[url] {
		ch <- data
		//sleep to simulate a delay
		time.Sleep(time.Millisecond * 50)
	}
}

func TestSubscribe(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		endpoints []string
		messages  map[string][]Checkpoint
		want      []Checkpoint
	}{
		{
			"Case 1 - 1 result",
			[]string{"Endpoint1"},
			map[string][]Checkpoint{
				"Endpoint1": {
					{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
				},
			},
			[]Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
			},
		},
		{
			"Case 1 - 2 result",
			[]string{"Endpoint1"},
			map[string][]Checkpoint{
				"Endpoint1": {
					{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
					{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "3"},
				},
			},
			[]Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "3"},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			done := make(chan struct{})
			sub := SubscribeOpts{
				Endpoints:  tc.endpoints,
				StreamURL:  "",
				Subscriber: testSubscriber{data: tc.messages},
			}
			ch := Subscribe(done, sub)

			got := make([]Checkpoint, 0)
			go func() {
				for c := range ch {
					got = append(got, c)
				}
			}()
			duration := len(tc.want) * 50
			time.Sleep(time.Millisecond * time.Duration(duration))
			done <- struct{}{}

			for i, want := range tc.want {
				assert.Equal(t, want, got[i])
			}
		})
	}
}
