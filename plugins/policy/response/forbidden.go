package response

import (
	"io"
	"net/http"
	"strings"
)

func Forbidden() *http.Response {
	body := "403 Forbidden"

	return &http.Response{
		Status:     "403 Forbidden",
		StatusCode: http.StatusForbidden,
		Header: http.Header{
			"Content-Type": []string{"text/plain; charset=utf-8"},
		},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       nil,
	}
}
