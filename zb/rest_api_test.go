package zb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSymbols(t *testing.T) {
	RestClient{}.GetSymbols()
}

func TestGetLatestQuote(t *testing.T) {
	quote, _ := RestClient{}.GetLatestQuote("btc_usdt")
	assert.True(t, quote.Last > 0)
}

func TestGetKlines(t *testing.T) {
	RestClient{}.GetKlines("btc_usdt", "5min", 1516029900000, 20)
}

func TestGetTrades(t *testing.T) {
	RestClient{}.GetTrades("btc_usdt", 0)
}

func TestGetDepth(t *testing.T) {
	RestClient{}.GetDepth("btc_usdt", 10)
}
