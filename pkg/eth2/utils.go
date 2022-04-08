package eth2

import "encoding/json"

func parseEventData(data []byte) (checkpoint, error) {
	var c checkpoint
	err := json.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}
