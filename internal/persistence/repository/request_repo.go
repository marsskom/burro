package repository

import (
	"context"
	"fmt"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence/mapper"
)

type RequestRepository struct {
	q *database.Queries
}

func NewRequestRepo(q *database.Queries) *RequestRepository {
	return &RequestRepository{
		q: q,
	}
}

func (r *RequestRepository) SaveRequest(ctx context.Context, requestContext *model.RequestContext) error {
	storedRequest, err := mapper.ToStoredRequest(requestContext)
	if err != nil {
		return fmt.Errorf("cannot convert request data for db: %w", err)
	}

	err = r.q.UpsertRequest(ctx, database.UpsertRequestParams{
		ID:                storedRequest.ID,
		SessionID:         storedRequest.SessionID,
		Host:              storedRequest.Host,
		Url:               storedRequest.Url,
		Method:            storedRequest.Method,
		RequestRaw:        storedRequest.RequestRaw,
		ResponseRaw:       storedRequest.ResponseRaw,
		StartTime:         storedRequest.StartTime,
		State:             storedRequest.State,
		IsFinished:        storedRequest.IsFinished,
		Metadata:          storedRequest.Metadata,
		CreatedAt:         storedRequest.CreatedAt,
		UpdatedAt:         storedRequest.UpdatedAt,
		ResponseBody:      storedRequest.ResponseBody,
		RequestBody:       storedRequest.RequestBody,
		Scheme:            storedRequest.Scheme,
		Path:              storedRequest.Path,
		Proto:             storedRequest.Proto,
		Headers:           storedRequest.Headers,
		Cookies:           storedRequest.Cookies,
		QueryParams:       storedRequest.QueryParams,
		ContentLength:     storedRequest.ContentLength,
		RemoteAddr:        storedRequest.RemoteAddr,
		RespStatus:        storedRequest.RespStatus,
		RespStatusCode:    storedRequest.RespStatusCode,
		RespProto:         storedRequest.RespProto,
		RespHeaders:       storedRequest.RespHeaders,
		RespContentLength: storedRequest.RespContentLength,
	})
	if err != nil {
		return fmt.Errorf("error on request save into db: %w", err)
	}

	return nil
}
