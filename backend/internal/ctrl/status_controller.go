package ctrl

import (
	"encoding/json"
	"net/http"

	"yxbrew/statuspage/internal/svc"

	"github.com/go-chi/chi/v5"
)

// StatusController handles status endpoints.
type StatusController struct {
	statusService svc.StatusService
}

// NewStatusController creates a status controller.
func NewStatusController(statusService svc.StatusService) *StatusController {
	return &StatusController{statusService: statusService}
}

// RegisterRoutes attaches status routes.
func (c *StatusController) RegisterRoutes(r chi.Router) {
	r.Get("/health", c.GetHealth)
}

// GetHealth returns service health.
func (c *StatusController) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(c.statusService.GetHealth()); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
