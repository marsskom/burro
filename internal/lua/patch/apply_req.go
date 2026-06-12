package patch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gitlab.com/marsskom/burro/internal/model"
)

func ApplyRequestPatch(r *http.Request, p *model.RequestPatch) error {
	if p == nil {
		return nil
	}

	if p.Host != nil {
		r.Host = *p.Host
		r.URL.Host = *p.Host
	}

	if p.Scheme != nil {
		r.URL.Scheme = *p.Scheme
	}

	if p.Method != nil {
		r.Method = *p.Method
	}

	if p.Path != nil {
		r.URL.Path = *p.Path
	}

	if p.URL != nil {
		u, err := url.Parse(*p.URL)
		if err != nil {
			return fmt.Errorf("req patch cannot parse url: %w", err)
		}

		r.URL = u
		r.Host = u.Host
	}

	if p.Body != nil {
		r.Body = io.NopCloser(bytes.NewReader(*p.Body))
		r.ContentLength = int64(len(*p.Body))
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(*p.Body)), nil
		}
	}

	for k, vals := range p.Headers {
		if vals == nil {
			r.Header.Del(k)
		} else {
			r.Header[k] = vals
		}
	}

	if len(p.Cookies) == 0 {
		return nil
	}

	cookies := r.Cookies()
	r.Header.Del("Cookie")

	handled := map[string]struct{}{}
	for _, c := range p.Cookies {
		handled[c.Name] = struct{}{}

		if c.Delete {
			continue
		}

		r.AddCookie(&http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
			MaxAge:   c.MaxAge,
		})
	}
	for _, c := range cookies {
		if _, ok := handled[c.Name]; ok {
			continue
		}

		r.AddCookie(c)
	}

	return nil
}
