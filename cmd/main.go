package main

import (
	"flag"
	"nchl/pkg/chart"
	"nchl/pkg/model/product"
	"nchl/pkg/simulation"
	"nchl/pkg/trade"
	"nchl/pkg/user"
	"os"
)

func main() {

	username := flag.String("username", "Connor", "a user first or full username")
	domain := flag.String("domain", "trade", "a program domain to execute")

	flag.Parse()

	switch *domain {
	case "user":
		user.DisplayAccountInfo(*username)
	case "sim":
		chart.ServeCharts(simulation.NewSimulation(*username, product.IdToSimulateTrade()))
	case "now":
		chart.ServeCharts(simulation.NewRecentSimulation(*username, product.IdToSimulateTrade()))
	case "trades":
		exit := make(chan string)
		for _, productId := range product.IdsToTrade() {
			go trade.CreateTrades(*username, productId)
		}
		for {
			select {
			case <-exit:
				os.Exit(0)
			}
		}
	case "tidy":
		user.CreateEntryOrders(*username)
	default:
		panic("domain not recognized yeah ignoramus")
	}
}
