package actions

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func redactJSONField(data []byte, field string) []byte {
	var obj map[string]any

	if err := json.Unmarshal(data, &obj); err != nil {
		// Ignores non-JSON body.
		return data
	}

	delete(obj, field)

	out, err := json.Marshal(obj)
	if err != nil {
		// Ignores error as well.
		return data
	}

	return out
}

func readRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	// Has fun with a huge bodies.
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// Restores request body.
	req.Body = io.NopCloser(bytes.NewBuffer(b))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(b)), nil
	}

	return b, nil
}

func replaceRequestBody(req *http.Request, body []byte) {
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	req.ContentLength = int64(len(body))
}
