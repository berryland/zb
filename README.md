# A golang client for ZB.com

## Set Up
```bash
go get -u github.com/pojozhang/exchange/zb
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
    c.Start()
    c.SubscribeQuote("btc_usdt", func(quote *Quote) {
        println(quote.Last)
    })
}
```
