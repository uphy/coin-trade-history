package services

import (
	"sort"
	"time"
)

// Service represents the cryptocurrency trading service.
type Service interface {
	// GetTradeHistory gets the trade history.
	GetTradeHistory(currencyPair string, from time.Time, to time.Time) ([]TradeData, error)
	// GetCurrencyPairs returns the available currency pairs.
	GetCurrencyPairs() ([]string, error)
	Name() string
}

// TradeAction is the action of the cryptocurency trade.
type TradeAction int

func (t TradeAction) String() string {
	switch t {
	case TradeActionBuy:
		return "Buy"
	case TradeActionSell:
		return "Sell"
	default:
		panic("unexpected")
	}
}

const (
	// TradeActionBuy represents the buy action in the trade.
	TradeActionBuy TradeAction = iota
	// TradeActionSell represents the sell action in the trade.
	TradeActionSell
)

// TradeData represents the trading data in cryptocurrency service.
type TradeData struct {
	// ServiceName is the name of the cryptocurrency service.
	ServiceName string
	// CurrencyPair means the type of the cryptocurrency
	CurrencyPair string
	// Time is the time of the trade.
	Time time.Time
	// Action is the action of the trade; sell or buy.
	Action TradeAction
	// Price is the price of the cryptocurrency.
	Price float64
	// Amount is the amount of the cryptocurrency.
	Amount float64
	// Fee is the fee of the trade.
	Fee float64
	// Profit is the total profit of this trade.
	Profit float64
	// Remarks is the optional field for additional information.
	Remarks []string
}

// SortTradeData sorts the TradeData order by time.
func SortTradeData(data []TradeData) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Time.Unix() < data[j].Time.Unix()
	})
}
