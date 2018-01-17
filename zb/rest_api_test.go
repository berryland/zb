package zb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSymbols(t *testing.T) {
	NewRestClient().GetSymbols()
}

func TestGetLatestQuote(t *testing.T) {
	quote, _ := NewRestClient().GetLatestQuote("btc_usdt")
	assert.True(t, quote.Last > 0)
}

func TestGetKlines(t *testing.T) {
	NewRestClient().GetKlines("btc_usdt", "5min", 1516029900000, 20)
}

func TestGetTrades(t *testing.T) {
	NewRestClient().GetTrades("btc_usdt", 0)
}

func TestGetDepth(t *testing.T) {
	depth, _ := NewRestClient().GetDepth("btc_usdt", 10)
	assert.NotNil(t, depth)

	_, err := NewRestClient().GetDepth("wrong_symbol", 10)
	assert.NotNil(t, err)
}
