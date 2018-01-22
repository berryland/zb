package zb

import (
	json "github.com/buger/jsonparser"
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

func (c *RestClient) GetSymbols() (map[string]SymbolConfig, error) {
	configs := map[string]SymbolConfig{}
	resp, err := c.doGet(DataApiUrl + "markets")
	if err != nil {
		return configs, err
	}

	bytes := resp.ReadBytes()
	err = extractError(bytes)
	if err != nil {
		return configs, err
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

func (c *RestClient) GetLatestQuote(symbol string) (Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", symbol)
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Quote{}, err
	}

	bytes := resp.ReadBytes()
	err = extractError(bytes)
	if err != nil {
		return Quote{}, err
	}

	return marshalQuote(bytes), nil
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
		return klines, err
	}

	bytes := resp.ReadBytes()
	err = extractError(bytes)
	if err != nil {
		return klines, err
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

func (c *RestClient) GetTrades(symbol string, since uint64) ([]Trade, error) {
	var trades []Trade
	u, _ := url.Parse(DataApiUrl + "trades")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("since", strconv.FormatUint(since, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return trades, err
	}

	bytes := resp.ReadBytes()
	err = extractError(bytes)
	if err != nil {
		return trades, err
	}

	json.ArrayEach(bytes, func(value []byte, dataType json.ValueType, offset int, err error) {
		id, _ := json.GetInt(value, "tid")
		tradeType, _ := json.GetString(value, "type")
		amountString, _ := json.GetString(value, "amount")
		priceString, _ := json.GetString(value, "price")
		time, _ := json.GetInt(value, "date")

		amount, _ := strconv.ParseFloat(amountString, 64)
		price, _ := strconv.ParseFloat(priceString, 64)

		trades = append(trades, Trade{Id: uint64(id), TradeType: ParseTradeType(tradeType), Price: price, Amount: amount, Time: uint64(time)})
	})

	return trades, nil
}

func (c *RestClient) GetDepth(symbol string, size uint8) (Depth, error) {
	u, _ := url.Parse(DataApiUrl + "depth")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Depth{}, err
	}

	bytes := resp.ReadBytes()
	err = extractError(bytes)
	if err != nil {
		return Depth{}, err
	}

	time, _ := json.GetInt(bytes, "timestamp")
	asks, bids := marshalDepthEntries(bytes, "asks"), marshalDepthEntries(bytes, "bids")

	return Depth{Asks: asks, Bids: bids, Time: uint64(time)}, nil
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
		return Account{}, err
	}

	bytes := resp.ReadBytes()
	err = extractTradeError(bytes)
	if err != nil {
		return Account{}, err
	}

	var assets []Asset
	result, _, _, _ := json.Get(bytes, "result")
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

func (c *RestClient) PlaceOrder(symbol string, price, amount float64, tradeType TradeType, accessKey, secretKey string) (uint64, error) {
	u, _ := url.Parse(TradeApiUrl + "order")
	q := u.Query()
	q.Set("currency", symbol)
	q.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	q.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	q.Set("tradeType", strconv.FormatUint(uint64(tradeType), 8))
	q.Set("accesskey", accessKey)
	q.Set("method", "order")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return 0, err
	}

	bytes := resp.ReadBytes()
	err = extractTradeError(bytes)
	if err != nil {
		return 0, err
	}

	idString, _ := json.GetString(bytes, "id")
	id, _ := strconv.ParseUint(idString, 10, 64)
	return id, nil
}

func (c *RestClient) CancelOrder(symbol string, id uint64, accessKey, secretKey string) error {
	u, _ := url.Parse(TradeApiUrl + "cancelOrder")
	q := u.Query()
	q.Set("currency", symbol)
	q.Set("id", strconv.FormatUint(id, 10))
	q.Set("accesskey", accessKey)
	q.Set("method", "cancelOrder")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return err
	}

	bytes := resp.ReadBytes()
	err = extractTradeError(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *RestClient) GetOrder(symbol string, id uint64, accessKey, secretKey string) (Order, error) {
	u, _ := url.Parse(TradeApiUrl + "getOrder")
	q := u.Query()
	q.Set("currency", symbol)
	q.Set("id", strconv.FormatUint(id, 10))
	q.Set("accesskey", accessKey)
	q.Set("method", "getOrder")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return Order{}, err
	}

	bytes := resp.ReadBytes()
	err = extractTradeError(bytes)
	if err != nil {
		return Order{}, err
	}

	return parseOrder(bytes), nil
}

func (c *RestClient) GetOrders(symbol string, tradeType TradeType, page uint64, size uint16, accessKey, secretKey string) ([]Order, error) {
	u := getUrlToGetOrders(symbol, tradeType, page, size, accessKey, secretKey)
	resp, err := c.doGet(u.String())
	if err != nil {
		return []Order{}, err
	}

	bytes := resp.ReadBytes()
	err = extractTradeError(bytes)
	if err != nil {
		return []Order{}, err
	}

	var orders []Order
	json.ArrayEach(bytes, func(value []byte, dataType json.ValueType, offset int, err error) {
		orders = append(orders, parseOrder(value))
	})

	return orders, nil
}

func parseOrder(value []byte) Order {
	idString, _ := json.GetString(value, "id")
	id, _ := strconv.ParseUint(idString, 10, 64)
	currency, _ := json.GetString(value, "currency")
	price, _ := json.GetFloat(value, "price")
	status, _ := json.GetInt(value, "status")
	totalAmount, _ := json.GetFloat(value, "total_amount")
	tradeAmount, _ := json.GetFloat(value, "trade_amount")
	tradePrice, _ := json.GetFloat(value, "trade_price")
	tradeMoney, _ := json.GetFloat(value, "trade_money")
	tradeDate, _ := json.GetInt(value, "trade_date")
	tradeType, _ := json.GetInt(value, "type")
	return Order{Id: id, Price: price, Average: tradePrice, TotalAmount: totalAmount, TradeAmount: tradeAmount, TradeMoney: tradeMoney, Symbol: currency, Status: OrderStatus(status), TradeType: TradeType(tradeType), Time: uint64(tradeDate)}
}

func getUrlToGetOrders(symbol string, tradeType TradeType, page uint64, size uint16, accessKey, secretKey string) *url.URL {
	switch tradeType {
	case All:
		return getOrdersIgnoreTradeType(symbol, page, size, accessKey, secretKey)
	case Buy, Sell:
		return getOrdersNew(symbol, tradeType, page, size, accessKey, secretKey)
	default:
		panic("Unknown trade type: " + string(tradeType))
	}
}

func getOrdersIgnoreTradeType(symbol string, page uint64, size uint16, accessKey, secretKey string) *url.URL {
	u, _ := url.Parse(TradeApiUrl + "getOrdersIgnoreTradeType")
	q := u.Query()
	q.Set("currency", symbol)
	q.Set("pageIndex", strconv.FormatUint(page, 10))
	q.Set("pageSize", strconv.FormatUint(uint64(size), 10))
	q.Set("accesskey", accessKey)
	q.Set("method", "getOrdersIgnoreTradeType")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()
	return u
}

func getOrdersNew(symbol string, tradeType TradeType, page uint64, size uint16, accessKey, secretKey string) *url.URL {
	u, _ := url.Parse(TradeApiUrl + "getOrdersNew")
	q := u.Query()
	q.Set("currency", symbol)
	q.Set("tradeType", strconv.FormatUint(uint64(tradeType), 8))
	q.Set("pageIndex", strconv.FormatUint(page, 10))
	q.Set("pageSize", strconv.FormatUint(uint64(size), 10))
	q.Set("accesskey", accessKey)
	q.Set("method", "getOrdersNew")
	q.Set("sign", sign(secretKey, q))
	q.Set("reqTime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()
	return u
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

func extractTradeError(value []byte) error {
	code, err := json.GetInt(value, "code")
	if err == json.KeyPathNotFoundError || code == 1000 {
		return nil
	}
	msg, _ := json.GetString(value, "message")
	return &ApiError{Code: uint16(code), Message: msg}
}

type response http.Response

func (r *response) ReadBytes() ([]byte) {
	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)
	return bytes
}

func (c *RestClient) doGet(url string) (*response, error) {
	resp, err := c.client.Get(url)
	r := response(*resp)
	return &r, err
}