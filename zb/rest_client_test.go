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
	klines, _ := NewRestClient().GetKlines("btc_usdt", "5min", 1516029900000, 20)
	assert.True(t, klines[0].High > 0)
}

func TestRestClient_GetTrades(t *testing.T) {
	trades, _ := NewRestClient().GetTrades("btc_usdt", 0)
	assert.True(t, trades[0].Price > 0)
}

func TestRestClient_GetDepth(t *testing.T) {
	depth, _ := NewRestClient().GetDepth("btc_usdt", 10)
	assert.NotNil(t, depth)
	assert.True(t, depth.Time > 0)

	_, err := NewRestClient().GetDepth("wrong_symbol", 10)
	assert.NotNil(t, err)
}

func TestRestClient_GetAccount(t *testing.T) {
	account, _ := NewRestClient().GetAccount(accessKey, secretKey)
	assert.NotNil(t, account.Username)
}

func TestRestClient_GetOrders(t *testing.T) {
	NewRestClient().GetOrders("btc_usdt", All, 0, 10, accessKey, secretKey)
}

func TestRestClient_GetOrder(t *testing.T) {
	NewRestClient().GetOrder("btc_usdt", 2018012160893558, accessKey, secretKey)
}

func TestRestClient_PlaceOrder(t *testing.T) {
	NewRestClient().PlaceOrder("btc_usdt", 15000, 0.01, Sell, accessKey, secretKey)
}

func TestRestClient_CancelOrder(t *testing.T) {
	NewRestClient().CancelOrder("btc_usdt", 2018012261281063, accessKey, secretKey)
}
