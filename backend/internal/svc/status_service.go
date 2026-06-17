package svc

import (
	"context"
	"errors"

	"yxbrew/statuspage/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatusService defines business logic for status endpoints.
type StatusService interface {
	GetHealth(ctx context.Context) (model.HealthStatus, error)
}

type statusService struct {
	db *pgxpool.Pool
}

// NewStatusService creates a status service.
func NewStatusService(db *pgxpool.Pool) StatusService {
	return &statusService{db: db}
}

func (s *statusService) GetHealth(ctx context.Context) (model.HealthStatus, error) {
	health := model.HealthStatus{
		Status:  "ok",
		Service: "statuspage-backend",
		Version: "1.0.0",
	}

	if s.db == nil {
		health.Status = "down"
		return health, errors.New("database pool is not initialized")
	}

	if err := s.db.Ping(ctx); err != nil {
		health.Status = "down"
		return health, err
	}

	return health, nil
}
