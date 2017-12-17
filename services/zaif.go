package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewZaif(key, secret string) Service {
	return &zaif{key, secret, time.Now().Unix(), 10}
}

type zaif struct {
	key    string
	secret string
	nonce  int64
	retry  int
}

func (b *zaif) Name() string {
	return "zaif"
}

func (z *zaif) request(method string, params url.Values) ([]byte, error) {
	uri := "https://api.zaif.jp/tapi"
	params.Add("method", method)
	params.Add("nonce", strconv.FormatInt(z.nonce, 10))
	z.nonce++

	encodedParams := params.Encode()
	req, _ := http.NewRequest("POST", uri, strings.NewReader(encodedParams))

	hash := hmac.New(sha512.New, []byte(z.secret))
	hash.Write([]byte(encodedParams))
	signature := hex.EncodeToString(hash.Sum(nil))

	req.Header.Add("Key", z.key)
	req.Header.Add("Sign", signature)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	return byteArray, nil
}

func (z *zaif) GetTradeHistory(currencyPair string, from time.Time, to time.Time) ([]TradeData, error) {
	type Data struct {
		CurrencyPair string `json:"currency_pair"`
		Action       string
		Amount       float64
		Price        float64
		Fee          float64
		FeeAmount    float64 `json:"fee_amount"`
		YourAction   string  `json:"your_action"`
		Bonus        float64
		Timestamp    int64 `json:"timestamp,string"`
		Comment      string
	}
	type Response struct {
		Success int
		Return  map[string]Data
		Error   string `json:"error"`
	}
	// send request
	params := url.Values{}
	params.Add("since", fmt.Sprint(from.Unix()))
	params.Add("currency_pair", currencyPair)

	var response Response
	for retry := 0; retry < z.retry; retry++ {
		resp, err := z.request("trade_history", params)
		if err != nil {
			return nil, err
		}
		// parse response
		if err := json.Unmarshal(resp, &response); err != nil {
			return nil, err
		}
		if response.Success != 1 {
			if retry < z.retry-1 && strings.Contains(response.Error, "time wait restriction") {
				log.Println("Time wait restriction.  Sleeping...")
				time.Sleep(time.Second * 10)
				continue
			}
			return nil, errors.New(response.Error)
		}
		break
	}
	//response.Return
	var data []TradeData
	for _, r := range response.Return {
		profit := 0.
		var action TradeAction
		remarks := []string{}
		if r.YourAction == "bid" {
			profit -= r.Price * r.Amount
			action = TradeActionBuy
		} else if r.YourAction == "ask" {
			profit += r.Price * r.Amount
			action = TradeActionSell
		} else {
			return nil, errors.New("unsupported action: " + r.Action)
		}
		profit -= r.FeeAmount * r.Price
		remarks = append(remarks, "Action: "+r.Action)
		if r.FeeAmount > 0 || r.Fee > 0 {
			remarks = append(remarks, fmt.Sprintf("Fee: %f", r.Fee))
			remarks = append(remarks, fmt.Sprintf("Fee Amount: %f", r.FeeAmount))
		}
		if r.Bonus > 0 {
			remarks = append(remarks, fmt.Sprintf("Bonus: %f", r.Bonus))
			profit += r.Bonus
		}
		tradeData := TradeData{
			ServiceName:  "Zaif",
			CurrencyPair: currencyPair,
			Time:         time.Unix(r.Timestamp, 0),
			Profit:       profit,
			Price:        r.Price,
			Amount:       r.Amount,
			Fee:          r.FeeAmount * r.Price,
			Action:       action,
			Remarks:      remarks,
		}
		data = append(data, tradeData)
	}
	return data, nil
}

func (z *zaif) GetCurrencyPairs() ([]string, error) {
	resp, err := http.Get("https://api.zaif.jp/api/1/currency_pairs/all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type Data struct {
		CurrencyPair string `json:"currency_pair"`
	}
	var data []Data

	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	currencyPairs := []string{}
	for _, d := range data {
		currencyPairs = append(currencyPairs, d.CurrencyPair)
	}
	return currencyPairs, nil
}
