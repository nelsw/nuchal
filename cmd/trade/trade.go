package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"nchl/config"
	"nchl/pkg"
	"os"
)

func main() {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err)
		return
	}

	exit := make(chan string)

	for _, user := range cfg.Users {
		for _, posture := range cfg.Postures {
			go createTrades(user, posture)
		}
	}

	for {
		select {
		case <-exit:
			os.Exit(0)
		}
	}
}

func createTrades(u config.User, p config.Posture) {
	if err := pkg.CreateTrades(u, p); err != nil {
		log.Warn().Err(err).Msg(fmt.Sprintf("error creating [%s] trades for [%s]", p, u))
		createTrades(u, p)
	}
}
