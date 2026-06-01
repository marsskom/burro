package grpc

import (
	"gitlab.com/marsskom/burro/internal/broker"
	pt "gitlab.com/marsskom/burro/internal/proto"
)

func brokerEventToProtoEvent(e broker.Event) *pt.Event {
	return &pt.Event{
		Type: string(e.Type),

		Id:        e.ID,
		SessionId: e.SessionID,

		Proto:  e.Proto,
		Scheme: e.Scheme,
		Host:   e.Host,
		Method: e.Method,
		Url:    e.URL,
		Path:   e.Path,

		ContentLength: int64(e.ContentLength),
		RemoteAddr:    e.RemoteAddr,

		StartTime:  e.StartTime,
		State:      int32(e.State),
		IsFinished: e.IsFinished,

		QueryParams: e.QueryParams,
		Headers:     e.Headers,
		Cookies:     e.Cookies,

		RequestBody: e.RequestBody,

		ResponseStatus:        e.ResponseStatus,
		ResponseStatusCode:    int32(e.ResponseStatusCode),
		ResponseProto:         e.ResponseProto,
		ResponseHeaders:       e.ResponseHeaders,
		ResponseContentLength: int64(e.ResponseContentLength),
		ResponseBody:          e.ResponseBody,

		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,

		Metadata: e.Metadata,
	}
}
