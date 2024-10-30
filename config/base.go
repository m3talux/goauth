package config

import (
	"time"

	"github.com/Netflix/go-env"
	"github.com/rs/zerolog/log"
)

const (
	appName = "Goauth"

	apiBasePath    = "/api"
	apiVersionPath = "/v1"

	connectionTimeout     = 10 * time.Second
	initializationTimeout = 60 * time.Second
)

var baseEnvs base

type base struct {
	GinMode string `env:"GIN_MODE"`
}

func initBaseVariables() {
	_, err := env.UnmarshalFromEnviron(&baseEnvs)
	if err != nil {
		log.Err(err).Msg("Could not load base environment variables")
	}
}

func AppName() string {
	return appName
}

func GinMode() string {
	return baseEnvs.GinMode
}

func APIPath() string {
	return apiBasePath + apiVersionPath
}

func ConnectionTimeout() time.Duration {
	return connectionTimeout
}

func InitializationTimeout() time.Duration {
	return initializationTimeout
}
