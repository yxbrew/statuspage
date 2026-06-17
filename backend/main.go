package main

import (
	"yxbrew/statuspage/cmd"
	"yxbrew/statuspage/internal/logger"
)

func main() {
	logger.GetLogger().Info("starting status page backend")
	cmd.Start()
}
