package main

import (
	"github.com/rs/zerolog/log"
	"nchl/config"
	"nchl/pkg"
)

func main() {
	if cfg, err := config.NewConfig(); err != nil {
		log.Error().Err(err)
	} else {
		pkg.DisplayAccountInfo(cfg)
	}
}
