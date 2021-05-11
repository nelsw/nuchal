package main

import (
	"flag"
	"nchl/pkg/chart"
	"nchl/pkg/product"
	"nchl/pkg/simulation"
	"nchl/pkg/trade"
	"nchl/pkg/user"
	"os"
	"regexp"
	"strings"
)

var domains = regexp.MustCompile(`trade|sim|user|tidy|now`)

func main() {

	domain := flag.String("domain", "trade", "a program domain to execute")
	symbol := flag.String("symbol", "btc", "a crypto product symbol or csv of symbols")
	username := flag.String("username", "Connor", "a users first or full username")
	key := flag.String("key", "example_key", "a Coinbase Pro API key")
	pass := flag.String("pass", "example_pass_phrase", "a Coinbase Pro API passphrase")
	secret := flag.String("secret", "example_secret", "a Coinbase Pro API secret")

	flag.Parse()

	// validate domain value
	if domain == nil {
		panic("domain cannot be nil yeah dummy")
	} else if !domains.MatchString(*domain) {
		panic("domain not recognized yeah ignoramus")
	}

	if *domain == "user" {
		user.CreateUser(*username, *key, *pass, *secret)
		return
	}

	if symbol == nil {
		panic("symbol cannot be nil yeah dingus")
	}

	if *domain == "sim" {
		chart.ServeCharts(simulation.NewSimulation(*username, product.ProductId(symbol)))
		return
	}

	if *domain == "now" {
		chart.ServeCharts(simulation.NewRecentSimulation(*username, product.ProductId(symbol)))
		return
	}

	if *domain == "trades" {
		exit := make(chan string)
		for _, s := range strings.Split(*symbol, ",") {
			go trade.CreateTrades(*username, product.ProductId(&s))
		}
		for {
			select {
			case <-exit:
				os.Exit(0)
			}
		}
	}

	if *domain == "trade" {
		trade.CreateTrades(*username, product.ProductId(symbol))
	}

	if *domain == "tidy" {
		user.CreateEntryOrders(*username)
	}

}
