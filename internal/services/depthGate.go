package services

import (
	"log/slog"
	"strconv"
	"sync"
)

type symbol = string
type price = float64

type currentDepths = map[symbol]DepthWriterRequest

type DepthGateService struct {
	log           *slog.Logger
	currentDepths currentDepths
	mu            sync.Mutex
	reader        DepthReader
	writer        DepthWriter
	closed        bool
}

type DepthReaderResponse struct {
	Stream string
	Data   struct {
		Symbol symbol
		Bids   [][]string
		Asks   [][]string
	}
}

type DepthWriterRequest struct {
	Symbol symbol `json:"symbol"`
	Bid    price  `json:"bid"`
	Ask    price  `json:"ask"`
}

type DepthReader interface {
	ReadJSON(target *DepthReaderResponse) error
}

type DepthWriter interface {
	WriteJSON(target DepthWriterRequest) error
	BulkWriteJSON(target []DepthWriterRequest) error
}

func NewDepthGateService(l *slog.Logger, depthReader DepthReader, depthWriter DepthWriter) *DepthGateService {
	return &DepthGateService{
		reader:        depthReader,
		writer:        depthWriter,
		log:           l,
		currentDepths: make(currentDepths),
	}
}

func (d *DepthGateService) convertDepthRToW(data DepthReaderResponse) (DepthWriterRequest, error) {
	const op = "internal.services.depthGate.ConvertDepthRToW"

	logger := d.log.With(slog.String("op", op))

	ask := "0"
	if len(data.Data.Asks) > 0 && len(data.Data.Asks[0]) > 0 {
		ask = data.Data.Asks[0][0]
	}

	bid := "0"
	if len(data.Data.Bids) > 0 && len(data.Data.Bids[0]) > 0 {
		bid = data.Data.Bids[0][0]
	}

	askPrice, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		logger.Error("error with strconv.ParseFloat ask", slog.String("error", err.Error()))
	}

	bidsPrice, err := strconv.ParseFloat(bid, 64)
	if err != nil {
		logger.Error("error with strconv.ParseFloat bids", slog.String("error", err.Error()))
	}

	return DepthWriterRequest{Symbol: data.Data.Symbol, Bid: bidsPrice, Ask: askPrice}, nil
}

func (d *DepthGateService) Serve() {
	const op = "internal.services.depthGate.Serve"

	logger := d.log.With(slog.String("op", op))

	if d.closed {
		logger.Error("server already closed")

		return
	}

	var readerResponse DepthReaderResponse
	var writerRequest DepthWriterRequest
	var err error

	for !(d.closed) {
		func() {
			if err = d.reader.ReadJSON(&readerResponse); err != nil {
				logger.Error("error with ReadJSON", slog.String("error", err.Error()))
				return
			}

			writerRequest, err = d.convertDepthRToW(readerResponse)

			if err != nil {
				logger.Error("error with convertDepthRToW", slog.String("error", err.Error()))
				return
			}

			d.mu.Lock()
			defer d.mu.Unlock()

			d.currentDepths[writerRequest.Symbol] = writerRequest

			d.writer.WriteJSON(writerRequest)
		}()

		logger.Debug("writerRequest", slog.Any("writerRequest", writerRequest))
	}

	logger.Info("serve stoped")
}

func (d *DepthGateService) Shutdown() {
	const op = "internal.services.depthGate.Serve"

	logger := d.log.With(slog.String("op", op))

	logger.Info("Shutting down...")

	d.closed = true

	logger.Info("Shutting success")
}

func (d *DepthGateService) WriteCurrentDeps() error {
	const op = "internal.services.depthGate.WriteCurrentDeps"

	logger := d.log.With(slog.String("op", op))

	d.mu.Lock()
	defer d.mu.Unlock()

	values := make([]DepthWriterRequest, 0, len(d.currentDepths))

	for _, value := range d.currentDepths {
		values = append(values, value)
	}

	if err := d.writer.BulkWriteJSON(values); err != nil {
		logger.Error("error with WriteJSON", slog.String("error", err.Error()))

		return err
	}

	return nil
}
