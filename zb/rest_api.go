package zb

import (
	"github.com/valyala/fasthttp"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type MarketConfig struct {
	AmountScale byte
	PriceScale  byte
}

func GetMarkets() *map[string]MarketConfig {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(DataApiUrl + "markets")
	return nil
}
