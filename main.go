package main

import (
	"github.com/m3talux/goauth/config"
	"github.com/m3talux/goauth/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config.Initialize()

	s := server.New()

	if err := s.Run(); err != nil {
		log.Err(err).Msgf("Could not start %s, router failed to run", config.AppName())
	}
}
