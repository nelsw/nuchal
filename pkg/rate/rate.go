package rate

import (
	"fmt"
	"math"
	"time"
)

type Candlestick struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductId string  `json:"product" gorm:"primaryKey"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (v *Candlestick) IsDown() bool {
	return v.Open > v.Close
}

func (v *Candlestick) IsUp() bool {
	return !v.IsDown()
}

func (v *Candlestick) IsInit() bool {
	return v != nil && v != (&Candlestick{})
}

func (v *Candlestick) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Candlestick) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}

func IsTweezer(t, u, v Candlestick) bool {
	return isTweezerPattern(t, u, v) && isTweezerValue(t, u)
}

func tweezer(v Candlestick) float64 {
	if v.Close < 1 {
		return .001
	} else if v.Close < 10 {
		return .01
	} else if v.Close < 100 {
		return .1
	} else {
		return 1
	}
}

func isTweezerValue(u, v Candlestick) bool {
	t := tweezer(v)
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= t
	s := ">"
	if b {
		s = "<="
	}
	fmt.Printf("... tweezer value? [%f] %s [%f] [%v]\n:", f, s, t, b)
	return b
}

func isTweezerPattern(t, u, v Candlestick) bool { //@f0
	isTweezerPattern :=
		t.IsInit() &&
			u.IsInit() &&
			t.IsDown() &&
			u.IsDown() &&
			v.IsUp()
	fmt.Println("... is tweezer pattern?", isTweezerPattern)
	return isTweezerPattern
} //@f1
