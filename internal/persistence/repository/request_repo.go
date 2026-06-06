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
		ID:         storedRequest.ID,
		SessionID:  storedRequest.SessionID,
		StartTime:  storedRequest.StartTime,
		State:      storedRequest.State,
		IsFinished: storedRequest.IsFinished,
		Metadata:   storedRequest.Metadata,
		CreatedAt:  storedRequest.CreatedAt,
		UpdatedAt:  storedRequest.UpdatedAt,

		ReqHost:          storedRequest.ReqHost,
		ReqUrl:           storedRequest.ReqUrl,
		ReqMethod:        storedRequest.ReqMethod,
		ReqBody:          storedRequest.ReqBody,
		ReqScheme:        storedRequest.ReqScheme,
		ReqPath:          storedRequest.ReqPath,
		ReqProto:         storedRequest.ReqProto,
		ReqHeaders:       storedRequest.ReqHeaders,
		ReqCookies:       storedRequest.ReqCookies,
		ReqQueryParams:   storedRequest.ReqQueryParams,
		ReqContentLength: storedRequest.ReqContentLength,
		ReqRemoteAddr:    storedRequest.ReqRemoteAddr,

		ResStatus:        storedRequest.ResStatus,
		ResStatusCode:    storedRequest.ResStatusCode,
		ResProto:         storedRequest.ResProto,
		ResHeaders:       storedRequest.ResHeaders,
		ResContentLength: storedRequest.ResContentLength,
		ResBody:          storedRequest.ResBody,
	})
	if err != nil {
		return fmt.Errorf("error on request save into db: %w", err)
	}

	return nil
}
