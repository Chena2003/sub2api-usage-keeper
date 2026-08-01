package modelsdev

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type catalogSource interface {
	Fetch(context.Context) (Catalog, error)
}

type Response struct {
	Providers []Provider `json:"providers"`
	FetchedAt time.Time  `json:"fetchedAt"`
	Stale     bool       `json:"stale"`
}

type Service struct {
	source catalogSource
	ttl    time.Duration
	now    func() time.Time

	mu     sync.RWMutex
	cached *Response
	group  singleflight.Group
}

func NewService(source catalogSource, ttl time.Duration) *Service {
	return &Service{
		source: source,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *Service) Get(ctx context.Context) (Response, error) {
	if s == nil || s.source == nil {
		return Response{}, fmt.Errorf("fetch official model pricing: models.dev source is unavailable")
	}
	if cached, ok := s.freshCached(); ok {
		return cached, nil
	}

	value, err, _ := s.group.Do("refresh", func() (any, error) {
		if cached, ok := s.freshCached(); ok {
			return cached, nil
		}

		catalog, fetchErr := s.source.Fetch(ctx)
		if fetchErr != nil {
			if stale, ok := s.cachedResponse(); ok {
				stale.Stale = true
				return stale, nil
			}
			return Response{}, fmt.Errorf("fetch official model pricing: %w", fetchErr)
		}

		response := Response{
			Providers: cloneProviders(catalog.Providers),
			FetchedAt: s.now().UTC(),
			Stale:     false,
		}
		s.mu.Lock()
		cached := cloneResponse(response)
		s.cached = &cached
		s.mu.Unlock()
		return response, nil
	})
	if err != nil {
		return Response{}, err
	}
	response, ok := value.(Response)
	if !ok {
		return Response{}, fmt.Errorf("fetch official model pricing: unexpected cache result")
	}
	return cloneResponse(response), nil
}

func (s *Service) freshCached() (Response, bool) {
	now := s.now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cached == nil || now.Sub(s.cached.FetchedAt) >= s.ttl {
		return Response{}, false
	}
	return cloneResponse(*s.cached), true
}

func (s *Service) cachedResponse() (Response, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cached == nil {
		return Response{}, false
	}
	return cloneResponse(*s.cached), true
}

func cloneResponse(response Response) Response {
	response.Providers = cloneProviders(response.Providers)
	return response
}

func cloneProviders(providers []Provider) []Provider {
	if providers == nil {
		return nil
	}
	cloned := make([]Provider, len(providers))
	for providerIndex, provider := range providers {
		cloned[providerIndex] = provider
		if provider.Models == nil {
			continue
		}
		cloned[providerIndex].Models = make([]Model, len(provider.Models))
		for modelIndex, model := range provider.Models {
			clonedModel := model
			clonedModel.Cost = cloneCost(model.Cost)
			if model.Tiers != nil {
				clonedModel.Tiers = make([]CostTier, len(model.Tiers))
				for tierIndex, tier := range model.Tiers {
					clonedTier := tier
					clonedTier.Cost = cloneCost(tier.Cost)
					clonedModel.Tiers[tierIndex] = clonedTier
				}
			}
			cloned[providerIndex].Models[modelIndex] = clonedModel
		}
	}
	return cloned
}

func cloneCost(cost Cost) Cost {
	return Cost{
		Input:      cloneFloat64(cost.Input),
		Output:     cloneFloat64(cost.Output),
		CacheRead:  cloneFloat64(cost.CacheRead),
		CacheWrite: cloneFloat64(cost.CacheWrite),
	}
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
