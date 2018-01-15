package zb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSymbols(t *testing.T) {
	GetSymbols()
}

func TestGetLatestQuote(t *testing.T) {
	quote, _ := GetLatestQuote("btc_usdt")
	assert.True(t, quote.Last > 0)
}
