package main

import (
	"sushkov/internal/app"
	"sushkov/internal/config"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.MustLoad()

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize app")
	}

	a.Run()
}
