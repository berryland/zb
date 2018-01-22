package zb

import (
	json "github.com/buger/jsonparser"
	"strconv"
)

type SymbolConfig struct {
	AmountScale byte
	PriceScale  byte
}

type Quote struct {
	Volume float64
	Last   float64
	Sell   float64
	Buy    float64
	High   float64
	Low    float64
	Time   uint64
}

func marshalQuote(value []byte) Quote {
	ticker, _, _, _ := json.Get(value, "ticker")
	volumeString, _ := json.GetString(ticker, "vol")
	lastString, _ := json.GetString(ticker, "last")
	sellString, _ := json.GetString(ticker, "sell")
	buyString, _ := json.GetString(ticker, "buy")
	highString, _ := json.GetString(ticker, "high")
	lowString, _ := json.GetString(ticker, "low")
	timeString, _ := json.GetString(value, "date")

	volume, _ := strconv.ParseFloat(volumeString, 64)
	last, _ := strconv.ParseFloat(lastString, 64)
	sell, _ := strconv.ParseFloat(sellString, 64)
	buy, _ := strconv.ParseFloat(buyString, 64)
	high, _ := strconv.ParseFloat(highString, 64)
	low, _ := strconv.ParseFloat(lowString, 64)
	time, _ := strconv.ParseUint(timeString, 10, 64)

	return Quote{Volume: volume, Last: last, Sell: sell, Buy: buy, High: high, Low: low, Time: time}
}

type Kline struct {
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
	Time   uint64
}

type Trade struct {
	Id        uint64
	TradeType TradeType
	Price     float64
	Amount    float64
	Time      uint64
}

type TradeType int8

const (
	All  TradeType = iota - 1
	Sell
	Buy
)

func ParseTradeType(string string) TradeType {
	switch string {
	case "buy":
		return Buy
	case "sell":
		return Sell
	default:
		panic("Unknown trade type: " + string)
	}
}

type Depth struct {
	Asks []DepthEntry
	Bids []DepthEntry
	Time uint64
}

type DepthEntry struct {
	Price  float64
	Volume float64
}

func marshalDepthEntries(value []byte, keys ...string) []DepthEntry {
	var entry []DepthEntry
	json.ArrayEach(value, func(value []byte, dataType json.ValueType, offset int, err error) {
		price, _ := json.GetFloat(value, "[0]")
		volume, _ := json.GetFloat(value, "[1]")
		entry = append(entry, DepthEntry{Price: price, Volume: volume})
	}, keys...)
	return entry
}

type Account struct {
	Username             string
	TradePasswordEnabled bool
	AuthGoogleEnabled    bool
	AuthMobileEnabled    bool
	Assets               []Asset
}

type Asset struct {
	Freeze    float64
	Available float64
	Coin      Coin
}

type Coin struct {
	CnName string
	EnName string
	Key    string
	Unit   string
	Scale  uint8
}

type Order struct {
	Id          uint64
	Price       float64
	Average     float64
	TotalAmount float64
	TradeAmount float64
	TradeMoney  float64
	Symbol      string
	Status      OrderStatus
	TradeType   TradeType
	Time        uint64
}

type OrderStatus uint8

const (
	Pending         OrderStatus = iota
	Cancelled
	Finished
	PartiallyFilled
)