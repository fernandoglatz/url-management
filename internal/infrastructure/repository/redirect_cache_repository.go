package repository

import (
	"context"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/entity"
	"fernandoglatz/url-management/internal/core/port/repository"
)

const REDIRECT_CACHE_KEY_PREFIX = "url-management:redirect:"

type RedirectCacheRepository struct {
	repository repository.IRedirectRepository
}

func NewRedirectCacheRepository(repository repository.IRedirectRepository) *RedirectCacheRepository {
	return &RedirectCacheRepository{
		repository: repository,
	}
}

func (cacheRepository *RedirectCacheRepository) Get(ctx context.Context, id string) (entity.Redirect, *exceptions.WrappedError) {
	return cacheRepository.repository.Get(ctx, id)
}

func (cacheRepository *RedirectCacheRepository) GetByDNS(ctx context.Context, dns string) (entity.Redirect, *exceptions.WrappedError) {
	return cacheRepository.repository.GetByDNS(ctx, dns)
}

func (cacheRepository *RedirectCacheRepository) GetAll(ctx context.Context) ([]entity.Redirect, *exceptions.WrappedError) {
	return cacheRepository.repository.GetAll(ctx)
}

func (cacheRepository *RedirectCacheRepository) Save(ctx context.Context, redirect *entity.Redirect) *exceptions.WrappedError {
	return cacheRepository.repository.Save(ctx, redirect)
}

func (cacheRepository *RedirectCacheRepository) Remove(ctx context.Context, redirect entity.Redirect) *exceptions.WrappedError {
	return cacheRepository.repository.Remove(ctx, redirect)
}

/*
func (cacheRepository *RedirectCacheRepository) Save(ctx context.Context, redirect *entity.Redirect) error {
	err := cacheRepository.repository.Save(ctx, redirect)

	if err == nil {
		tenant := redirect.Tenant
		userId := redirect.UserId
		cacheKey := REDIRECT_CACHE_KEY_PREFIX + tenant + "-" + userId

		err := utils.RedisDatabase.Del(ctx, cacheKey)
		if err != nil {
			log.Error(ctx).Msg("Error removing redirect from cache: " + err.Error())
		}
	}

	return err
}

func (cacheRepository *RedirectCacheRepository) Get(ctx context.Context, tenant string, userId string) (entity.Redirect, error) {
	var redirect entity.Redirect

	cacheKey := REDIRECT_CACHE_KEY_PREFIX + tenant + "-" + userId
	err := utils.RedisDatabase.GetStruct(ctx, cacheKey, &redirect)

	if err == nil {
		return redirect, nil

	} else if err != redis.Nil {
		log.Error(ctx).Msg("Error retrieving redirect from cache: " + err.Error())
	}

	redirect, err = cacheRepository.repository.Get(ctx, tenant, userId)
	if err == nil {
		expiration := config.ApplicationConfig.Data.Redis.TTL.Redirect
		err = utils.RedisDatabase.SetStruct(ctx, cacheKey, redirect, expiration)
		if err != nil {
			log.Error(ctx).Msg("Error adding redirect to cache: " + err.Error())
		}
	}

	return redirect, err
}
*/
