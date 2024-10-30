package config

import (
	"strings"
	"time"

	"github.com/Netflix/go-env"
	"github.com/rs/zerolog/log"
)

var corsEnvs cors

type cors struct {
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS"`
	MaxAge         int    `env:"CORS_MAX_AGE,default=3600"`
}

func initCORSVariables() {
	_, err := env.UnmarshalFromEnviron(&corsEnvs)
	if err != nil {
		log.Err(err).Msg("Could not load CORS environment variables")
	}
}

func CorsAllowedOrigins() []string {
	corsAllowedOrigins := corsEnvs.AllowedOrigins
	if corsAllowedOrigins == "" {
		return nil
	}

	allowedOriginsJoined := strings.TrimSpace(corsAllowedOrigins)
	allowedOriginsSplit := strings.Split(allowedOriginsJoined, ",")

	allowedOriginsSlice := make([]string, len(allowedOriginsSplit))

	for i, origin := range allowedOriginsSplit {
		trimmed := strings.TrimSpace(origin)
		lowered := strings.ToLower(trimmed)
		allowedOriginsSlice[i] = lowered
	}

	return allowedOriginsSlice
}

func CorsMaxAge() time.Duration {
	return time.Duration(corsEnvs.MaxAge) * time.Second
}
