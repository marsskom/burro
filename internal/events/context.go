package events

import "net/http"

type Context struct {
	Request  *http.Request
	Response *http.Response

	LastError error

	Metadata map[string]any

	IsHandled bool
}
