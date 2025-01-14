package services

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type WsService struct {
	log             *slog.Logger
	wsRWConnCreator wsRWConnCreator
	conn            WsConnection
}

type WsConnection interface {
	ReadJSON(v interface{}) error
	Disconnect(l *slog.Logger) error
}

type wsRWConnCreator interface {
	Connect(url string) (WsConnection, error)
}

func NewWsService(l *slog.Logger, wsRWConnCreator wsRWConnCreator) (*WsService, error) {
	const op = "services.websocket.NewWsService"

	logger := l.With(slog.String("op", op))

	logger.Debug("Start init ws service")

	return &WsService{log: l, conn: nil, wsRWConnCreator: wsRWConnCreator}, nil
}

func (s *WsService) Connect(l *slog.Logger, url string) error {

	const op = "services.websocket.Connect"

	logger := l.With(slog.String("op", op))

	if s.conn != nil {
		logger.Error("connection already exists")

		return fmt.Errorf("%s: %s", op, "connection already exists")
	}

	conn, err := s.wsRWConnCreator.Connect(url)

	s.conn = conn

	if err != nil {
		logger.Error("error with init connection", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *WsService) Disconnect() error {
	const op = "services.websocket.Disconnect"

	logger := s.log.With(slog.String("op", op))

	if s.conn == nil {
		logger.Error("connection not exists")

		return fmt.Errorf("%s: %s", op, "connection not exists")
	}

	err := s.conn.Disconnect(s.log)

	if err != nil {
		logger.Error("error with disconnect", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *WsService) ReadJSON(target []any) error {
	const op = "services.websocket.ReadJSON"

	logger := s.log.With(slog.String("op", op))

	if s.conn == nil {
		logger.Error("connection not exists")

		return fmt.Errorf("%s: %s", op, "connection not exists")
	}

	// blocks flow until the WS is closed with
	err := s.conn.ReadJSON(target)

	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			logger.Debug("Success end ReadJSON", slog.String("error", err.Error()))
		} else {
			logger.Error("error with ReadJSON", slog.String("error", err.Error()))

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
