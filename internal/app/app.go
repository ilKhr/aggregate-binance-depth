package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aggregate-binance-depth/infra"
	"github.com/aggregate-binance-depth/internal/adapters"
	internalServices "github.com/aggregate-binance-depth/internal/services"
	"github.com/aggregate-binance-depth/services"
	"github.com/aggregate-binance-depth/services/binance"
)

/*
здесь мне нужно получить источник и цель, куда перенаправлять поток
*/

type App struct{}

func NewApp(l *slog.Logger, symbols []string) (*App, error) {
	const op = "internal.app.NewApp"

	logger := l.With(slog.String("op", op))

	var wsc infra.WebsocketConnection

	wss, err := services.NewWsService(l, wsc)

	if err != nil {
		logger.Error("error with create websocket service", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	depthServiceWs, err := binance.NewDepthServiceWs(l, symbols, wss)

	if err != nil {
		logger.Error("error with create depth service", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	depthGateService := internalServices.NewDepthGateService(l, &adapters.DepthServiceWsAdapter{DepthService: depthServiceWs})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go depthGateService.Serve(ctx)

	wss.Disconnect()

	return &App{}, nil
}
