package patch

import (
	"bytes"
	"io"
	"net/http"

	"gitlab.com/marsskom/burro/internal/model"
)

func ApplyResponsePatch(resp *http.Response, p *model.ResponsePatch) error {
	if p == nil {
		return nil
	}

	if p.StatusCode != nil {
		resp.StatusCode = *p.StatusCode
		resp.Status = http.StatusText(*p.StatusCode)
	}

	if p.Body != nil {
		resp.Body = io.NopCloser(bytes.NewReader(*p.Body))
		resp.ContentLength = int64(len(*p.Body))
	}

	for k, vals := range p.Headers {
		if vals == nil {
			resp.Header.Del(k)
		} else {
			resp.Header[k] = vals
		}
	}

	return nil
}
