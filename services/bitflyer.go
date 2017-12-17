package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewBitFlyer creates new bitFlyer service.
func NewBitFlyer(key, secret string) Service {
	return &bitFlyer{&http.Client{}, key, secret}
}

type bitFlyer struct {
	client *http.Client
	key    string
	secret string
}

func (b *bitFlyer) Name() string {
	return "bitflyer"
}

func (b *bitFlyer) get(path string, params url.Values, result interface{}) error {
	method := "GET"
	timestamp := time.Now().Unix()
	if params != nil {
		path = path + "?" + params.Encode()
	}
	text := fmt.Sprintf("%d%s%s", timestamp, method, path)
	hash := hmac.New(sha256.New, []byte(b.secret))
	hash.Write([]byte(text))
	signature := hex.EncodeToString(hash.Sum(nil))
	urlstring := "http://api.bitflyer.jp" + path
	req, err := http.NewRequest(method, urlstring, nil)
	if err != nil {
		return err
	}
	req.Header.Add("ACCESS-KEY", b.key)
	req.Header.Add("ACCESS-TIMESTAMP", fmt.Sprint(timestamp))
	req.Header.Add("ACCESS-SIGN", signature)
	req.Header.Add("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(respBody, result); err != nil {
		return err
	}
	return nil
}

func (b *bitFlyer) GetTradeHistory(currencyPair string, from time.Time, to time.Time) ([]TradeData, error) {
	type Execution struct {
		ID                     int64
		Side                   string
		Price                  float64
		Size                   float64
		ExecDate               string `json:"exec_date"`
		OrderID                string `json:"order_id"`
		Commission             float64
		ChildOrderAcceptanceID string
	}
	executions := []Execution{}
	b.get("/v1/me/getexecutions", url.Values{
		"product_code": []string{currencyPair},
		"count":        []string{"10000"},
	}, &executions)
	result := []TradeData{}
	for _, execution := range executions {
		profit := 0.
		var action TradeAction
		if execution.Side == "BUY" {
			profit -= execution.Price * execution.Size
			action = TradeActionBuy
		} else if execution.Side == "SELL" {
			profit += execution.Price * execution.Size
			action = TradeActionSell
		} else {
			return nil, errors.New("unsupported side: " + execution.Side)
		}
		fee := execution.Commission * execution.Price
		profit -= fee
		for i := len(execution.ExecDate); i < 23; i++ {
			execution.ExecDate += "0"
		}
		//fmt.Println(execution.ExecDate)
		if strings.HasSuffix(execution.ExecDate, "0000") {
			execution.ExecDate = execution.ExecDate[0:19] + ".000"
		}
		t, err := time.Parse("2006-01-02T15:04:05.000", execution.ExecDate)
		if err != nil {
			return nil, err
		}
		t = t.Add(time.Hour * 9)
		result = append(result, TradeData{
			ServiceName:  "bitFlyer",
			CurrencyPair: currencyPair,
			Time:         t,
			Profit:       profit,
			Price:        execution.Price,
			Amount:       execution.Size,
			Fee:          fee,
			Action:       action,
		})
	}
	return result, nil
}

func (b *bitFlyer) GetCurrencyPairs() ([]string, error) {
	resp, err := http.Get("https://api.bitflyer.jp/v1/markets")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Market struct {
		ProductCode string `json:"product_code"`
	}
	var markets []Market
	if err := json.Unmarshal(body, &markets); err != nil {
		return nil, err
	}
	currencyPairs := []string{}
	for _, m := range markets {
		currencyPairs = append(currencyPairs, m.ProductCode)
	}
	return currencyPairs, nil
}
