package binance

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/aggregate-binance-depth/services"
	"github.com/aggregate-binance-depth/services/binance/common"
)

const (
	streamPostfix = "@depth"
)

type DepthServiceWs struct {
	log     *slog.Logger
	wss     *services.WsService
	binance Binance
}

type StreamData struct {
}

func NewDepthServiceWs(l *slog.Logger, symbols []string, wss *services.WsService) (*DepthServiceWs, error) {
	const op = "services.binance.depthServiceWs"

	logger := l.With(slog.String("op", op))

	var binance Binance

	streamNames := make([]string, 0, len(symbols))

	for symbol := range slices.Values(symbols) {
		if err := validateSymbol(symbol); err != nil {
			logger.Error("symbol is not valid", slog.String("symbol", symbol))

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		streamNames = append(streamNames, strings.Join([]string{symbol, streamPostfix}, ""))
	}

	url, err := binance.CreateWsUrl(streamNames)

	if err != nil {
		logger.Error("error with create wsl url", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := wss.Connect(l, url); err != nil {
		logger.Error("error with init connection", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &DepthServiceWs{wss: wss, log: l}, nil
}

func (d *DepthServiceWs) ReadJSON(target *DepthStreamResponse) error {
	const op = "services.binance.ReadJSON"

	logger := d.log.With(slog.String("op", op))

	t, r, err := d.wss.ReadMessage()

	logger.Debug("Type", slog.Any("t", t))

	if err != nil {
		logger.Error("error with ReadMessage", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := json.Unmarshal(r, target); err != nil {
		logger.Error("error with Unmarshal", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func validateSymbol(string) error {
	return nil
}

// DepthStreamResponse represents the entire WebSocket message
type DepthStreamResponse struct {
	Stream string `json:"stream"`
	Data   struct {
		Symbol string     `json:"s"` // Symbol (e.g., "BTCUSDT")
		Bids   [][]string `json:"b"` // Bids (array of [price, quantity])
		Asks   [][]string `json:"a"` // Asks (array of [price, quantity])
	} `json:"data"`
}

// Ask is a type alias for PriceLevel.
type Ask = common.PriceLevel

// Bid is a type alias for PriceLevel.
type Bid = common.PriceLevel
