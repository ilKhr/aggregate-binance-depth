package services

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
)

// инициируем бизнес логику
// в ней мы получаем данные из depthServiceWs
// преобразуем данные в нужный формат
// пишем данные в getDepthServiceData
// в случае отмены контекста, закрываем приложение
// если не успевают вычитывать, мы записываем данные поверх
//

type symbol = string
type price = float64

type currentDepths = map[symbol]DepthWriterRequest

type DepthGateService struct {
	log           *slog.Logger
	writers       map[string]DepthWriter
	currentDepths currentDepths
	mu            sync.Mutex
	reader        DepthReader
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
	Symbol symbol
	Bid    price
	Ask    price
}

type DepthReader interface {
	ReadJSON(target *DepthReaderResponse) error
}

type DepthWriter interface {
	WriteJSON(target DepthWriterRequest) error
	BulkWriteJSON(target []DepthWriterRequest)
}

func NewDepthGateService(l *slog.Logger, depthReader DepthReader) *DepthGateService {
	return &DepthGateService{
		reader:        depthReader,
		log:           l,
		currentDepths: make(currentDepths),
	}
}

func (d *DepthGateService) AddWriter(id string, writer DepthWriter) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.writers[id] = writer

	values := make([]DepthWriterRequest, 0, len(d.currentDepths))

	for _, value := range d.currentDepths {
		values = append(values, value)
	}

	writer.BulkWriteJSON(values)
}

func (d *DepthGateService) RemoveWriter(id string, writer DepthWriter) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.writers, id)
}

func (d *DepthGateService) convertDepthRToW(data DepthReaderResponse) (DepthWriterRequest, error) {
	const op = "internal.services.deptGate.ConvertDepthRToW"

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

func (d *DepthGateService) broadcast(data DepthWriterRequest) {
	const op = "internal.services.deptGate.Broadcast"

	logger := d.log.With(slog.String("op", op))

	for _, writers := range d.writers {
		err := writers.WriteJSON(data)

		if err != nil {
			logger.Error("error with WriteJSON", slog.String("error", err.Error()))
		}
	}
}

func (d *DepthGateService) Serve(ctx context.Context) {
	const op = "internal.services.deptGate.Serve"

	logger := d.log.With(slog.String("op", op))

	var readerResponse DepthReaderResponse
	var writerRequest DepthWriterRequest
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		default:
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

				d.broadcast(writerRequest)
			}()
		}

		logger.Debug("writerRequest", slog.Any("writerRequest", writerRequest))
	}
}
