package model

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type RequestState int

const (
	StateReceived RequestState = iota
	StatePrepared
	StateForwarding
	StateResponding
	StateFinished
	StateCanceled
	StateFailed
)

var transitions = map[RequestState][]RequestState{
	StateReceived:   {StatePrepared, StateCanceled, StateFailed},
	StatePrepared:   {StateForwarding, StateCanceled, StateFailed},
	StateForwarding: {StateResponding, StateCanceled, StateFailed},
	StateResponding: {StateFinished},
}

type RequestContext struct {
	ID        string
	StartTime time.Time
	State     atomic.Int32

	CreatedAt time.Time
	UpdatedAt time.Time

	Session *Session

	Request  *http.Request
	Response *http.Response

	RequestSnapshot  *RequestSnapshot
	ResponseSnapshot *ResponseSnapshot

	Context context.Context
	Cancel  context.CancelFunc

	Metadata map[string]any

	IsFinished bool

	mu sync.RWMutex
}

func NewCtx(session *Session, r *http.Request) *RequestContext {
	base := r.Context()

	ctx, cancel := context.WithTimeout(base, 30*time.Second)

	return &RequestContext{
		ID:        uuid.NewString(),
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Session:   session,
		Request:   r,
		Context:   ctx,
		Cancel:    cancel,
	}
}

func NewCtxFromParent(parent *RequestContext, r *http.Request) *RequestContext {
	parent.mu.RLock()
	defer parent.mu.RUnlock()

	ctx, cancel := context.WithCancel(parent.Context)

	var mtdata map[string]any
	if parent.Metadata != nil {
		maps.Copy(mtdata, parent.Metadata)
	}

	return &RequestContext{
		ID:        uuid.NewString(),
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Session:   parent.Session,
		Request:   r,
		Context:   ctx,
		Cancel:    cancel,
		Metadata:  mtdata,
	}
}

func (c *RequestContext) GetState() RequestState {
	return RequestState(c.State.Load())
}

func (c *RequestContext) Transition(next RequestState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := RequestState(c.State.Load())

	if !slices.Contains(transitions[current], next) {
		return fmt.Errorf("invalid request state transition: %v -> %v", current, next)
	}

	c.State.Store(int32(next))
	c.UpdatedAt = time.Now()

	return nil
}

func (c *RequestContext) Finish(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Response = resp
	c.IsFinished = true
	c.UpdatedAt = time.Now()
}

type CookieSnapshot struct {
	Name        string
	Value       string
	Quoted      bool
	Path        string
	Domain      string
	Expires     time.Time
	MaxAge      int
	Secure      bool
	HTTPOnly    bool
	SameSite    int
	Partitioned bool
}

func ExtractCookies(r *http.Request) ([]*CookieSnapshot, error) {
	cookies := make([]*CookieSnapshot, 0, len(r.Cookies()))
	for _, c := range r.Cookies() {
		cookies = append(cookies, &CookieSnapshot{
			Name:        c.Name,
			Value:       c.Value,
			Quoted:      c.Quoted,
			Path:        c.Path,
			Domain:      c.Domain,
			Expires:     c.Expires,
			MaxAge:      c.MaxAge,
			Secure:      c.Secure,
			HTTPOnly:    c.HttpOnly,
			SameSite:    int(c.SameSite),
			Partitioned: c.Partitioned,
		})
	}

	return cookies, nil
}

type RequestSnapshot struct {
	Proto         string
	Host          string
	Method        string
	Scheme        string
	URL           string
	Path          string
	QueryParams   map[string][]string
	Headers       map[string][]string
	Cookies       []*CookieSnapshot
	ContentLength int
	RemoteAddr    string
	Body          []byte
}

func MakeRequestSnapshot(r *http.Request) (*RequestSnapshot, error) {
	queryParams := make(map[string][]string, len(r.URL.Query()))
	for k, v := range r.URL.Query() {
		if k == "" {
			continue
		}

		queryParams[k] = append([]string(nil), v...)
	}

	reqHeaders := r.Header.Clone()
	headers := make(map[string][]string, len(reqHeaders))
	for k, v := range reqHeaders {
		if k == "" {
			continue
		}

		headers[k] = append([]string(nil), v...)
	}

	cookies, err := ExtractCookies(r)
	if err != nil {
		return &RequestSnapshot{}, fmt.Errorf("error on cookie extraction from the request: %w", err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return &RequestSnapshot{}, fmt.Errorf("error on read request body for snapshot: %w", err)
	}

	r.Body.Close()
	// Restores request body for next request.
	r.Body = io.NopCloser(bytes.NewReader(body))

	snapshot := &RequestSnapshot{
		Proto:         r.Proto,
		Host:          r.Host,
		Method:        r.Method,
		Scheme:        r.URL.Scheme,
		URL:           BuildAbsoluteURL(r),
		Path:          r.URL.Path,
		QueryParams:   queryParams,
		Headers:       headers,
		Cookies:       cookies,
		ContentLength: len(body),
		RemoteAddr:    r.RemoteAddr,
		Body:          body,
	}

	slog.Debug("Request snapshot was created", "request", snapshot)

	return snapshot, nil
}

func BuildAbsoluteURL(r *http.Request) string {
	u := *r.URL

	if u.Scheme == "" {
		if r.TLS != nil {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
	}

	if u.Host == "" {
		u.Host = r.Host
	}

	return u.String()
}

type ResponseSnapshot struct {
	Status        string
	StatusCode    int
	Proto         string
	Headers       map[string][]string
	ContentLength int
	Body          []byte

	TimeDNS     time.Duration
	TimeConnect time.Duration
	TimeSSL     time.Duration
	TimeWait    time.Duration
}

func MakeResponseSnapshot(res *http.Response, t *Timings) (*ResponseSnapshot, error) {
	resHeaders := res.Header.Clone()
	headers := make(map[string][]string, len(resHeaders))
	for k, v := range resHeaders {
		headers[k] = append([]string(nil), v...)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &ResponseSnapshot{}, fmt.Errorf("error on read response body for snapshot: %w", err)
	}

	res.Body.Close()
	// Restores response body.
	res.Body = io.NopCloser(bytes.NewReader(body))

	snapshot := &ResponseSnapshot{
		Status:        res.Status,
		StatusCode:    res.StatusCode,
		Proto:         res.Proto,
		Headers:       resHeaders,
		ContentLength: len(body),
		Body:          body,

		TimeDNS:     t.DNSEnd.Sub(t.DNSStart),
		TimeConnect: t.ConnectEnd.Sub(t.ConnectStart),
		TimeSSL:     t.TLSEnd.Sub(t.TLSStart),
		TimeWait:    t.FirstByte.Sub(t.WroteRequest),
	}

	slog.Debug("Response snapshot was created", "response", snapshot)

	return snapshot, nil
}
