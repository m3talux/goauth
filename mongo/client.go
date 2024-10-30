package mongo

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/m3talux/goauth/config"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/*
Used to create a singleton object of MongoDB client.
Initialized and exposed through DB().
*/
var clientInstance *mongo.Client

// Used to execute client creation procedure only once.
var singleExecution sync.Once

func initialize(ctx context.Context) error {
	var clientInstanceError error

	singleExecution.Do(func() {
		// Set client options
		host := config.MongoDBHost()
		port := config.MongoDBPort()
		name := config.MongoDBName()

		// FIXME: use uri env instead
		var mongoURI string
		if config.MongoDBUseAtlas() {
			mongoURI = fmt.Sprintf("mongodb+srv://%s/%s?%s", host, name, "retryWrites=true&w=majority")
		} else {
			mongoURI = fmt.Sprintf("mongodb://%s:%d/%s", host, port, name)
		}

		clientOptions := options.Client().ApplyURI(mongoURI)

		// Network compression allows to improve performance when requesting large volume of data.
		if config.MongoDBUseCompression() {
			clientOptions.SetCompressors([]string{"zstd"})
		}

		// Define a context with timeout
		ctxT, cancel := context.WithTimeout(ctx, config.ConnectionTimeout())
		defer cancel()

		// Connect to MongoDB
		client, err := mongo.Connect(ctxT, clientOptions)
		if err != nil {
			clientInstanceError = err

			return
		}

		err = client.Ping(ctxT, nil)
		if err != nil {
			clientInstanceError = err

			return
		}

		log.Info().Str("mongoURI", uriForLog(mongoURI)).Msg("The MongoDB client has been initialized")

		clientInstance = client
	})

	return clientInstanceError
}

func DB(ctx context.Context) (*mongo.Database, error) {
	if err := initialize(ctx); err != nil {
		return nil, err
	}

	if clientInstance == nil {
		return nil, errors.New("mongo client was not initialized")
	}

	return clientInstance.Database(config.MongoDBName()), nil
}

func Check() []error {
	errs := make([]error, 0)

	if clientInstance == nil {
		errs = append(errs, errors.New("the mongo client is nil"))

		return errs
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout())
	defer cancel()

	if err := clientInstance.Ping(ctx, readpref.Primary()); err != nil {
		details := fmt.Sprintf("the mongo check has failed: %s", err)
		errs = append(errs, errors.New(details))
	}

	return errs
}
