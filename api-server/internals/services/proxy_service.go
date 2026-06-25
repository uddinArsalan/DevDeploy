package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
)

type ProxyService struct {
	cache cache.Cache
}

func NewProxyService(cache cache.Cache) *ProxyService {
	return &ProxyService{
		cache,
	}
}

func (ps *ProxyService) Route(ctx context.Context, hostname string) (*url.URL, error) {
	port, err := ps.cache.GetPort(ctx, hostname)
	if err != nil {
		return nil, errors.New("404 not found")
	}

	return url.Parse(
		fmt.Sprintf("http://localhost:%v", port),
	)
}
