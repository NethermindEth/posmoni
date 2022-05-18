package networking

import (
	"encoding/json"
)

/*
unmarshalData :
Unmarshal json data into a given struct.

params :-
a. data []byte
JSON data to unmarshal
b. object J
Struct to unmarshal data into

returns :-
a. J
Unmarshalled struct
b. error
Error if any
*/
func unmarshalData[J any](data []byte, object J) (J, error) {
	err := json.Unmarshal(data, &object)
	if err != nil {
		return object, err
	}
	return object, nil
}
