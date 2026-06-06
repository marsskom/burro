package mapper

import (
	"database/sql"
	"errors"
	"fmt"

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

	respData := struct {
		Status        sql.NullString
		StatusCode    sql.NullInt64
		Proto         sql.NullString
		Headers       sql.NullString
		ContentLength sql.NullInt64
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
		ID:        requestContext.ID,
		SessionID: requestContext.Session.ID,
		StartTime: requestContext.StartTime.UnixMilli(),
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

		ReqProto:         requestContext.RequestSnapshot.Proto,
		ReqHost:          requestContext.RequestSnapshot.Host,
		ReqScheme:        requestContext.RequestSnapshot.Scheme,
		ReqUrl:           requestContext.RequestSnapshot.URL,
		ReqPath:          requestContext.RequestSnapshot.Path,
		ReqQueryParams:   queryParams,
		ReqMethod:        requestContext.RequestSnapshot.Method,
		ReqHeaders:       headers,
		ReqCookies:       cookiesAsText,
		ReqContentLength: int64(requestContext.RequestSnapshot.ContentLength),
		ReqRemoteAddr:    requestContext.RequestSnapshot.RemoteAddr,
		ReqBody:          requestContext.RequestSnapshot.Body,

		ResStatus:        respData.Status,
		ResStatusCode:    respData.StatusCode,
		ResProto:         respData.Proto,
		ResHeaders:       respData.Headers,
		ResContentLength: respData.ContentLength,
		ResBody:          respData.Body,
	}, nil
}
