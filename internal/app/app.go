package app

import (
	"fmt"
	"log/slog"

	"github.com/aggregate-binance-depth/infra"
	"github.com/aggregate-binance-depth/internal/adapters"
	internalServices "github.com/aggregate-binance-depth/internal/services"
	"github.com/aggregate-binance-depth/services"
	"github.com/aggregate-binance-depth/services/binance"
	"github.com/aggregate-binance-depth/ws"
)

type App struct {
	DepthGateService *internalServices.DepthGateService
	WsServer         *ws.WebsocketServer
	Wss              *services.WsService
}

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

	wsServer := ws.NewWebsocketServer(l)

	depthGateService := internalServices.NewDepthGateService(l, &adapters.DepthServiceWsAdapter{DepthService: depthServiceWs}, wsServer)

	wsServer.RegisterDepthGateService(depthGateService)

	return &App{
		DepthGateService: depthGateService,
		WsServer:         wsServer,
		Wss:              wss,
	}, nil
}
