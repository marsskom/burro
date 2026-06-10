package grpc

import (
	"gitlab.com/marsskom/burro/internal/broker"
	pt "gitlab.com/marsskom/burro/internal/proto/burro/v1"
)

func brokerEventToProtoEvent(e broker.Event) *pt.Event {
	var httpEvent *pt.HTTPEvent
	if e.HTTP != nil {
		httpEvent = &pt.HTTPEvent{
			Proto:                 e.HTTP.Proto,
			Scheme:                e.HTTP.Scheme,
			Host:                  e.HTTP.Host,
			Method:                e.HTTP.Method,
			Url:                   e.HTTP.URL,
			Path:                  e.HTTP.Path,
			ContentLength:         int64(e.HTTP.ContentLength),
			RemoteAddr:            e.HTTP.RemoteAddr,
			StartTime:             e.HTTP.StartTime,
			State:                 int32(e.HTTP.State),
			IsFinished:            e.HTTP.IsFinished,
			QueryParams:           e.HTTP.QueryParams,
			Headers:               e.HTTP.Headers,
			Cookies:               e.HTTP.Cookies,
			RequestBody:           e.HTTP.RequestBody,
			ResponseStatus:        e.HTTP.ResponseStatus,
			ResponseStatusCode:    int32(e.HTTP.ResponseStatusCode),
			ResponseProto:         e.HTTP.ResponseProto,
			ResponseHeaders:       e.HTTP.ResponseHeaders,
			ResponseContentLength: int64(e.HTTP.ResponseContentLength),
			ResponseBody:          e.HTTP.ResponseBody,
			CreatedAt:             e.HTTP.CreatedAt,
			UpdatedAt:             e.HTTP.UpdatedAt,
			Metadata:              e.HTTP.Metadata,
		}
	}

	var wsEvent *pt.WSEvent
	if e.WS != nil {
		wsEvent = &pt.WSEvent{
			Direction: e.WS.Direction,
			Opcode:    int32(e.WS.OpCode),
			Data:      e.WS.Data,
			Text:      e.WS.Text,
			Timestamp: e.WS.Timestamp,
		}
	}

	transportType := pt.TransportType_TRANSPORT_TYPE_HTTP
	if e.TransportType == broker.TransportWS {
		transportType = pt.TransportType_TRANSPORT_TYPE_HTTP
	}

	return &pt.Event{
		TransportType: transportType,
		Type:          getProtoEventType(e.Type),

		Id:        e.ID,
		SessionId: e.SessionID,

		Timestamp: e.Timestamp,

		Http: httpEvent,
		Ws:   wsEvent,
	}
}

var eventTypeToProto = map[broker.EventType]pt.EventType{
	broker.EventConnect: pt.EventType_EVENT_TYPE_CONNECT,

	broker.EventBeforeRequestSend: pt.EventType_EVENT_TYPE_BEFORE_REQUEST_SEND,
	broker.EventAfterRequestSend:  pt.EventType_EVENT_TYPE_AFTER_REQUEST_SEND,

	broker.EventBeforeResponseSend: pt.EventType_EVENT_TYPE_BEFORE_RESPONSE_SEND,
	broker.EventAfterResponseSend:  pt.EventType_EVENT_TYPE_AFTER_RESPONSE_SEND,

	broker.EventError: pt.EventType_EVENT_TYPE_ERROR,
	broker.EventClose: pt.EventType_EVENT_TYPE_CLOSE,

	broker.EventWSConnect: pt.EventType_EVENT_TYPE_WS_CONNECT,
	broker.EventWSMessage: pt.EventType_EVENT_TYPE_WS_MESSAGE,
	broker.EventWSClose:   pt.EventType_EVENT_TYPE_WS_CLOSE,
}

var eventTypeFromProto = map[pt.EventType]broker.EventType{
	pt.EventType_EVENT_TYPE_CONNECT: broker.EventConnect,

	pt.EventType_EVENT_TYPE_BEFORE_REQUEST_SEND: broker.EventBeforeRequestSend,
	pt.EventType_EVENT_TYPE_AFTER_REQUEST_SEND:  broker.EventAfterRequestSend,

	pt.EventType_EVENT_TYPE_BEFORE_RESPONSE_SEND: broker.EventBeforeResponseSend,
	pt.EventType_EVENT_TYPE_AFTER_RESPONSE_SEND:  broker.EventAfterResponseSend,

	pt.EventType_EVENT_TYPE_ERROR: broker.EventError,
	pt.EventType_EVENT_TYPE_CLOSE: broker.EventClose,

	pt.EventType_EVENT_TYPE_WS_CONNECT: broker.EventWSConnect,
	pt.EventType_EVENT_TYPE_WS_MESSAGE: broker.EventWSMessage,
	pt.EventType_EVENT_TYPE_WS_CLOSE:   broker.EventWSClose,
}

func getProtoEventType(eType broker.EventType) pt.EventType {
	if v, ok := eventTypeToProto[eType]; ok {
		return v
	}

	return pt.EventType_EVENT_TYPE_CONNECT
}

func getBrokerEventType(eType pt.EventType) broker.EventType {
	if v, ok := eventTypeFromProto[eType]; ok {
		return v
	}

	return broker.EventConnect
}
