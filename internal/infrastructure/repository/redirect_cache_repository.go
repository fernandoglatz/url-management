package repository

import (
	"context"
	"fernandoglatz/url-management/internal/core/common/utils"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/common/utils/log"
	"fernandoglatz/url-management/internal/core/entity"
	"fernandoglatz/url-management/internal/core/port/repository"
	"fernandoglatz/url-management/internal/infrastructure/config"
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
	cacheKey := REDIRECT_CACHE_KEY_PREFIX + id
	var redirect entity.Redirect
	cacheErr := utils.RedisDatabase.GetStruct(ctx, cacheKey, &redirect)
	if cacheErr != nil {
		log.Error(ctx).Msg("Error retrieving redirect from cache: " + cacheErr.Error())
	}
	redirect, errw := cacheRepository.repository.Get(ctx, id)
	if errw != nil {
		return redirect, errw
	}
	ttl := config.ApplicationConfig.Data.Redis.TTL.Redirect
	if err := utils.RedisDatabase.SetStruct(ctx, cacheKey, redirect, ttl); err != nil {
		log.Error(ctx).Msg("Error adding redirect to cache: " + err.Error())
	}
	return redirect, nil
}

func (cacheRepository *RedirectCacheRepository) GetByDNS(ctx context.Context, dns string) (entity.Redirect, *exceptions.WrappedError) {
	cacheKey := REDIRECT_CACHE_KEY_PREFIX + dns
	var redirect entity.Redirect
	cacheErr := utils.RedisDatabase.GetStruct(ctx, cacheKey, &redirect)
	if cacheErr != nil {
		log.Error(ctx).Msg("Error retrieving redirect by DNS from cache: " + cacheErr.Error())
	}
	redirect, errw := cacheRepository.repository.GetByDNS(ctx, dns)
	if errw != nil {
		return redirect, errw
	}
	ttl := config.ApplicationConfig.Data.Redis.TTL.Redirect
	if err := utils.RedisDatabase.SetStruct(ctx, cacheKey, redirect, ttl); err != nil {
		log.Error(ctx).Msg("Error adding redirect by DNS to cache: " + err.Error())
	}
	return redirect, nil
}

func (cacheRepository *RedirectCacheRepository) GetAll(ctx context.Context) ([]entity.Redirect, *exceptions.WrappedError) {
	return cacheRepository.repository.GetAll(ctx)
}

func (cacheRepository *RedirectCacheRepository) Save(ctx context.Context, redirect *entity.Redirect) *exceptions.WrappedError {
	errw := cacheRepository.repository.Save(ctx, redirect)
	if errw != nil {
		return errw
	}

	cacheKey := REDIRECT_CACHE_KEY_PREFIX + redirect.ID
	if err := utils.RedisDatabase.Del(ctx, cacheKey); err != nil {
		log.Error(ctx).Msg("Error removing redirect from cache: " + err.Error())
	}

	dnsCacheKey := REDIRECT_CACHE_KEY_PREFIX + redirect.DNS
	if err := utils.RedisDatabase.Del(ctx, dnsCacheKey); err != nil {
		log.Error(ctx).Msg("Error removing DNS entry from cache: " + err.Error())
	}
	return nil
}

func (cacheRepository *RedirectCacheRepository) Remove(ctx context.Context, redirect entity.Redirect) *exceptions.WrappedError {
	errw := cacheRepository.repository.Remove(ctx, redirect)
	if errw != nil {
		return errw
	}

	cacheKey := REDIRECT_CACHE_KEY_PREFIX + redirect.ID
	if err := utils.RedisDatabase.Del(ctx, cacheKey); err != nil {
		log.Error(ctx).Msg("Error removing redirect from cache: " + err.Error())
	}

	dnsCacheKey := REDIRECT_CACHE_KEY_PREFIX + redirect.DNS
	if err := utils.RedisDatabase.Del(ctx, dnsCacheKey); err != nil {
		log.Error(ctx).Msg("Error removing DNS entry from cache: " + err.Error())
	}

	return nil
}
