package zb

import (
	json "github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"crypto/sha1"
	"crypto/hmac"
	"crypto/md5"
	"fmt"
	"time"
	"sort"
	"strings"
	"net/http"
	"io/ioutil"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type RestClient struct {
	client *http.Client
}

func NewRestClient() *RestClient {
	c := new(RestClient)
	c.client = &http.Client{}
	return c
}

type SymbolConfig struct {
	AmountScale byte
	PriceScale  byte
}

func (c *RestClient) GetSymbols() (map[string]SymbolConfig, error) {
	configs := map[string]SymbolConfig{}
	resp, err := c.doGet(DataApiUrl + "markets")
	if err != nil {
		return configs, errors.WithStack(err)
	}

	bytes := resp.Bytes()
	err = extractError(bytes)
	if err != nil {
		return configs, errors.WithStack(err)
	}

	json.ObjectEach(bytes, func(key []byte, value []byte, dataType json.ValueType, offset int) error {
		symbol, _ := json.ParseString(key)
		amountScale, _ := json.GetInt(value, "amountScale")
		priceScale, _ := json.GetInt(value, "priceScale")
		configs[symbol] = SymbolConfig{byte(amountScale), byte(priceScale)}
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

func (c *RestClient) GetLatestQuote(symbol string) (Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", symbol)
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Quote{}, errors.WithStack(err)
	}

	bytes := resp.Bytes()
	err = extractError(bytes)
	if err != nil {
		return Quote{}, errors.WithStack(err)
	}

	ticker, _, _, _ := json.Get(bytes, "ticker")
	volumeString, _ := json.GetString(ticker, "vol")
	lastString, _ := json.GetString(ticker, "last")
	sellString, _ := json.GetString(ticker, "sell")
	buyString, _ := json.GetString(ticker, "buy")
	highString, _ := json.GetString(ticker, "high")
	lowString, _ := json.GetString(ticker, "low")
	timeString, _ := json.GetString(bytes, "date")

	volume, _ := strconv.ParseFloat(volumeString, 64)
	last, _ := strconv.ParseFloat(lastString, 64)
	sell, _ := strconv.ParseFloat(sellString, 64)
	buy, _ := strconv.ParseFloat(buyString, 64)
	high, _ := strconv.ParseFloat(highString, 64)
	low, _ := strconv.ParseFloat(lowString, 64)
	time, _ := strconv.ParseUint(timeString, 10, 64)

	return Quote{Volume: volume, Last: last, Sell: sell, Buy: buy, High: high, Low: low, Time: time}, nil
}

type Kline struct {
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
	Time   uint64
}

func (c *RestClient) GetKlines(symbol string, period string, since uint64, size uint16) ([]Kline, error) {
	var klines []Kline
	u, _ := url.Parse(DataApiUrl + "kline")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("type", period)
	q.Set("since", strconv.FormatUint(since, 10))
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return klines, errors.WithStack(err)
	}

	bytes := resp.Bytes()
	err = extractError(bytes)
	if err != nil {
		return klines, errors.WithStack(err)
	}

	json.ArrayEach(bytes, func(value []byte, dataType json.ValueType, offset int, err error) {
		time, _ := json.GetInt(value, "[0]")
		open, _ := json.GetFloat(value, "[1]")
		high, _ := json.GetFloat(value, "[2]")
		low, _ := json.GetFloat(value, "[3]")
		close, _ := json.GetFloat(value, "[4]")
		volume, _ := json.GetFloat(value, "[5]")
		klines = append(klines, Kline{Time: uint64(time), Open: open, High: high, Low: low, Close: close, Volume: volume})
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

func (c *RestClient) GetTrades(symbol string, since uint64) ([]Trade, error) {
	var trades []Trade
	u, _ := url.Parse(DataApiUrl + "trades")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("since", strconv.FormatUint(since, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return trades, errors.WithStack(err)
	}

	bytes := resp.Bytes()
	err = extractError(bytes)
	if err != nil {
		return trades, errors.WithStack(err)
	}

	json.ArrayEach(bytes, func(value []byte, dataType json.ValueType, offset int, err error) {
		tradeId, _ := json.GetInt(value, "tid")
		tradeType, _ := json.GetString(value, "type")
		amountString, _ := json.GetString(value, "amount")
		priceString, _ := json.GetString(value, "price")
		time, _ := json.GetInt(value, "date")

		amount, _ := strconv.ParseFloat(amountString, 64)
		price, _ := strconv.ParseFloat(priceString, 64)

		trades = append(trades, Trade{TradeId: uint64(tradeId), TradeType: tradeType, Price: price, Amount: amount, Time: uint64(time)})
	})

	return trades, nil
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

func (c *RestClient) GetDepth(symbol string, size uint8) (Depth, error) {
	u, _ := url.Parse(DataApiUrl + "depth")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Depth{}, errors.WithStack(err)
	}

	bytes := resp.Bytes()
	err = extractError(bytes)
	if err != nil {
		return Depth{}, errors.WithStack(err)
	}

	time, _ := json.GetInt(bytes, "timestamp")
	asks, bids := getDepthEntries(bytes, "asks"), getDepthEntries(bytes, "bids")

	return Depth{Asks: asks, Bids: bids, Time: uint64(time)}, nil
}

func getDepthEntries(value []byte, keys ...string) []DepthEntry {
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

func (c *RestClient) GetAccount(accessKey string, secretKey string) (Account, error) {
	u, _ := url.Parse(TradeApiUrl + "getAccountInfo")
	q := u.Query()
	q.Set("accesskey", accessKey)
	q.Set("method", "getAccountInfo")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Account{}, errors.WithStack(err)
	}

	var assets []Asset
	result, _, _, _ := json.Get(resp.Bytes(), "result")
	json.ArrayEach(result, func(value []byte, dataType json.ValueType, offset int, err error) {
		freezeString, _ := json.GetString(value, "freez")
		freeze, _ := strconv.ParseFloat(freezeString, 64)
		availableString, _ := json.GetString(value, "available")
		available, _ := strconv.ParseFloat(availableString, 64)
		coinCnName, _ := json.GetString(value, "cnName")
		coinEnName, _ := json.GetString(value, "enName")
		coinKey, _ := json.GetString(value, "key")
		coinUnit, _ := json.GetString(value, "unitTag")
		coinScale, _ := json.GetInt(value, "unitDecimal")
		assets = append(assets, Asset{Freeze: freeze, Available: available, Coin: Coin{CnName: coinCnName, EnName: coinEnName, Key: coinKey, Unit: coinUnit, Scale: uint8(coinScale)}})
	}, "coins")

	base, _, _, _ := json.Get(result, "base")
	username, _ := json.GetString(base, "username")
	tradePasswordEnabled, _ := json.GetBoolean(base, "trade_password_enabled")
	authGoogleEnabled, _ := json.GetBoolean(base, "auth_google_enabled")
	authMobileEnabled, _ := json.GetBoolean(base, "auth_mobile_enabled")

	return Account{Username: username, TradePasswordEnabled: tradePasswordEnabled, AuthGoogleEnabled: authGoogleEnabled, AuthMobileEnabled: authMobileEnabled, Assets: assets}, nil
}

func sign(secretKey string, params map[string][]string) string {
	h := hmac.New(md5.New, []byte(fmt.Sprintf("%x", sha1.Sum([]byte(secretKey)))))
	h.Write([]byte(buildQueryString(params)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildQueryString(params map[string][]string) string {
	keys := make([]string, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var kvs []string
	for _, k := range keys {
		for _, v := range params[k] {
			kvs = append(kvs, fmt.Sprintf("%v=%v", k, v))
		}
	}

	return strings.Join(kvs, "&")
}

func extractError(value []byte) error {
	msg, err := json.GetString(value, "error")
	if err == json.KeyPathNotFoundError {
		return nil
	}
	return &ApiError{Code: 1001, Message: msg}
}

type response http.Response

func (r *response) Bytes() ([]byte) {
	bytes, _ := ioutil.ReadAll(r.Body)
	return bytes
}

func (c *RestClient) doGet(url string) (*response, error) {
	resp, err := c.client.Get(url)
	r := response(*resp)
	return &r, err
}
