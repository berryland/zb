package zb

import (
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/pkg/errors"
	"net/url"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type ContractConfig struct {
	AmountScale byte
	PriceScale  byte
}

func GetContractConfigs() (*map[string]ContractConfig, error) {
	result := &map[string]ContractConfig{}
	err := getJsonResponse(DataApiUrl+"markets", result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

type quote struct {
	Ticker    ticker
	Timestamp uint64 `json:"date,string"`
}

type ticker struct {
	Volume float32 `json:"vol,string"`
	Last   float32 `json:",string"`
	Sell   float32 `json:",string"`
	Buy    float32 `json:",string"`
	High   float32 `json:",string"`
	Low    float32 `json:",string"`
}

type Quote struct {
	ticker
	Timestamp uint64
}

func GetLatestQuote(contract string) (*Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", contract)
	u.RawQuery = q.Encode()

	qt := &quote{}
	err := getJsonResponse(u.String(), qt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := &Quote{qt.Ticker, qt.Timestamp}
	return result, nil
}

func getJsonResponse(url string, v interface{}) error {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		return errors.WithStack(err)
	}

	err = jsoniter.Unmarshal(resp.Body(), v)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
