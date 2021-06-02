package cbp

import (
	"github.com/rs/zerolog/log"
	"math"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {

	// Id is concatenation of two currencies. eg. BTC-USD
	Id string `yaml:"id"`

	// Gain is a percentage used to produce the goal sell price from the entry buy price.
	Gain float64 `yaml:"gain"`

	// Loss is a percentage used to derive a limit sell price from the entry buy price.
	Loss float64 `yaml:"loss"`

	// Size is the amount of the transaction, using the products native quote increment.
	Size float64 `yaml:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `yaml:"delta"`
}

func (p *Pattern) GoalPrice(price float64) float64 {
	return price + (price * p.Gain)
}

func (p *Pattern) LossPrice(price float64) float64 {
	return price - (price * p.Loss)
}

func (p *Pattern) MatchesTweezerBottomPattern(then, that, this Rate) bool {
	return isTweezerBottomTrend(then, that, this) && isTweezerBottomValue(that, this, p.Delta)
}

func isTweezerBottomValue(u, v Rate, d float64) bool {
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= d
	if b {
		log.Info().Str("product", v.ProductId).Float64("tweezer", d-f)
	}
	return b
}

func isTweezerBottomTrend(t, u, v Rate) bool {
	return t.IsInit() && u.IsInit() && t.IsDown() && u.IsDown() && v.IsUp()
}
