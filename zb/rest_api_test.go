package zb

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
)

var (
	accessKey = os.Getenv("ZB_ACCESS_KEY")
	secretKey = os.Getenv("ZB_SECRET_KEY")
)

func TestRestClient_GetSymbols(t *testing.T) {
	NewRestClient().GetSymbols()
}

func TestRestClient_GetLatestQuote(t *testing.T) {
	quote, _ := NewRestClient().GetLatestQuote("btc_usdt")
	assert.True(t, quote.Last > 0)
}

func TestRestClient_GetKlines(t *testing.T) {
	NewRestClient().GetKlines("btc_usdt", "5min", 1516029900000, 20)
}

func TestRestClient_GetTrades(t *testing.T) {
	NewRestClient().GetTrades("btc_usdt", 0)
}

func TestRestClient_GetDepth(t *testing.T) {
	depth, _ := NewRestClient().GetDepth("btc_usdt", 10)
	assert.NotNil(t, depth)

	_, err := NewRestClient().GetDepth("wrong_symbol", 10)
	assert.NotNil(t, err)
}

func TestRestClient_GetAccount(t *testing.T) {
	NewRestClient().GetAccount(accessKey, secretKey)
}
