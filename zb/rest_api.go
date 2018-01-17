package zb

import (
	json "github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/url"
	"strconv"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type RestClient struct {
}

func NewRestClient() *RestClient {
	return &RestClient{}
}

type SymbolConfig struct {
	AmountScale byte
	PriceScale  byte
}

func (c *RestClient) GetSymbols() (map[string]*SymbolConfig, error) {
	resp, err := doGet(DataApiUrl + "markets")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	err = extractError(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	configs := map[string]*SymbolConfig{}
	json.ObjectEach(body, func(key []byte, value []byte, dataType json.ValueType, offset int) error {
		symbol, _ := json.ParseString(key)
		amountScale, _ := json.GetInt(value, "amountScale")
		priceScale, _ := json.GetInt(value, "priceScale")
		configs[symbol] = &SymbolConfig{byte(amountScale), byte(priceScale)}
		return nil
	})
	return configs, nil
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

func (c *RestClient) GetLatestQuote(symbol string) (*Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", symbol)
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	err = extractError(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ticker, _, _, _ := json.Get(body, "ticker")
	volumeString, _ := json.GetString(ticker, "vol")
	lastString, _ := json.GetString(ticker, "last")
	sellString, _ := json.GetString(ticker, "sell")
	buyString, _ := json.GetString(ticker, "buy")
	highString, _ := json.GetString(ticker, "high")
	lowString, _ := json.GetString(ticker, "low")
	timeString, _ := json.GetString(body, "date")

	volume, _ := strconv.ParseFloat(volumeString, 64)
	last, _ := strconv.ParseFloat(lastString, 64)
	sell, _ := strconv.ParseFloat(sellString, 64)
	buy, _ := strconv.ParseFloat(buyString, 64)
	high, _ := strconv.ParseFloat(highString, 64)
	low, _ := strconv.ParseFloat(lowString, 64)
	time, _ := strconv.ParseUint(timeString, 10, 64)

	return &Quote{Volume: volume, Last: last, Sell: sell, Buy: buy, High: high, Low: low, Time: time}, nil
}

type Kline struct {
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
	Time   uint64
}

func (c *RestClient) GetKlines(symbol string, period string, since uint64, size uint16) ([]*Kline, error) {
	u, _ := url.Parse(DataApiUrl + "kline")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("type", period)
	q.Set("since", strconv.FormatUint(since, 10))
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	err = extractError(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var klines []*Kline
	json.ArrayEach(body, func(value []byte, dataType json.ValueType, offset int, err error) {
		time, _ := json.GetInt(value, "[0]")
		open, _ := json.GetFloat(value, "[1]")
		high, _ := json.GetFloat(value, "[2]")
		low, _ := json.GetFloat(value, "[3]")
		close, _ := json.GetFloat(value, "[4]")
		volume, _ := json.GetFloat(value, "[5]")
		klines = append(klines, &Kline{Time: uint64(time), Open: open, High: high, Low: low, Close: close, Volume: volume})
	}, "data")

	return klines, nil
}

type Trade struct {
	TradeId   uint64
	TradeType string
	Price     float64
	Amount    float64
	Time      uint64
}

func (c *RestClient) GetTrades(symbol string, since uint64) ([]*Trade, error) {
	u, _ := url.Parse(DataApiUrl + "trades")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("since", strconv.FormatUint(since, 10))
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	err = extractError(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var trades []*Trade
	json.ArrayEach(body, func(value []byte, dataType json.ValueType, offset int, err error) {
		tradeId, _ := json.GetInt(value, "tid")
		tradeType, _ := json.GetString(value, "type")
		amountString, _ := json.GetString(value, "amount")
		priceString, _ := json.GetString(value, "price")
		time, _ := json.GetInt(value, "date")

		amount, _ := strconv.ParseFloat(amountString, 64)
		price, _ := strconv.ParseFloat(priceString, 64)

		trades = append(trades, &Trade{TradeId: uint64(tradeId), TradeType: tradeType, Price: price, Amount: amount, Time: uint64(time)})
	})

	return trades, nil
}

type Depth struct {
	Asks []*DepthEntry
	Bids []*DepthEntry
	Time uint64
}

type DepthEntry struct {
	Price  float64
	Volume float64
}

func (c *RestClient) GetDepth(symbol string, size uint8) (*Depth, error) {
	u, _ := url.Parse(DataApiUrl + "depth")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := resp.Body()
	err = extractError(body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	time, _ := json.GetInt(body, "timestamp")
	asks, bids := getDepthEntries(body, "asks"), getDepthEntries(body, "bids")

	return &Depth{Asks: asks, Bids: bids, Time: uint64(time)}, nil
}

func getDepthEntries(value []byte, keys ...string) []*DepthEntry {
	var entry []*DepthEntry
	json.ArrayEach(value, func(value []byte, dataType json.ValueType, offset int, err error) {
		price, _ := json.GetFloat(value, "[0]")
		volume, _ := json.GetFloat(value, "[1]")
		entry = append(entry, &DepthEntry{Price: price, Volume: volume})
	}, keys...)
	return entry
}

func extractError(value []byte) error {
	msg, err := json.GetString(value, "error")
	if err == json.KeyPathNotFoundError {
		return nil
	}
	return &ApiError{Code: 1001, Message: msg}
}

func doGet(url string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	return resp, errors.WithStack(err)
}
