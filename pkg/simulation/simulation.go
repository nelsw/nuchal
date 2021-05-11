package simulation

import (
	"fmt"
	"nchl/pkg/coinbase"
	"nchl/pkg/db"
	"nchl/pkg/model/fee"
	"nchl/pkg/model/product"
	"nchl/pkg/model/rate"
	"nchl/pkg/model/report"
	"nchl/pkg/util"
	"time"
)

const (
	asc     = "unix asc"
	desc    = "unix desc"
	query   = "product_id = ?"
	timeVal = "2021-04-01T00:00:00+00:00"
)

func GetRecentRates(name, productId string) []rate.Candlestick {
	fmt.Println("finding recent rates")

	var r rate.Candlestick
	db.Client.Where(query, productId).Order(desc).First(&r)

	from := time.Now().AddDate(0, 0, -1)
	db.Client.Save(coinbase.CreateHistoricRates(name, productId, from))

	var allRates []rate.Candlestick

	db.Client.Where("product_id = ?", productId).
		Where("unix > ?", from.UnixNano()).
		Order(asc).
		Find(&allRates)

	fmt.Println("found recent rates", len(allRates))

	return allRates
}

func GetRates(name, productId string) []rate.Candlestick {

	fmt.Println("finding rates")

	var r rate.Candlestick
	db.Client.Where(query, productId).Order(desc).First(&r)

	var from time.Time
	if r != (rate.Candlestick{}) {
		from = r.Time()
	} else {
		from, _ = time.Parse(time.RFC3339, timeVal)
	}

	chunks := coinbase.CreateHistoricRates(name, productId, from)
	fmt.Println(util.Pretty(chunks))
	db.Client.Save(chunks)

	var allRates []rate.Candlestick
	db.Client.Where(query, productId).Order(asc).Find(&allRates)

	fmt.Println("found rates", len(allRates))

	return allRates
}

func NewRecentSimulation(name, productId string) report.Result {
	fmt.Println("creating recent simulation")
	s := newSimulation(GetRecentRates(name, productId), productId)
	fmt.Println("created recent simulation")
	return s
}

func NewSimulation(name, productId string) report.Result {
	fmt.Println("creating simulation")
	s := newSimulation(GetRates(name, productId), productId)
	fmt.Println("crated simulation")
	return s
}

func newSimulation(rates []rate.Candlestick, productId string) report.Result {

	var positionIndexes []int
	var then, that rate.Candlestick

	for i, this := range rates {
		if rate.IsTweezer(then, that, this) {
			positionIndexes = append(positionIndexes, i)
		}
		then = that
		that = this
	}

	var won, lost, vol float64
	var scenarios []report.Scenario

	for _, i := range positionIndexes {

		alpha := i - 2

		var entry, exit, result float64
		market := rates[i].Open

		gain := product.PricePlusStopGain(productId, market)
		loss := product.PriceMinusStopLoss(productId, market)

		for j, r := range rates[i:] {

			if r.High >= gain {
				entry = gain
				if r.Low <= entry {
					exit = entry
				} else if r.Close >= exit {
					exit = r.Close
					continue
				}
				result = exit - market - (market * fee.Base) - (exit * fee.Base)
			} else if r.Low <= loss {
				result = loss - market - (market * fee.Base) - (loss * fee.Base)
			} else {
				continue
			}

			result *= util.Float64(product.Size(market))
			if result > 0 {
				won += result
			} else {
				lost += result
			}

			vol += entry

			scenarios = append(scenarios, report.Scenario{
				r.Time(),
				rates[alpha : i+j+2],
				market,
				entry,
				exit,
				result,
			})
			break
		}
	}

	simulation := report.Result{
		won,
		lost,
		vol,
		scenarios,
		productId,
		rates[0].Time(),
		that.Time(),
	}

	fmt.Println()
	fmt.Println("productId", simulation.ProductId)
	fmt.Println("     from", simulation.From)
	fmt.Println("       to", simulation.To)
	fmt.Println("scenarios", len(simulation.Scenarios))
	fmt.Println("      won", simulation.Won)
	fmt.Println("     lost", simulation.Lost)
	fmt.Println("   report", simulation.Sum())
	fmt.Println("   volume", simulation.Vol)
	fmt.Println("   return", simulation.Result())
	fmt.Println()

	return simulation
}
