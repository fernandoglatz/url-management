package service

import (
	"context"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/entity"
	"fernandoglatz/url-management/internal/core/port/repository"
)

type RedirectService struct {
	repository repository.IRedirectRepository
}

func NewRedirectService(repository repository.IRedirectRepository) *RedirectService {
	return &RedirectService{
		repository: repository,
	}
}

func (service *RedirectService) Get(ctx context.Context, id string) (entity.Redirect, *exceptions.WrappedError) {
	return service.repository.Get(ctx, id)
}

func (service *RedirectService) GetByDNS(ctx context.Context, dns string) (entity.Redirect, *exceptions.WrappedError) {
	return service.repository.GetByDNS(ctx, dns)
}

func (service *RedirectService) GetAll(ctx context.Context) ([]entity.Redirect, *exceptions.WrappedError) {
	return service.repository.GetAll(ctx)
}

func (service *RedirectService) Save(ctx context.Context, redirect *entity.Redirect) *exceptions.WrappedError {
	return service.repository.Save(ctx, redirect)
}

func (service *RedirectService) Remove(ctx context.Context, redirect entity.Redirect) *exceptions.WrappedError {
	return service.repository.Remove(ctx, redirect)
}
