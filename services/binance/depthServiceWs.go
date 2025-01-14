package binance

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/aggregate-binance-depth/services"
	"github.com/aggregate-binance-depth/services/binance/common"
)

const (
	endpoint      = "/stream"
	streamPostfix = "@depth"
)

type depthServiceWs struct {
	log     *slog.Logger
	wss     *services.WsService
	binance Binance
}

type config struct {
	symbols []string
}

type StreamData struct {
}

func NewDepthServiceWs(l *slog.Logger, cfg config, wss *services.WsService) (*depthServiceWs, error) {
	const op = "services.binance.depthServiceWs"

	logger := l.With(slog.String("op", op))

	var binance Binance

	streamNames := make([]string, len(cfg.symbols))

	for symbol := range slices.Values(cfg.symbols) {
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

	if err := wss.Connect(l, services.WssConfig{Url: url}); err != nil {
		logger.Error("error with init connection", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &depthServiceWs{wss: wss, log: l}, nil
}

func (d *depthServiceWs) ReadJSON(target []any) error {
	const op = "services.binance.ReadJSON"

	logger := d.log.With(slog.String("op", op))

	if err := d.wss.ReadJSON(target); err != nil {
		logger.Error("error with readJSON", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func validateSymbol(string) error {
	return nil
}

// DepthResponse define depth info with bids and asks
type DepthResponse struct {
	LastUpdateID int64 `json:"lastUpdateId"`
	Bids         []Bid `json:"bids"`
	Asks         []Ask `json:"asks"`
}

// Ask is a type alias for PriceLevel.
type Ask = common.PriceLevel

// Bid is a type alias for PriceLevel.
type Bid = common.PriceLevel
