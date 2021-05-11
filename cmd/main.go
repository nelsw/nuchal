package main

import (
	"flag"
	"nchl/pkg"
	"os"
	"time"
)

func main() {

	username := flag.String("username", "Connor", "a user first or full username")
	domain := flag.String("domain", "trade", "a program domain to execute")
	duration := flag.String("duration", "5h30m40s", "the duration to execute the domain function")

	flag.Parse()

	cfg := pkg.NewDefaultConfig()

	u := cfg.FindUserByFirstName(*username)

	if *domain == "user" {
		pkg.DisplayAccountInfo(u)
		return
	}

	dur, err := time.ParseDuration(*duration)
	if err != nil {
		// we must want to simulate everything ...
	}

	p := cfg.SimulationProduct()

	if *domain == "sim" {
		from := time.Now().Add(-dur)
		pkg.NewSimulation(u, &from, p)
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

func createTrades(u pkg.User, p pkg.Product) {
	if err := pkg.CreateTrades(u, p); err != nil {
		createTrades(u, p)
	}
}
