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

	if requestContext.IsNewRequest {
		_, err := r.q.CreateRequest(ctx, database.CreateRequestParams{
			ID:           storedRequest.ID,
			SessionID:    storedRequest.SessionID,
			Host:         storedRequest.Host,
			Url:          storedRequest.Url,
			Method:       storedRequest.Method,
			RequestRaw:   storedRequest.RequestRaw,
			RequestBody:  storedRequest.RequestBody,
			ResponseRaw:  storedRequest.ResponseRaw,
			ResponseBody: storedRequest.ResponseBody,
			StartTime:    storedRequest.StartTime,
			State:        storedRequest.State,
			IsFinished:   storedRequest.IsFinished,
			Metadata:     storedRequest.Metadata,
			CreatedAt:    storedRequest.CreatedAt,
			UpdatedAt:    storedRequest.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("error on request insert: %w", err)
		}

		return nil
	}

	err = r.q.UpdateRequest(ctx, database.UpdateRequestParams{
		ID:           storedRequest.ID,
		RequestRaw:   storedRequest.RequestRaw,
		RequestBody:  storedRequest.RequestBody,
		ResponseRaw:  storedRequest.ResponseRaw,
		ResponseBody: storedRequest.ResponseBody,
		State:        storedRequest.State,
		IsFinished:   storedRequest.IsFinished,
		Metadata:     storedRequest.Metadata,
		UpdatedAt:    storedRequest.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("error on request update: %w", err)
	}
	return nil
}
