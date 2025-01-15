package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	internalServices "github.com/aggregate-binance-depth/internal/services"
	"github.com/gorilla/websocket"
)

type id = int

type client struct {
	id   id
	conn *websocket.Conn
}

type WebsocketServer struct {
	log              *slog.Logger
	upgrader         websocket.Upgrader
	clients          map[id]client
	depthGateService depthGateService
	mu               sync.Mutex
	server           *http.Server

	// TODO: take id from connection
	maxId int
}

type depthGateService interface {
	WriteCurrentDeps() error
}

func (ws *WebsocketServer) WriteJSON(target internalServices.DepthWriterRequest) error {
	const op = "services.ws.WriteJSON"

	logger := ws.log.With(slog.String("op", op))

	for _, client := range ws.clients {
		if err := client.conn.WriteJSON(target); err != nil {
			logger.Error("error with WriteJSON", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (ws *WebsocketServer) BulkWriteJSON(target []internalServices.DepthWriterRequest) error {
	const op = "services.ws.BulkWriteJSON"

	logger := ws.log.With(slog.String("op", op))

	for _, client := range ws.clients {
		if err := client.conn.WriteJSON(target); err != nil {
			logger.Error("error with WriteJSON", slog.String("error", err.Error()))
			// TODO: collect error
		}
	}

	return nil
}

func (ws *WebsocketServer) RegisterDepthGateService(dgs depthGateService) error {
	const op = "services.ws.RegisterDepthGateService"

	logger := ws.log.With(slog.String("op", op))

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.depthGateService != nil {
		logger.Error("ws already set")

		return fmt.Errorf("ws already set")
	}

	ws.depthGateService = dgs

	return nil
}

func NewWebsocketServer(l *slog.Logger) *WebsocketServer {
	return &WebsocketServer{
		clients: make(map[id]client),
		log:     l,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (ws *WebsocketServer) handleClient(conn *websocket.Conn) {
	const op = "services.websocket.handleClient"

	logger := ws.log.With(slog.String("op", op))

	clientId := ws.maxId

	func() {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		ws.clients[clientId] = client{id: clientId, conn: conn}
		ws.maxId++
	}()

	logger.Debug("client connected")

	defer func() {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		delete(ws.clients, clientId)
		logger.Debug("client disconnected")
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error("error recive message", slog.String("error", err.Error()))

			break
		}

		logger.Debug("received", slog.Any("message", message), slog.Int("messageType", messageType))
	}
}

func (ws *WebsocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	const op = "services.websocket.HandleWebSocket"

	logger := ws.log.With(slog.String("op", op))

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade error", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	ws.handleClient(conn)

	return
}

func (ws *WebsocketServer) Serve() {
	const op = "services.ws.Serve"

	logger := ws.log.With(slog.String("op", op))

	http.HandleFunc("/ws", ws.HandleWebSocket)

	logger.Info("starting WebSocket server on :8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.HandleWebSocket)

	ws.server = &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	if err := ws.server.ListenAndServe(); err != nil {
		logger.Error("starting WebSocket server on :8080")
	}
}

func (ws *WebsocketServer) Shutdown(ctx context.Context) {
	const op = "services.ws.Serve"

	logger := ws.log.With(slog.String("op", op))
	logger.Debug("Shutting down server...")

	if err := ws.server.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", slog.String("error", err.Error()))
	}

	func() {
		ws.mu.Lock()
		defer ws.mu.Unlock()

		for _, client := range ws.clients {
			client.conn.Close()
		}
	}()

	logger.Info("Server stopped")
}
