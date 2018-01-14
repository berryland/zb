package zb

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetCurrencyConfigs(t *testing.T) {
	GetContractConfigs()
}

func TestGetLatestQuote(t *testing.T) {
	quote, _ := GetLatestQuote("btc_usdt")
	assert.True(t, quote.Last > 0)
}
