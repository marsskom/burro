package events

import "net/http"

type Context struct {
	Request  *http.Request
	Response *http.Response

	Metadata map[string]any

	IsHandled bool
}
