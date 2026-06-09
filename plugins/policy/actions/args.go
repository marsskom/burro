package actions

import (
	"bytes"
	"encoding/json"
)

func decode[T any](in any) (T, error) {
	var out T
	b, err := json.Marshal(in)
	if err != nil {
		return out, err
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	err = dec.Decode(&out)

	return out, err
}
