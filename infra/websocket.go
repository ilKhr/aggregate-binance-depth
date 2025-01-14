package infra

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	handshakeTimeout    = 30 * time.Second
	closeMessageTimeout = 5 * time.Second
)

type WebsocketClient struct {
	conn *websocket.Conn
}

func (ws WebsocketClient) Connect(url string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  handshakeTimeout,
		EnableCompression: false,
	}

	conn, _, err := dialer.Dial(url, nil)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (ws *WebsocketClient) Disconnect(l *slog.Logger) error {
	const op = "infra.websocket.Disconnect"

	logger := l.With(slog.String("op", op))

	logger.Debug("start disconnect ws")

	if ws == nil {
		logger.Error("ws in undefined")

		return fmt.Errorf("ws should be defined")
	}

	if ws.conn == nil {
		logger.Error("ws conn in undefined")

		return fmt.Errorf("ws conn should be defined")
	}

	err := ws.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Goodbye"),
		time.Now().Add(closeMessageTimeout),
	)

	if err != nil {
		logger.Error("error sending close message", slog.String("error", err.Error()))
	}

	err = ws.conn.Close()

	if err != nil {
		logger.Error("error with close connection", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Debug("success disconnect ws")

	return nil
}
