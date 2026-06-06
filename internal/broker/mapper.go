package broker

import (
	"fmt"
	"time"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
)

func ToHTTPBrokerEvent(t EventType, requestContext *model.RequestContext) (Event, error) {
	if requestContext.ID == "" {
		return Event{}, fmt.Errorf("request context not initalized")
	}

	mtdata := ""
	if requestContext.Metadata != nil {
		m, err := persistence.MapToText(requestContext.Metadata)
		if err != nil {
			return Event{}, fmt.Errorf("cannot convert metadata to string: %w", err)
		}
		mtdata = m
	}

	respData := struct {
		Status        string
		StatusCode    int
		Proto         string
		Headers       string
		ContentLength int
		Body          []byte
	}{
		Status:        "",
		StatusCode:    0,
		Proto:         "",
		Headers:       "",
		ContentLength: 0,
		Body:          make([]byte, 0),
	}
	if requestContext.ResponseSnapshot != nil {
		respData.Status = requestContext.ResponseSnapshot.Status
		respData.StatusCode = requestContext.ResponseSnapshot.StatusCode
		respData.Proto = requestContext.ResponseSnapshot.Proto
		respData.ContentLength = requestContext.ResponseSnapshot.ContentLength
		respData.Body = requestContext.ResponseSnapshot.Body

		respHeaders, err := persistence.MapToText(requestContext.ResponseSnapshot.Headers)
		if err != nil {
			return Event{}, fmt.Errorf("cannot convert response headers map into string: %w", err)
		}

		respData.Headers = respHeaders
	}

	if requestContext.RequestSnapshot == nil {
		// TODO: `OnConnect` works before request snapshot is ready.
		return Event{
			TransportType: TransportHTTP,
			Type:          t,
			ID:            requestContext.ID,
			SessionID:     requestContext.Session.ID,
			Timestamp:     time.Now().UnixMilli(),

			HTTP: &HTTPEvent{
				Proto:         requestContext.Request.Proto,
				Scheme:        requestContext.Request.URL.Scheme,
				Host:          requestContext.Request.Host,
				Method:        requestContext.Request.Method,
				URL:           requestContext.Request.URL.String(),
				Path:          requestContext.Request.URL.Path,
				ContentLength: 0,
				RemoteAddr:    requestContext.Request.RemoteAddr,
				RequestBody:   make([]byte, 0),

				StartTime:  requestContext.StartTime.UnixMilli(),
				State:      int(requestContext.State.Load()),
				IsFinished: requestContext.IsFinished,

				QueryParams: "",
				Headers:     "",
				Cookies:     "",

				ResponseStatus:        "",
				ResponseStatusCode:    0,
				ResponseProto:         "",
				ResponseHeaders:       "",
				ResponseContentLength: 0,
				ResponseBody:          make([]byte, 0),

				CreatedAt: requestContext.CreatedAt.UnixMilli(),
				UpdatedAt: requestContext.UpdatedAt.UnixMilli(),

				Metadata: mtdata,
			},
			WS: &WSEvent{},
		}, nil
	}

	headers, err := persistence.MapToText(requestContext.RequestSnapshot.Headers)
	if err != nil {
		return Event{}, fmt.Errorf("cannot convert request headers map into string: %w", err)
	}

	queryParams, err := persistence.MapToText(requestContext.RequestSnapshot.QueryParams)
	if err != nil {
		return Event{}, fmt.Errorf("cannot convert query params map into string: %w", err)
	}

	cookiesAsText, err := persistence.MapToText(requestContext.RequestSnapshot.Cookies)
	if err != nil {
		return Event{}, fmt.Errorf("cannot convert cookies into string: %w", err)
	}

	return Event{
		TransportType: TransportHTTP,
		Type:          t,
		ID:            requestContext.ID,
		SessionID:     requestContext.Session.ID,
		Timestamp:     time.Now().UnixMilli(),

		HTTP: &HTTPEvent{
			Proto:         requestContext.RequestSnapshot.Proto,
			Scheme:        requestContext.RequestSnapshot.Scheme,
			Host:          requestContext.RequestSnapshot.Host,
			Method:        requestContext.RequestSnapshot.Method,
			URL:           requestContext.RequestSnapshot.URL,
			Path:          requestContext.RequestSnapshot.Path,
			ContentLength: requestContext.RequestSnapshot.ContentLength,
			RemoteAddr:    requestContext.RequestSnapshot.RemoteAddr,
			RequestBody:   requestContext.RequestSnapshot.Body,

			StartTime:  requestContext.StartTime.UnixMilli(),
			State:      int(requestContext.State.Load()),
			IsFinished: requestContext.IsFinished,

			QueryParams: queryParams,
			Headers:     headers,
			Cookies:     cookiesAsText,

			ResponseStatus:        respData.Status,
			ResponseStatusCode:    respData.StatusCode,
			ResponseProto:         respData.Proto,
			ResponseHeaders:       respData.Headers,
			ResponseContentLength: respData.ContentLength,
			ResponseBody:          respData.Body,

			CreatedAt: requestContext.CreatedAt.UnixMilli(),
			UpdatedAt: requestContext.UpdatedAt.UnixMilli(),

			Metadata: mtdata,
		},
		WS: &WSEvent{},
	}, nil
}

func ToWSBrokerEvent(t EventType, requestContext *model.RequestContext, msg *model.WSMessage) (Event, error) {
	var wsEvent *WSEvent
	if msg != nil {
		wsEvent = &WSEvent{
			Direction: string(msg.Direction),
			OpCode:    int(msg.OpCode),
			Data:      msg.Data,
			Text:      msg.Text,
			Timestamp: msg.Timestamp,
		}
	}

	return Event{
		TransportType: TransportWS,
		Type:          t,
		ID:            requestContext.ID,
		SessionID:     requestContext.Session.ID,
		Timestamp:     time.Now().UnixMilli(),

		HTTP: &HTTPEvent{},
		WS:   wsEvent,
	}, nil
}
