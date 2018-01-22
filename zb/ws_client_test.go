package zb

import (
	"testing"
	"time"
)

func TestWebSocketClient_SubscribeQuote(t *testing.T) {
	c := NewWebSocketClient()
	c.Start()
	c.SubscribeQuote("btc_usdt", func(quote Quote) {
		println(quote.Time)
		c.Stop()
	})

	for {
		time.Sleep(5 * time.Second)
	}
}
