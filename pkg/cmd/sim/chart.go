/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package sim

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/nelsw/nuchal/pkg/cbp"
	"math"
	"strings"
	"time"
)

// A Chart represents the data used to represent chart activity and trade results.
type Chart struct {

	// Product is an aggregate of the product to trade, and the pattern which used to trade.
	cbp.Product

	// Rates are used to build a chart. The first 3 rates are the tweezer pattern prefix and the last rate is the exit.
	Rates []cbp.Rate

	// Duration is the amount of time the chart spans.
	Duration time.Duration

	// Entry is the actual price which the trade was entered.
	Entry float64

	// Goal is the target price where we place a stop loss order.
	Goal float64

	// Limit is the target price where we sell to avoid taking a bath.
	Limit float64

	// Exit is the actual price which the trade was exited.
	Exit float64

	// MakerFee is a fee for placing a limit order.
	MakerFee float64

	// TakerFee is a fee for placing a market order.
	TakerFee float64
}

// symbol returns an emoji correlated the the status of the chart
func (c *Chart) symbol() string {
	if c.isTrading() {
		return "🌌"
	} else if c.isLoser() {
		return "💩"
	} else if c.isEven() {
		return "👊🏻"
	} else {
		return "💎"
	}
}

func (c *Chart) isWinner() bool {
	return c.Exit != 0 && c.result() > 0
}

func (c *Chart) isLoser() bool {
	return c.Exit != 0 && c.result() < 0
}

func (c *Chart) isTrading() bool {
	return c.Exit == 0
}

func (c *Chart) isEven() bool {
	return c.entryPlusFee() == c.exitPlusFee()
}

func (c *Chart) result() float64 {
	return (c.exitPlusFee() - c.entryPlusFee()) * c.Size
}

func (c *Chart) entryPlusFee() float64 {
	return c.Entry + (c.Entry * c.TakerFee)
}

func (c *Chart) exitPlusFee() float64 {
	return c.Exit + (c.Exit * c.MakerFee)
}

func newChart(makerFee, takerFee float64, rates []cbp.Rate, posture cbp.Product) *Chart {

	c := new(Chart)

	c.Product = posture
	c.MakerFee = makerFee
	c.TakerFee = takerFee

	iterableRates := rates[3:]
	if len(iterableRates) < 1 {
		return c
	}

	c.Entry = iterableRates[0].Open
	c.Goal = posture.GoalPrice(c.Entry)
	c.Limit = posture.LossPrice(c.Entry)

	firstRateTime := iterableRates[0].Time()

	var j int
	var rate cbp.Rate

	for j, rate = range iterableRates {

		// if the low meets or exceeds our loss limit ...
		if rate.Low <= c.Limit {

			// ok, not the worst thing in the world, maybe a stop order already sold this for us
			if c.Exit == 0 {
				// nope, we never established a stop order for this chart, we took a bath
				c.Exit = c.Limit
			}
			break
		}

		// else if the high meets or exceeds our gain limit ...
		if rate.High >= c.Goal {

			// is this the first time this has happened?
			if c.Exit == 0 {
				// great, we have a stop (limit) entry order placed, continue on.
				c.Exit = c.Goal
			}

			// now if the rate closes less than our exit, the entry order would have been triggered.
			if rate.Close < c.Exit {
				break
			}

			// otherwise we're trending up, ride the wave.
			if rate.Close >= c.Exit {
				c.Exit = rate.Close
			}
		}

		// else the low and highs of this rate do not exceed either respective limit
		// we must now navigate each rate and attempt to sell at profit
		// and avoid the ether
		if j == 0 {
			continue
		}

		if c.Exit == 0 && iterableRates[j-1].Time().Sub(firstRateTime) > time.Minute*45 && rate.High >= c.entryPlusFee() {
			c.Exit = c.entryPlusFee()
			break
		}
	}

	f := math.Min(float64(j+4), float64(len(iterableRates)))
	c.Rates = rates[:int(f)]
	return c
}

func (c *Chart) kline() *charts.Kline {

	kline := charts.NewKLine()

	kline.SetGlobalOptions(

		charts.WithTitleOpts(opts.Title{
			Title: c.title(),
		}),

		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 1,
		}),

		charts.WithYAxisOpts(opts.YAxis{
			SplitNumber: 10,
			Scale:       true,
		}),

		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	for _, rate := range c.Rates {
		x = append(x, rate.Label())
		y = append(y, opts.KlineData{Value: rate.Data()})
	}

	kline.SetXAxis(x).
		AddSeries("kline", y).
		SetSeriesOptions(
			charts.WithMarkPointStyleOpts(opts.MarkPointStyle{
				Label: &opts.Label{
					Show: true,
				},
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color0:       "#ec0000",
				Color:        "#00da3c",
				BorderColor0: "#8A0000",
				BorderColor:  "#008F28",
			}),
		)

	return kline
}

func (c *Chart) title() string {

	k := []string{"⊣", "≥", "⊢", "≈", "="}
	for i := 0; i < len(k); i++ {
		k[i] = k[i] + " %s"
	}

	v1 := c.exitPlusFee()
	if c.result() < 0 {
		v1 = c.entryPlusFee()
	}
	f := c.symbol() + "\t" + strings.Join(k, "\t\t\t\t")

	return fmt.Sprintf(f,
		round(c.Entry),
		round(c.Goal),
		round(c.Exit),
		round(v1),
		round(c.result()),
	)
}

func round(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
