package main

import (
	"github.com/rs/zerolog/log"
	"nchl/pkg/sim"
)

func main() {
	if err := sim.New(); err != nil {
		log.Error().Err(err)
		panic(err)
	}
}
