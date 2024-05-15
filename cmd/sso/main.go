package main

import (
	"GrpcAuth/internal/app"
	"GrpcAuth/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	EnvLocal = "local"
	EnvProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := SetupLogger(cfg.Env)
	log.Info("starting server", slog.Any("config", cfg))

	application := app.New(log, cfg.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustStart()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	application.GRPCSrv.Stop()
	log.Info("server stopped")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case EnvLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case EnvProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
