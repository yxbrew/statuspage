package cmd

import (
	"net/http"

	"yxbrew/statuspage/internal/logger"
	"yxbrew/statuspage/internal/router"
	"yxbrew/statuspage/internal/utils"

	"go.uber.org/zap"
)

func Start() {
	log := logger.GetLogger()
	defer logger.Sync(log)

	port := utils.GetEnvOrDefault("PORT", "8080")
	addr := ":" + port

	log.Info("http server starting", zap.String("address", addr))
	if err := http.ListenAndServe(addr, router.New()); err != nil {
		log.Fatal("http server failed", zap.Error(err))
	}
}
