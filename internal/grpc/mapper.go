package grpc

import (
	"gitlab.com/marsskom/burro/internal/broker"
	pt "gitlab.com/marsskom/burro/internal/proto"
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

	transportType := pt.TransportType_TRANSPORT_HTTP
	if e.TransportType == broker.TransportWS {
		transportType = pt.TransportType_TRANSPORT_WS
	}

	return &pt.Event{
		TransportType: transportType,
		Type:          getEventType(e.Type),

		Id:        e.ID,
		SessionId: e.SessionID,

		Timestamp: e.Timestamp,

		Http: httpEvent,
		Ws:   wsEvent,
	}
}

func getEventType(eType broker.EventType) pt.EventType {
	switch eType {
	case broker.EventConnect:
		return pt.EventType_EVENT_CONNECT
	case broker.EventRequest:
		return pt.EventType_EVENT_REQUEST
	case broker.EventResponse:
		return pt.EventType_EVENT_RESPONSE
	case broker.EventError:
		return pt.EventType_EVENT_ERROR
	case broker.EventClose:
		return pt.EventType_EVENT_CLOSE

	case broker.EventWSConnect:
		return pt.EventType_EVENT_WS_CONNECT
	case broker.EventWSMessage:
		return pt.EventType_EVENT_WS_MESSAGE
	case broker.EventWSClose:
		return pt.EventType_EVENT_WS_CLOSE

	default:
		return pt.EventType_EVENT_CONNECT
	}
}
