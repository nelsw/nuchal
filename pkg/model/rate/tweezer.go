package rate

import (
	"fmt"
	"math"
)

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
	} else if v.Close < 1000 {
		return 1
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
}
