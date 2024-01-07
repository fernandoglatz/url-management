package service

import (
	"context"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/entity"
)

type IRedirectService interface {
	Get(ctx context.Context, id string) (entity.Redirect, *exceptions.WrappedError)
	GetByDNS(ctx context.Context, dns string) (entity.Redirect, *exceptions.WrappedError)
	GetAll(ctx context.Context) ([]entity.Redirect, *exceptions.WrappedError)
	Save(ctx context.Context, redirect *entity.Redirect) *exceptions.WrappedError
	Remove(ctx context.Context, redirect entity.Redirect) *exceptions.WrappedError
}
