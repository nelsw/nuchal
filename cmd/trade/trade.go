package main

import (
	"github.com/rs/zerolog/log"
	"nchl/pkg/trade"
)

func main() {
	if err := trade.New(); err != nil {
		log.Error().Err(err)
		panic(err)
	}
}
