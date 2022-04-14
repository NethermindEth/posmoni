package networking

import (
	"encoding/json"
)

func unmarshalData[J any](data []byte, object J) (J, error) {
	err := json.Unmarshal(data, &object)
	if err != nil {
		return object, err
	}
	return object, nil
}
