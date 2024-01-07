package repository

import (
	"context"
	"fernandoglatz/url-management/internal/core/common/utils"
	"fernandoglatz/url-management/internal/core/common/utils/constants"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/entity"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RedirectRepository struct {
	collection *mongo.Collection
}

func NewRedirectRepository() *RedirectRepository {
	return &RedirectRepository{
		collection: utils.MongoDatabase.GetCollection("redirect"),
	}
}

func (repository *RedirectRepository) Get(ctx context.Context, id string) (entity.Redirect, *exceptions.WrappedError) {
	filter := bson.M{"id": id}
	return repository.getByFilter(ctx, filter)
}

func (repository *RedirectRepository) GetByDNS(ctx context.Context, dns string) (entity.Redirect, *exceptions.WrappedError) {
	filter := bson.M{"dns": dns}
	return repository.getByFilter(ctx, filter)
}

func (repository *RedirectRepository) getByFilter(ctx context.Context, filter interface{}) (entity.Redirect, *exceptions.WrappedError) {
	var redirect entity.Redirect

	err := repository.collection.FindOne(ctx, filter).Decode(&redirect)
	if err == mongo.ErrNoDocuments {
		return redirect, &exceptions.WrappedError{
			BaseError: exceptions.RecordNotFound,
		}
	} else if err != nil {
		return redirect, &exceptions.WrappedError{
			Error: err,
		}
	}

	repository.correctTimezone(&redirect)
	return redirect, nil
}

func (repository *RedirectRepository) GetAll(ctx context.Context) ([]entity.Redirect, *exceptions.WrappedError) {
	var redirects []entity.Redirect = []entity.Redirect{}

	cursor, err := repository.collection.Find(ctx, bson.D{})
	if err != nil {
		return redirects, &exceptions.WrappedError{
			Error: err,
		}
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var redirect entity.Redirect
		err = cursor.Decode(&redirect)
		if err != nil {
			return redirects, &exceptions.WrappedError{
				Error: err,
			}
		}

		repository.correctTimezone(&redirect)
		redirects = append(redirects, redirect)
	}

	return redirects, nil
}

func (repository *RedirectRepository) Save(ctx context.Context, redirect *entity.Redirect) *exceptions.WrappedError {
	now := time.Now()
	redirect.UpdatedAt = now

	if len(redirect.ID) == constants.ZERO {
		uuidObj, _ := uuid.NewRandom()
		uuidStr := uuidObj.String()
		redirect.ID = strings.Replace(uuidStr, "-", "", -1)
	}

	if redirect.CreatedAt.IsZero() {
		redirect.CreatedAt = now

		_, err := repository.collection.InsertOne(ctx, redirect)
		if err != nil {
			return &exceptions.WrappedError{
				Error: err,
			}
		}

	} else {
		filter := bson.M{"id": redirect.ID}
		_, err := repository.collection.ReplaceOne(ctx, filter, redirect)
		if err != nil {
			return &exceptions.WrappedError{
				Error: err,
			}
		}
	}

	return nil
}

func (repository *RedirectRepository) Remove(ctx context.Context, redirect entity.Redirect) *exceptions.WrappedError {
	filter := bson.M{"id": redirect.ID}
	_, err := repository.collection.DeleteOne(ctx, filter)
	if err != nil {
		return &exceptions.WrappedError{
			Error: err,
		}
	}

	return nil
}

func (repository *RedirectRepository) correctTimezone(redirect *entity.Redirect) {
	location, _ := time.LoadLocation(utils.GetTimezone())
	redirect.CreatedAt = redirect.CreatedAt.In(location)
	redirect.UpdatedAt = redirect.UpdatedAt.In(location)
}
