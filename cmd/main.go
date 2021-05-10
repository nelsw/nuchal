package main

import (
	"flag"
	. "nchl/pkg"
	"os"
	"regexp"
	"strings"
)

var domains = regexp.MustCompile(`trade|sim|user`)

func main() {

	domain := flag.String("domain", "trade", "a program domain to execute")
	symbol := flag.String("symbol", "btc", "a crypto product symbol")
	symbols := flag.String("symbols", "btc,eth,xtz", "a csv list of crypto product symbols")
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
		CreateUser(*username, *key, *pass, *secret)
		return
	}

	if *domain == "trades" {
		if symbols == nil {
			panic("must have symbols to process trades")
		}
		exit := make(chan string)
		for _, s := range strings.Split(*symbols, ",") {
			go CreateTrades(*username, productId(&s))
		}
		for {
			select {
			case <-exit:
				os.Exit(0)
			}
		}
	}

	if symbol == nil {
		panic("symbol cannot be nil yeah dingus")
	}

	if *domain == "sim" {
		ServeCharts(NewSimulation(*username, productId(symbol)))
		return
	}

	if *domain == "trade" {
		CreateTrades(*username, productId(symbol))
	}
}

func productId(symbol *string) string {
	return strings.ToUpper(*symbol) + "-USD"
}
