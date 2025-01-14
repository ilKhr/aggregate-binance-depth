package infra

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aggregate-binance-depth/services"
	"github.com/gorilla/websocket"
)

const (
	handshakeTimeout    = 30 * time.Second
	closeMessageTimeout = 5 * time.Second
)

type WebsocketConnection struct {
	*websocket.Conn
}

func (ws WebsocketConnection) Connect(url string) (services.WsConnection, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  handshakeTimeout,
		EnableCompression: false,
	}

	conn, _, err := dialer.Dial(url, nil)

	if err != nil {
		return nil, err
	}

	return &WebsocketConnection{Conn: conn}, nil
}

func (ws *WebsocketConnection) Disconnect(l *slog.Logger) error {
	const op = "infra.websocket.Disconnect"

	logger := l.With(slog.String("op", op))

	logger.Debug("start disconnect ws")

	if ws == nil {
		logger.Error("ws in undefined")

		return fmt.Errorf("ws should be defined")
	}

	err := ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Goodbye"),
		time.Now().Add(closeMessageTimeout),
	)

	if err != nil {
		logger.Error("error sending close message", slog.String("error", err.Error()))
	}

	err = ws.Close()

	if err != nil {
		logger.Error("error with close connection", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Debug("success disconnect ws")

	return nil
}
