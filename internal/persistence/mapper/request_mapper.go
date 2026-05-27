package mapper

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http/httputil"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
)

func ToStoredRequest(requestContext *model.RequestContext) (database.Request, error) {
	if requestContext.RequestSnapshot == nil {
		return database.Request{}, errors.New("request snapshot cannot be nil")
	}

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

	respData := struct {
		Status        sql.NullString
		StatusCode    sql.NullInt64
		Proto         sql.NullString
		Headers       sql.NullString
		ContentLength sql.NullInt64
		Duration      sql.NullInt64
		Body          []byte
	}{
		Status: sql.NullString{
			Valid:  false,
			String: "",
		},
		StatusCode: sql.NullInt64{
			Valid: false,
			Int64: 0,
		},
		Proto: sql.NullString{
			Valid:  false,
			String: "",
		},
		Headers: sql.NullString{
			Valid:  false,
			String: "",
		},
		ContentLength: sql.NullInt64{
			Valid: false,
			Int64: 0,
		},
		Body: make([]byte, 0),
	}
	if requestContext.ResponseSnapshot != nil {
		respData.Status = sql.NullString{
			Valid:  true,
			String: requestContext.ResponseSnapshot.Status,
		}
		respData.StatusCode = sql.NullInt64{
			Valid: true,
			Int64: int64(requestContext.ResponseSnapshot.StatusCode),
		}
		respData.Proto = sql.NullString{
			Valid:  true,
			String: requestContext.ResponseSnapshot.Proto,
		}
		respData.ContentLength = sql.NullInt64{
			Valid: true,
			Int64: int64(requestContext.ResponseSnapshot.ContentLength),
		}
		respData.Body = requestContext.ResponseSnapshot.Body

		respHeaders, err := persistence.MapToText(requestContext.ResponseSnapshot.Headers)
		if err != nil {
			return database.Request{}, fmt.Errorf("cannot convert response headers map into string: %w", err)
		}

		respData.Headers = sql.NullString{
			Valid:  true,
			String: respHeaders,
		}
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

	headers, err := persistence.MapToText(requestContext.RequestSnapshot.Headers)
	if err != nil {
		return database.Request{}, fmt.Errorf("cannot convert request headers map into string: %w", err)
	}

	queryParams, err := persistence.MapToText(requestContext.RequestSnapshot.QueryParams)
	if err != nil {
		return database.Request{}, fmt.Errorf("cannot convert query params map into string: %w", err)
	}

	cookiesAsText, err := persistence.MapToText(requestContext.RequestSnapshot.Cookies)
	if err != nil {
		return database.Request{}, fmt.Errorf("cannot convert cookies into string: %w", err)
	}

	return database.Request{
		ID:            requestContext.ID,
		SessionID:     requestContext.Session.ID,
		Proto:         requestContext.RequestSnapshot.Proto,
		Host:          requestContext.RequestSnapshot.Host,
		Scheme:        requestContext.RequestSnapshot.Scheme,
		Url:           requestContext.RequestSnapshot.URL,
		Path:          requestContext.RequestSnapshot.Path,
		QueryParams:   queryParams,
		Method:        requestContext.RequestSnapshot.Method,
		Headers:       headers,
		Cookies:       cookiesAsText,
		ContentLength: int64(requestContext.RequestSnapshot.ContentLength),
		RemoteAddr:    requestContext.RequestSnapshot.RemoteAddr,
		RequestBody:   requestContext.RequestSnapshot.Body,

		RequestRaw: requestRaw,

		StartTime: requestContext.StartTime.UnixMilli(),
		State: sql.NullInt64{
			Valid: true,
			Int64: int64(requestContext.State.Load()),
		},
		IsFinished: isFinished,
		Metadata:   mtdata,

		RespStatus:        respData.Status,
		RespStatusCode:    respData.StatusCode,
		RespProto:         respData.Proto,
		RespHeaders:       respData.Headers,
		RespContentLength: respData.ContentLength,
		ResponseBody:      respData.Body,

		ResponseRaw: responseRaw,

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
