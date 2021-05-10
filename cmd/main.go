package main

import (
	"flag"
	. "nchl/pkg"
	"regexp"
	"strings"
)

var domains = regexp.MustCompile(`trade|sim|user`)

func main() {

	domain := flag.String("d", "trade", "a program domain to execute")
	symbol := flag.String("s", "btc", "a crypto product symbol")
	username := flag.String("u", "Connor", "a users first or full username")
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

	if symbol == nil {
		panic("symbol cannot be nil yeah dingus")
	}

	productId := strings.ToUpper(*symbol) + "-USD"
	if *domain == "sim" {
		ServeCharts(NewSimulation(*username, productId))
		return
	}

	// *domain == trade
	CreateTrades(*username, productId)

}
