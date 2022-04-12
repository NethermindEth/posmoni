package networking

import (
	"encoding/json"
)

func parseEventData(data []byte) (Checkpoint, error) {
	var c Checkpoint
	err := json.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}
