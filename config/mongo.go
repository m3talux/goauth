package config

import (
	"errors"

	"github.com/Netflix/go-env"
	"github.com/rs/zerolog/log"
)

var mongoEnvs mongoDB

type mongoDB struct {
	Host           string `env:"MONGODB_HOST,required=true"`
	Port           uint   `env:"MONGODB_PORT,required=true"`
	Name           string `env:"MONGODB_NAME,default=goauth"`
	UseAtlas       bool   `env:"MONGODB_USE_ATLAS,default=false"`
	UseCompression bool   `env:"MONGODB_USE_COMPRESSION,default=false"`
}

func initMongoVariables() {
	_, err := env.UnmarshalFromEnviron(&mongoEnvs)
	if err != nil {
		log.Err(err).Msg("Could not load MongoDB environment variables")
	}
}

func checkMongoEnvs() []error {
	errs := make([]error, 0)

	if mongoEnvs.Host == "" {
		details := "the MongoDB host is not set"
		errs = append(errs, errors.New(details))
	}

	if mongoEnvs.Port == 0 {
		details := "the MongoDB port number is not set"
		errs = append(errs, errors.New(details))
	}

	return errs
}

func MongoDBHost() string {
	return mongoEnvs.Host
}

func MongoDBPort() uint {
	return mongoEnvs.Port
}

func MongoDBName() string {
	return mongoEnvs.Name
}

func MongoDBUseAtlas() bool {
	return mongoEnvs.UseAtlas
}

func MongoDBUseCompression() bool {
	return mongoEnvs.UseCompression
}
