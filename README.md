# A golang client for zb.com

[![Build Status](https://travis-ci.org/berryland/zb.svg?branch=master)](https://travis-ci.org/berryland/zb)

## Set Up
```bash
dep ensure -add github.com/berryland/zb
```

## Usage
### RestClient
```go
func TestRestClient_GetLatestQuote(t *testing.T) {
    quote, err := NewRestClient().GetLatestQuote("btc_usdt")
    //other codes
    //...
}
```

### WebSocketClient
```go
func TestWebSocketClient_SubscribeQuote(t *testing.T) {
    c := NewWebSocketClient()
    c.Connect()
    c.SubscribeQuote("btc_usdt", func(quote Quote) {
        println(quote.Last)
    })
}
```
