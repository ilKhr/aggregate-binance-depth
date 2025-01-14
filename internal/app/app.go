package app

import (
	"log/slog"

	"github.com/aggregate-binance-depth/infra"
	"github.com/aggregate-binance-depth/services"
)

/*
здесь мне нужно получить источник и цель, куда перенаправлять поток
*/

type App struct{}

func NewApp(l *slog.Logger) {
	var wsc infra.WebsocketClient

	wss, err := services.NewWsService(l, wsc)
}
