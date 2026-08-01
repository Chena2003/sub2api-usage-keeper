package modelsdev

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeCatalogSource struct {
	mu      sync.Mutex
	calls   int
	fetchFn func(call int) (Catalog, error)
}

func (f *fakeCatalogSource) Fetch(context.Context) (Catalog, error) {
	f.mu.Lock()
	f.calls++
	call := f.calls
	fn := f.fetchFn
	f.mu.Unlock()
	return fn(call)
}

func (f *fakeCatalogSource) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestServiceCachesCatalogUntilTTLExpires(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	source := &fakeCatalogSource{fetchFn: func(int) (Catalog, error) {
		return testCatalog("first"), nil
	}}
	service := NewService(source, time.Hour)
	service.now = func() time.Time { return now }

	first, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("first Get returned error: %v", err)
	}
	second, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("second Get returned error: %v", err)
	}
	if source.callCount() != 1 {
		t.Fatalf("expected one source call before expiry, got %d", source.callCount())
	}
	if first.FetchedAt != now.UTC() || second.FetchedAt != now.UTC() || first.Stale || second.Stale {
		t.Fatalf("unexpected cached responses: first=%+v second=%+v", first, second)
	}
}

func TestServiceRefreshesExpiredCatalog(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	source := &fakeCatalogSource{fetchFn: func(call int) (Catalog, error) {
		if call == 1 {
			return testCatalog("first"), nil
		}
		return testCatalog("second"), nil
	}}
	service := NewService(source, time.Hour)
	service.now = func() time.Time { return now }

	first, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("first Get returned error: %v", err)
	}
	now = now.Add(time.Hour + time.Second)
	second, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("second Get returned error: %v", err)
	}
	if source.callCount() != 2 {
		t.Fatalf("expected refresh after expiry, got %d calls", source.callCount())
	}
	if first.Providers[0].Models[0].ID != "first" || second.Providers[0].Models[0].ID != "second" {
		t.Fatalf("expected refreshed data, got first=%+v second=%+v", first, second)
	}
	if second.FetchedAt != now.UTC() || second.Stale {
		t.Fatalf("unexpected refreshed response: %+v", second)
	}
}

func TestServiceReturnsStaleCatalogWhenRefreshFails(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	refreshErr := errors.New("upstream unavailable")
	source := &fakeCatalogSource{fetchFn: func(call int) (Catalog, error) {
		if call == 1 {
			return testCatalog("first"), nil
		}
		return Catalog{}, refreshErr
	}}
	service := NewService(source, time.Hour)
	service.now = func() time.Time { return now }

	fresh, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("fresh Get returned error: %v", err)
	}
	now = now.Add(time.Hour + time.Second)
	stale, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("stale Get returned error: %v", err)
	}
	if !stale.Stale {
		t.Fatalf("expected stale response, got %+v", stale)
	}
	if stale.FetchedAt != fresh.FetchedAt {
		t.Fatalf("expected original fetch time %v, got %v", fresh.FetchedAt, stale.FetchedAt)
	}
	if stale.Providers[0].Models[0].ID != "first" {
		t.Fatalf("expected previous catalog, got %+v", stale)
	}
}

func TestServiceReturnsColdStartError(t *testing.T) {
	source := &fakeCatalogSource{fetchFn: func(int) (Catalog, error) {
		return Catalog{}, errors.New("upstream unavailable")
	}}
	service := NewService(source, time.Hour)

	response, err := service.Get(context.Background())
	if err == nil || err.Error() != "fetch official model pricing: upstream unavailable" {
		t.Fatalf("expected wrapped cold-start error, got response=%+v err=%v", response, err)
	}
	if len(response.Providers) != 0 {
		t.Fatalf("expected empty response on cold start failure, got %+v", response)
	}
}

func TestServiceDeduplicatesConcurrentExpiredRefresh(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	source := &fakeCatalogSource{fetchFn: func(call int) (Catalog, error) {
		if call == 1 {
			return testCatalog("first"), nil
		}
		if call == 2 {
			close(refreshStarted)
			<-releaseRefresh
			return testCatalog("second"), nil
		}
		return Catalog{}, errors.New("unexpected extra refresh")
	}}
	service := NewService(source, time.Hour)
	service.now = func() time.Time { return now }
	if _, err := service.Get(context.Background()); err != nil {
		t.Fatalf("seed Get returned error: %v", err)
	}
	now = now.Add(time.Hour + time.Second)

	const callers = 20
	start := make(chan struct{})
	results := make(chan Response, callers)
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			response, err := service.Get(context.Background())
			results <- response
			errs <- err
		}()
	}
	close(start)
	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("expected refresh to start")
	}
	close(releaseRefresh)
	wg.Wait()
	close(results)
	close(errs)

	if source.callCount() != 2 {
		t.Fatalf("expected one seed fetch and one shared refresh, got %d calls", source.callCount())
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Get returned error: %v", err)
		}
	}
	for response := range results {
		if response.Providers[0].Models[0].ID != "second" || response.Stale {
			t.Fatalf("unexpected concurrent response: %+v", response)
		}
	}
}

func TestServiceReturnsDeepClones(t *testing.T) {
	source := &fakeCatalogSource{fetchFn: func(int) (Catalog, error) {
		return testCatalog("original"), nil
	}}
	service := NewService(source, time.Hour)

	first, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("first Get returned error: %v", err)
	}
	first.Providers[0].ID = "mutated-provider"
	first.Providers[0].Models[0].ID = "mutated-model"
	first.Providers[0].Models[0].Tiers[0].Size = 1
	*first.Providers[0].Models[0].Cost.Input = 999

	second, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("second Get returned error: %v", err)
	}
	model := second.Providers[0].Models[0]
	if second.Providers[0].ID != "anthropic" || model.ID != "original" || model.Tiers[0].Size != 200000 || model.Cost.Input == nil || *model.Cost.Input != 3 {
		t.Fatalf("caller mutation corrupted cached response: %+v", second)
	}
}

func testCatalog(modelID string) Catalog {
	input := 3.0
	output := 15.0
	cacheRead := 0.3
	cacheWrite := 3.75
	tierInput := 6.0
	return Catalog{Providers: []Provider{{
		ID:   "anthropic",
		Name: "Anthropic",
		Models: []Model{{
			ID:   modelID,
			Name: modelID,
			Cost: Cost{Input: &input, Output: &output, CacheRead: &cacheRead, CacheWrite: &cacheWrite},
			Tiers: []CostTier{{
				Type: "context",
				Size: 200000,
				Cost: Cost{Input: &tierInput},
			}},
		}},
	}}}
}
