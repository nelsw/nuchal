package main

import (
	"flag"
	"nchl/pkg/conf"
	"nchl/pkg/simulater"
	"nchl/pkg/trade"
	"nchl/pkg/user"
	"os"
	"time"
)

func main() {

	username := flag.String("username", "Connor", "a user first or full username")
	domain := flag.String("domain", "trade", "a program domain to execute")
	duration := flag.String("duration", "5h30m40s", "the duration to execute the domain function")

	flag.Parse()

	cfg := conf.NewDefaultConfig()

	u := cfg.FindUserByFirstName(*username)

	if *domain == "user" {
		user.DisplayAccountInfo(u)
		return
	}

	dur, err := time.ParseDuration(*duration)
	if err != nil {
		// we must want to simulate everything ...
	}

	p := cfg.SimulationProduct()

	if *domain == "sim" {
		from := time.Now().Add(-dur)
		simulater.NewSimulation(u, &from, p)
	}

	if *domain == "trade" {
		exit := make(chan string)
		for range cfg.TradeProductIds {
			go createTrades(u, p)
		}
		for {
			select {
			case <-exit:
				os.Exit(0)
			}
		}
	}

	panic("domain not recognized")
}

func createTrades(u conf.User, p conf.Product) {
	if err := trade.CreateTrades(u, p); err != nil {
		createTrades(u, p)
	}
}
