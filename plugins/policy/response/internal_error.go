package response

import (
	"io"
	"net/http"
	"strings"
)

func InternalError(err error) *http.Response {
	body := err.Error()

	return &http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: http.StatusInternalServerError,
		Header: http.Header{
			"Content-Type": []string{"text/plain; charset=utf-8"},
		},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       nil,
	}
}
