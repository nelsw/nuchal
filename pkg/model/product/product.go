package product

import "nchl/pkg/util"

type Product struct {
	Id             string `json:"id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`
	StopGain       string `json:"stop_gain"`
	StopLoss       string `json:"stop_loss"`
	Tweezer        string `json:"tweezer"`
	Size           string `json:"size"`
}

func (p Product) stopLoss() float64 {
	return util.Float64(p.StopLoss)
}

func (p Product) stopGain() float64 {
	return util.Float64(p.StopGain)
}
