package mongo

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CrudDAO[T Document] interface {
	// GetCollection returns the collection associated to the document.
	// This method is useful to execute custom mongodb operations that are
	// not covered by the interface.
	GetCollection() *mongo.Collection

	// CreateIndexes launches the creation of defined indexes, for the
	// associated collection.
	CreateIndexes(ctx context.Context, indexes []mongo.IndexModel)

	// Create launches the basic mongodb creation processes, but with a few
	// personalised touches:
	// - If the document contains a "CreatedAt" field, it will be set,
	// - If the creation throws a unique constraint error, the method returns false and no error.
	Create(ctx context.Context, t *T) (bool, error)

	// Update launches a basic flexible mongodb update, but with a few
	// personalised touches:
	// - If the document contains a "UpdatedAt" field, it will be set,
	// - If the filter matches no document and withUpsert is set to false, or no document was updated, it will return an UpdateResult with NotFound set to true.
	// - If the filter matches a document but raises a unique exception, it will return an UpdateResult with UniqueError set to true.
	// - If the filter matches no document and withUpsert is set to true, the document will be created instead.
	Update(ctx context.Context, filter bson.M, update bson.M, withUpsert bool) (UpdateResult, error)

	// Exists launches a basic count mongo request and returns true if a document was found.
	Exists(ctx context.Context, filter bson.M, opts *options.CountOptions) (bool, error)

	// Count counts the number of documents, given a filter.
	Count(ctx context.Context, filter bson.M) int64

	// FindOne searches for a single document in the associated collection, but with a few
	// personalised touches:
	// - A projection and a sort can be set,
	// - If the filter matches no document, the method return nil and no error.
	FindOne(ctx context.Context, filter bson.M, opts *options.FindOneOptions) (*T, error)

	// FindMany searches for multiple documents in the associated collection, but with a few
	// personalised touches:
	// - A projection and a sort can be set,
	// - The decode process is improved with our concurrentDecode algorithm.
	FindMany(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]T, error)

	// Aggregate launches an aggregation pipeline, but with a few
	// personalised touches:
	// - The decode process is improved with our concurrentDecode algorithm.
	Aggregate(ctx context.Context, pipeline interface{}) ([]T, error)

	// Delete deletes a single document in the associated collection, but with a few
	// personalised touches:
	// - If the filter matches no document or no document was updated, the method return false and no error.
	Delete(ctx context.Context, filter bson.M) (bool, error)

	// DeleteMany deletes multiple documents in the associated collection, given a filter.
	DeleteMany(ctx context.Context, filter bson.M) (int64, error)
}

type crudDAO[T Document] struct {
	collection *mongo.Collection
	modelRef   T
}

func (dao *crudDAO[T]) GetCollection() *mongo.Collection {
	return dao.collection
}

func (dao *crudDAO[T]) CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) {
	names, err := dao.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Err(err).Msgf("Could not create indexes for %s model", dao.modelRef.NameSingular())

		return
	}

	log.Info().Strs("fields", names).Msgf("Successfully created indexes for %s model", dao.modelRef.NameSingular())
}

func (dao *crudDAO[T]) Create(ctx context.Context, t *T) (bool, error) {
	_, err := dao.collection.InsertOne(ctx, t)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Err(err).Msgf("Could not create %s due to a unique constraint error", dao.modelRef.NameSingular())

			return false, nil
		}

		log.Error().Fields(map[string]interface{}{
			dao.modelRef.NameSingular(): t,
			"error":                     err,
		}).Msgf("Could not create %s", dao.modelRef.NameSingular())

		return false, err
	}

	log.Debug().Msgf("Successfully created %s", dao.modelRef.NameSingular())

	return true, err
}

func (dao *crudDAO[T]) Update(ctx context.Context, filter bson.M, update bson.M, withUpsert bool) (UpdateResult, error) {
	opts := options.Update().SetUpsert(withUpsert)

	ur, err := dao.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Err(err).Msgf("Could not update %s due to a unique constraint error", dao.modelRef.NameSingular())

			return UpdateResult{
				UniqueError: true,
			}, nil
		}

		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"error":  err,
		}).Msgf("Could not update %s", dao.modelRef.NameSingular())

		return UpdateResult{}, err
	}

	if !withUpsert && ur.MatchedCount == 0 {
		log.Warn().Interface("filter", filter).Msgf("Trying to update a non-existent %s", dao.modelRef.NameSingular())

		return UpdateResult{
			NotFound: true,
		}, nil
	}

	log.Debug().Interface("filter", filter).Msgf("Successfully updated %s", dao.modelRef.NameSingular())

	return UpdateResult{
		Inserted: ur.UpsertedCount == 1,
	}, nil
}

func (dao *crudDAO[T]) Exists(ctx context.Context, filter bson.M, opts *options.CountOptions) (bool, error) {
	count, err := dao.collection.CountDocuments(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn().Interface("filter", filter).Msgf("No %s with given filter exists", dao.modelRef.NameSingular())

			return false, nil
		}

		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not check that %s exists", dao.modelRef.NameSingular())

		return false, err
	}

	return count > 0, nil
}

func (dao *crudDAO[T]) Count(ctx context.Context, filter bson.M) int64 {
	count, err := dao.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not count %s", dao.modelRef.NamePlural())

		return -1
	}

	return count
}

func (dao *crudDAO[T]) FindOne(ctx context.Context, filter bson.M, opts *options.FindOneOptions) (*T, error) {
	sr := dao.collection.FindOne(ctx, filter, opts)
	if err := sr.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn().Interface("filter", filter).Msgf("No %s was found", dao.modelRef.NameSingular())

			//nolint:nilnil // We have to return <nil,nil> here
			return nil, nil
		}

		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not find %s", dao.modelRef.NameSingular())

		return nil, err
	}

	res := new(T)

	if err := sr.Decode(res); err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not decode %s", dao.modelRef.NameSingular())

		return nil, err
	}

	log.Debug().Interface("filter", filter).Msgf("Successfully fetched %s", dao.modelRef.NameSingular())

	return res, nil
}

func (dao *crudDAO[T]) FindMany(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]T, error) {
	cur, err := dao.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not find %s", dao.modelRef.NamePlural())

		return nil, err
	}

	res, err := dao.concurrentDecode(ctx, cur)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not decode %s", dao.modelRef.NamePlural())

		return nil, err
	}

	if err = cur.Err(); err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Cursor error while decoding %s", dao.modelRef.NamePlural())

		return nil, err
	}

	if err = cur.Close(ctx); err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Cursor could not be closed after decoding %s", dao.modelRef.NamePlural())

		return nil, err
	}

	log.Debug().Interface("filter", filter).Msgf("Successfully fetched %s", dao.modelRef.NamePlural())

	return res, nil
}

func (dao *crudDAO[T]) Aggregate(ctx context.Context, pipeline interface{}) ([]T, error) {
	cur, err := dao.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"pipeline": pipeline,
			"err":      err,
		}).Msgf("Aggregation failed for %s", dao.modelRef.NamePlural())

		return nil, err
	}

	res, err := dao.concurrentDecode(ctx, cur)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"pipeline": pipeline,
			"err":      err,
		}).Msgf("Could not decode %s", dao.modelRef.NamePlural())

		return nil, err
	}

	if err = cur.Err(); err != nil {
		log.Error().Fields(map[string]interface{}{
			"pipeline": pipeline,
			"err":      err,
		}).Msgf("Cursor error while decoding %s", dao.modelRef.NamePlural())

		return nil, err
	}

	if err = cur.Close(ctx); err != nil {
		log.Error().Fields(map[string]interface{}{
			"pipeline": pipeline,
			"err":      err,
		}).Msgf("Cursor could not be closed after decoding %s", dao.modelRef.NamePlural())

		return nil, err
	}

	log.Debug().Interface("pipeline", pipeline).Msgf("Aggregatio successful for %s", dao.modelRef.NamePlural())

	return res, nil
}

func (dao *crudDAO[T]) Delete(ctx context.Context, filter bson.M) (bool, error) {
	dr, err := dao.collection.DeleteOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn().Interface("filter", filter).Msgf("Could not delete non-existent %s", dao.modelRef.NameSingular())

			return false, nil
		}

		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not delete %s", dao.modelRef.NameSingular())

		return false, err
	}

	if dr.DeletedCount == 0 {
		log.Warn().Interface("filter", filter).Msgf("No %s was deleted", dao.modelRef.NameSingular())

		return false, nil
	}

	log.Debug().Interface("filter", filter).Msgf("Successfully deleted %s", dao.modelRef.NameSingular())

	return true, nil
}

func (dao *crudDAO[T]) DeleteMany(ctx context.Context, filter bson.M) (int64, error) {
	dr, err := dao.collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"filter": filter,
			"err":    err,
		}).Msgf("Could not delete %s", dao.modelRef.NamePlural())

		return 0, err
	}

	return dr.DeletedCount, nil
}

func (dao *crudDAO[T]) concurrentDecode(ctx context.Context, cur *mongo.Cursor) ([]T, error) {
	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
		err   error
	)

	i := -1
	indexedRes := make(map[int]T)

	for cur.Next(ctx) {
		if err != nil {
			break
		}

		wg.Add(1)

		copyCur := *cur
		i++

		go func(cur mongo.Cursor, i int) {
			defer wg.Done()

			r := new(T)

			decodeError := cur.Decode(r)
			if decodeError != nil {
				if err == nil {
					err = decodeError
				}

				return
			}

			mutex.Lock()
			indexedRes[i] = *r
			mutex.Unlock()
		}(copyCur, i)
	}

	wg.Wait()

	if err != nil {
		return nil, err
	}

	resLen := len(indexedRes)

	res := make([]T, resLen)

	for j := 0; j < resLen; j++ {
		res[j] = indexedRes[j]
	}

	return res, nil
}

func NewCrudDAO[T Document](db *mongo.Database) CrudDAO[T] {
	dao := &crudDAO[T]{}

	dao.collection = db.Collection(dao.modelRef.CollectionName())

	if len(dao.modelRef.Indexes()) > 0 {
		// Here we pass a background context because this operation takes time
		go dao.CreateIndexes(context.Background(), dao.modelRef.Indexes())
	}

	return dao
}
