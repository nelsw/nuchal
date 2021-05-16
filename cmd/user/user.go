package main

import (
	"github.com/rs/zerolog/log"
	"nchl/pkg/status"
)

func main() {
	if err := status.New(); err != nil {
		log.Error().Err(err)
		panic(err)
	}
}
