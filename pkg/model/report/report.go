package report

import (
	"nchl/pkg/model/rate"
	"time"
)

type Result struct {
	Won, Lost, Vol float64
	Scenarios      []Scenario
	ProductId      string
	From, To       time.Time
}

func (s Result) Sum() float64 {
	return s.Won + s.Lost
}

func (s Result) Result() float64 {
	return s.Sum() / s.Vol
}

type Scenario struct {
	Time                        time.Time
	Rates                       []rate.Candlestick
	Market, Entry, Exit, Result float64
}
