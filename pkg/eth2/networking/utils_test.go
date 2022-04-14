package networking

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/posgonitor/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalData(t *testing.T) {
	tcs := []struct {
		name    string
		data    []byte
		object  any
		want    any
		isError bool
	}{
		{
			"Case 1 - Valid data - Checkpoint",
			[]byte(`{"block":"0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf","state":"0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9","epoch":"2"}`),
			Checkpoint{},
			Checkpoint{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
			false,
		},
		{
			"Case 2 - Invalid data - Checkpoint",
			[]byte(`{block:"0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf",state:"0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9"}`),
			Checkpoint{},
			Checkpoint{},
			true,
		},
		{
			"Case 3 - Empty data - Checkpoint",
			[]byte(``),
			Checkpoint{},
			Checkpoint{},
			true,
		},
		{
			"Case 4 - Nil data - Checkpoint",
			nil,
			Checkpoint{},
			Checkpoint{},
			true,
		},
		{
			"Case 5 - Invalid data - Checkpoint",
			[]byte(`{{{{{}6352dqwda}}`),
			Checkpoint{},
			Checkpoint{},
			true,
		},
		{
			"Case 6 - Valid data - ValidatorBalanceList",
			[]byte(`{"data":[{"index":"1","balance":"32000136946"}, {"index":"2","balance":"640006946"}]}`),
			ValidatorBalanceList{},
			ValidatorBalanceList{
				Data: []ValidatorBalance{
					{Index: "1", Balance: "32000136946"},
					{Index: "2", Balance: "640006946"},
				},
			},
			false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var gotC Checkpoint
			var gotVB ValidatorBalanceList
			var err error

			descr := fmt.Sprintf("unmarshalData(data) with formatted data %s", string(tc.data))

			switch tc.object.(type) {
			case Checkpoint:
				gotC, err = unmarshalData(tc.data, tc.object.(Checkpoint))
				assert.Equal(t, tc.want, gotC, descr)
			case ValidatorBalanceList:
				gotVB, err = unmarshalData(tc.data, tc.object.(ValidatorBalanceList))
				assert.Equal(t, tc.want, gotVB, descr)
			}

			if err = utils.CheckErr(descr, tc.isError, err); err != nil {
				t.Fatal(err)
			}
		})
	}
}
