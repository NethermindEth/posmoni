package eth2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEventData(t *testing.T) {
	tcs := []struct {
		name    string
		data    []byte
		want    checkpoint
		isError bool
	}{
		{
			"Case 1 - Valid data",
			[]byte(`{"block":"0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf","state":"0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9","epoch":"2"}`),
			checkpoint{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
			false,
		},
		{
			"Case 2 - Invalid data",
			[]byte(`{block:"0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf",state:"0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9"}`),
			checkpoint{},
			true,
		},
		{
			"Case 3 - Empty data",
			[]byte(``),
			checkpoint{},
			true,
		},
		{
			"Case 4 - Nil data",
			nil,
			checkpoint{},
			true,
		},
		{
			"Case 5 - Invalid data",
			[]byte(`{{{{{}6352dqwda}}`),
			checkpoint{},
			true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseEventData(tc.data)

			descr := fmt.Sprintf("parseEventData(data) with formatted data %s", string(tc.data))
			if ok := checkErr(t, descr, tc.isError, err); !ok {
				t.FailNow()
			}

			assert.Equal(t, tc.want, got, descr)
		})
	}
}
