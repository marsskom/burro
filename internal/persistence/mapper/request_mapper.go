package mapper

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
)

// TODO: make some DTO because managing this is going to be painful
func FromStoredRequest(storedRequest database.Request) (*model.RequestContext, error) {
	mtdata := map[string]any{}
	if storedRequest.Metadata.Valid {
		m, err := persistence.TextToMap(storedRequest.Metadata.String)
		if err != nil {
			return nil, fmt.Errorf("cannot convert metadata to map: %w", err)
		}
		mtdata = m
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(storedRequest.RequestRaw)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse raw request: %w", err)
	}

	var response *http.Response
	if len(storedRequest.RequestRaw) > 0 {
		res, err := http.ReadResponse(
			bufio.NewReader(bytes.NewReader(storedRequest.ResponseRaw)),
			req,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot parse raw response: %w", err)
		}
		response = res
	}

	requestContext := &model.RequestContext{
		ID:           storedRequest.ID,
		StartTime:    time.UnixMilli(storedRequest.StartTime),
		CreatedAt:    time.UnixMilli(storedRequest.CreatedAt.Int64),
		UpdatedAt:    time.UnixMilli(storedRequest.UpdatedAt.Int64),
		Request:      req,
		Response:     response,
		RequestBody:  storedRequest.RequestBody,
		ResponseBody: storedRequest.ResponseBody,
		Metadata:     mtdata,
		IsFinished:   persistence.IntToBool(int(storedRequest.IsFinished.Int64)),
	}

	requestContext.State.Store(int32(storedRequest.State.Int64))

	return requestContext, nil
}

func ToStoredRequest(requestContext *model.RequestContext) (database.Request, error) {
	mtdata := sql.NullString{
		Valid:  false,
		String: "",
	}
	if requestContext.Metadata != nil {
		m, err := persistence.MapToText(requestContext.Metadata)
		if err != nil {
			return database.Request{}, fmt.Errorf("cannot convert metadata to string: %w", err)
		}
		mtdata = sql.NullString{
			Valid:  true,
			String: m,
		}
	}

	requestRaw, err := httputil.DumpRequest(requestContext.Request, false)
	if err != nil {
		return database.Request{}, fmt.Errorf("cannot dump request: %w", err)
	}

	var responseRaw []byte
	if requestContext.Response != nil {
		resRaw, err := httputil.DumpResponse(requestContext.Response, false)
		if err != nil {
			return database.Request{}, fmt.Errorf("cannot dump response: %w", err)
		}
		responseRaw = resRaw
	}

	isFinished := sql.NullInt64{
		Valid: true,
		Int64: 0,
	}
	if requestContext.IsFinished {
		isFinished = sql.NullInt64{
			Valid: true,
			Int64: 1,
		}
	}

	return database.Request{
		ID:           requestContext.ID,
		SessionID:    requestContext.Session.ID,
		Host:         requestContext.Request.Host,
		Url:          requestContext.Request.URL.String(),
		Method:       requestContext.Request.Method,
		RequestRaw:   requestRaw,
		RequestBody:  requestContext.RequestBody,
		ResponseRaw:  responseRaw,
		ResponseBody: requestContext.ResponseBody,
		StartTime:    requestContext.StartTime.UnixMilli(),
		State: sql.NullInt64{
			Valid: true,
			Int64: int64(requestContext.State.Load()),
		},
		IsFinished: isFinished,
		Metadata:   mtdata,
		CreatedAt: sql.NullInt64{
			Valid: true,
			Int64: requestContext.CreatedAt.UnixMilli(),
		},
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: requestContext.UpdatedAt.UnixMilli(),
		},
	}, nil
}
