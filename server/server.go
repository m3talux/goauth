package server

import (
	"context"

	"github.com/m3talux/goauth/config"

	"github.com/m3talux/goauth/handler"
	"github.com/m3talux/goauth/mongo"
	"github.com/m3talux/goauth/router"
	"github.com/rs/zerolog/log"
)

type Server struct{}

func (s *Server) Run() error {
	initializationContext, cancel := context.WithTimeout(context.Background(), config.InitializationTimeout())
	defer cancel()

	// DB layer initialization
	_, err := mongo.DB(initializationContext)
	if err != nil {
		log.Err(err).Msg("Could not create the MongoDB database connector")

		return err
	}

	// DAO layer initialization

	// Manager layer initialization

	// Handler layer initialization
	checkHandler := handler.NewCheckHandler()

	r := router.NewRouter(
		router.Handlers{
			CheckHandler: checkHandler,
		},
	)

	return r.Run()
}

func New() *Server {
	return &Server{}
}
