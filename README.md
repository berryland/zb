# A Golang Client For ZB.com

[![Build Status](https://travis-ci.org/pojozhang/exchange.svg?branch=master)](https://travis-ci.org/pojozhang/exchange)

## Set Up
```bash
dep ensure -add github.com/pojozhang/exchange/zb
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
