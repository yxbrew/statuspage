package svc

import "yxbrew/statuspage/internal/model"

// StatusService defines business logic for status endpoints.
type StatusService interface {
	GetHealth() model.HealthStatus
}

type statusService struct{}

// NewStatusService creates a status service.
func NewStatusService() StatusService {
	return &statusService{}
}

func (s *statusService) GetHealth() model.HealthStatus {
	return model.HealthStatus{
		Status:  "ok",
		Service: "statuspage-backend",
		Version: "1.0.0",
	}
}
