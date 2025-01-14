package main

import (
	"log/slog"
	"os"

	"github.com/aggregate-binance-depth/internal/app"
	"github.com/aggregate-binance-depth/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	config := config.MustLoad()

	log := setupLogger(config.Env)

	log.Info("logger init successfull")

	_, err := app.NewApp(log, config.Binance.Depth.Symbols)

	if err != nil {
		log.Error("main error", slog.String("error", err.Error()))
	}

	// go application.GRPCSrv.MustRun()

	// Graceful shutdown
	// stop := make(chan os.Signal, 1)

	// signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// sign := <-stop

	// log.Info("application stoping", slog.String("signal", sign.String()))

	// application.GRPCSrv.Stop()

	log.Info("application stoped.")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
