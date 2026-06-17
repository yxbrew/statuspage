package router

import (
	"net/http"

	"yxbrew/statuspage/internal/ctrl"
	"yxbrew/statuspage/internal/svc"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// New creates a common HTTP router for the backend service.
func New(dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Recoverer)

	statusService := svc.NewStatusService(dbPool)
	statusController := ctrl.NewStatusController(statusService)

	r.Route("/api/v1", func(api chi.Router) {
		statusController.RegisterRoutes(api)
	})

	r.Get("/swagger/doc.json", swaggerDocHandler)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	return r
}

func swaggerDocHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(swaggerDoc))
}
