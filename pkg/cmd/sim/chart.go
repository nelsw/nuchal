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
	"github.com/nelsw/nuchal/pkg/config"
	"math"
	"strings"
	"time"
)

// A Chart represents the data used to represent chart activity and trade results.
type Chart struct {

	// Rates are used to build a chart. The first 3 rates are the tweezer pattern prefix and the last rate is the exit.
	Rates []cbp.Rate

	// Duration is the amount of time the chart spans.
	Duration time.Duration

	// Entry is the actual price which the trade was entered.
	Entry float64

	// Goal is the target price where we place a stop loss order.
	Goal float64

	// Loss is the target price where we sell to avoid taking a bath.
	Loss float64

	// Exit is the actual price which the trade was exited.
	Exit float64

	// MakerFee is a fee for placing a limit order.
	MakerFee float64

	// TakerFee is a fee for placing a market order.
	TakerFee float64

	Last float64

	SellIndex float64
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
	return c.exitPlusFee() - c.entryPlusFee()
}

func (c *Chart) entryPlusFee() float64 {
	return c.Entry + (c.Entry * c.TakerFee)
}

func (c *Chart) exitPlusFee() float64 {
	exit := c.Exit
	if exit == 0 {
		exit = c.Last
	}
	return exit + (exit * c.MakerFee)
}

func newChart(session *config.Session, rates []cbp.Rate, productID string) *Chart {

	iterableRates := rates[3:]
	if len(iterableRates) < 1 {
		return nil
	}

	c := new(Chart)
	c.MakerFee = cbp.Maker()
	c.TakerFee = cbp.Taker()
	c.Entry = iterableRates[0].Open
	c.Goal = session.GetPattern(productID).GoalPrice(c.Entry)
	c.Loss = session.GetPattern(productID).LossPrice(c.Entry)

	var j int
	var rate cbp.Rate

	for j, rate = range iterableRates {

		// if the low meets or exceeds our loss limit ...
		if rate.Low <= c.Loss {

			// ok, not the worst thing in the world, maybe a stop order already sold this for us
			if c.Exit == 0 {
				// nope, we never established a stop order for this chart, we took a bath
				c.Exit = c.Loss
			}
			c.SellIndex = math.Min(float64(j+4), float64(len(iterableRates)))
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
				c.SellIndex = math.Min(float64(j+3), float64(len(iterableRates)))
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

		if c.Exit == 0 && rate.Time().Sub(iterableRates[0].Time()) > time.Minute*75 && rate.High >= c.entryPlusFee() {
			c.Exit = c.entryPlusFee()
			c.SellIndex = math.Min(float64(j+4), float64(len(iterableRates)))
			break
		}
	}

	c.Last = rate.Close

	c.Rates = rates[:int(c.SellIndex)]
	return c
}

func (c *Chart) kline() *charts.Kline {

	kline := charts.NewKLine()

	k := []string{"IN: ", "GOAL: ", "OUT: ", "NET: "}
	for i := 0; i < len(k); i++ {
		k[i] = k[i] + " %s"
	}

	title := "\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t" + fmt.Sprintf(strings.Join(k, "\t\t\t\t"),
		fmt.Sprintf("%.3f", c.Entry),
		fmt.Sprintf("%.3f", c.Goal),
		fmt.Sprintf("%.3f", c.Exit),
		fmt.Sprintf("%.3f", c.result()),
	)

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithXAxisOpts(opts.XAxis{SplitNumber: 1}),
		charts.WithYAxisOpts(opts.YAxis{SplitNumber: 10, Scale: true}),
		charts.WithDataZoomOpts(opts.DataZoom{Start: 0, End: 100, XAxisIndex: []int{0}}),
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
			charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: true}}),
			charts.WithItemStyleOpts(opts.ItemStyle{"#00da3c", "#ec0000", "#008F28", "#8A0000", 1}),
		)
	return kline
}
