package binance

import "net/url"

const (
	marketWsHost         = "wss://data-stream.binance.vision"
	maxWsConnectionCount = 1024
)

type Binance struct{}

func (b Binance) CreateWsUrl(streamName []string) (string, error) {
	url, err := url.JoinPath(marketWsHost, streamName...)

	if err != nil {
		return "", err
	}

	return url, nil
}
