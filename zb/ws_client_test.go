package zb

import (
	"testing"
	"time"
)

func TestWebSocketClient_SubscribeQuote(t *testing.T) {
	NewWebSocketClient().SubscribeQuote()

	for {
		time.Sleep(5 * time.Second)
	}
}
