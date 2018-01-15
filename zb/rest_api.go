package zb

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/url"
	"strconv"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type SymbolConfig struct {
	AmountScale byte
	PriceScale  byte
}

func GetSymbols() (*map[string]SymbolConfig, error) {
	resp, err := doGet(DataApiUrl + "markets")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	configs := map[string]SymbolConfig{}
	jsonparser.ObjectEach(resp.Body(), func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		symbol, _ := jsonparser.ParseString(key)
		amountScale, _ := jsonparser.GetInt(value, "amountScale")
		priceScale, _ := jsonparser.GetInt(value, "priceScale")
		configs[symbol] = SymbolConfig{byte(amountScale), byte(priceScale)}
		return nil
	})
	return &configs, nil
}

type Quote struct {
	Volume    float32
	Last      float32
	Sell      float32
	Buy       float32
	High      float32
	Low       float32
	Timestamp uint64
}

func GetLatestQuote(symbol string) (*Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", symbol)
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	bytes := resp.Body()
	ticker, _, _, _ := jsonparser.Get(bytes, "ticker")
	volString, _ := jsonparser.GetString(ticker, "vol")
	lastString, _ := jsonparser.GetString(ticker, "last")
	sellString, _ := jsonparser.GetString(ticker, "sell")
	buyString, _ := jsonparser.GetString(ticker, "buy")
	highString, _ := jsonparser.GetString(ticker, "high")
	lowString, _ := jsonparser.GetString(ticker, "low")
	dateString, _ := jsonparser.GetString(bytes, "date")

	vol, _ := strconv.ParseFloat(volString, 32)
	last, _ := strconv.ParseFloat(lastString, 32)
	sell, _ := strconv.ParseFloat(sellString, 32)
	buy, _ := strconv.ParseFloat(buyString, 32)
	high, _ := strconv.ParseFloat(highString, 32)
	low, _ := strconv.ParseFloat(lowString, 32)
	date, _ := strconv.ParseUint(dateString, 10, 64)

	return &Quote{Volume: float32(vol), Last: float32(last), Sell: float32(sell), Buy: float32(buy), High: float32(high), Low: float32(low), Timestamp: date}, nil
}

func GetKlines(symbol string, peroid string, since uint64, size uint16) {
}

func doGet(url string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	return resp, errors.WithStack(err)
}
