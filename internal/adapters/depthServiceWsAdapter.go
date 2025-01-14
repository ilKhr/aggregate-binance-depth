package adapters

import (
	internalServices "github.com/aggregate-binance-depth/internal/services"
	"github.com/aggregate-binance-depth/services/binance"
)

type DepthServiceWsAdapter struct {
	DepthService *binance.DepthServiceWs
}

func (a *DepthServiceWsAdapter) ReadJSON(target *internalServices.DepthReaderResponse) error {
	var streamResp binance.DepthStreamResponse
	err := a.DepthService.ReadJSON(&streamResp)
	if err != nil {
		return err
	}

	target.Stream = streamResp.Stream
	target.Data.Symbol = streamResp.Data.Symbol
	target.Data.Bids = streamResp.Data.Bids
	target.Data.Asks = streamResp.Data.Asks
	return nil
}
