package cmd

import (
	"context"
	"net/http"

	"yxbrew/statuspage/internal/db"
	"yxbrew/statuspage/internal/logger"
	"yxbrew/statuspage/internal/router"
	"yxbrew/statuspage/internal/utils"

	"go.uber.org/zap"
)

func Start() {
	log := logger.GetLogger()
	defer logger.Sync(log)
	ctx := context.Background()

	dbCfg := db.LoadConfig()
	pool, err := db.Connect(ctx, dbCfg)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}
	log.Info("database migrations applied")

	port := utils.GetEnvOrDefault("PORT", "8080")
	addr := ":" + port

	log.Info("http server starting", zap.String("address", addr))
	if err := http.ListenAndServe(addr, router.New(pool)); err != nil {
		log.Fatal("http server failed", zap.Error(err))
	}
}
