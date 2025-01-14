package binance

import (
	"net/url"
	"strings"
)

const (
	marketWsHost         = "wss://data-stream.binance.vision"
	maxWsConnectionCount = 1024
	wsPrefix             = "stream"
	queryWsStreamName    = "?streams="
)

type Binance struct{}

func (b Binance) CreateWsUrl(streamName []string) (string, error) {
	streams := strings.Join(streamName, "/")
	endpoint, err := url.JoinPath(marketWsHost, wsPrefix)
	resUrl := strings.Join([]string{endpoint, queryWsStreamName, streams}, "")

	if err != nil {
		return "", err
	}

	return resUrl, nil
}
