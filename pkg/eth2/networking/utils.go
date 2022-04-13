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

func unmarshalData[J any](data []byte, object J) (J, error) {
	err := json.Unmarshal(data, &object)
	if err != nil {
		return object, err
	}
	return object, nil
}
