package modelsdev

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

const DefaultMaxResponseBytes int64 = 16 << 20

var officialProviders = []struct {
	ID   string
	Name string
}{
	{ID: "anthropic", Name: "Anthropic"},
	{ID: "openai", Name: "OpenAI"},
	{ID: "google", Name: "Google"},
	{ID: "xai", Name: "xAI"},
	{ID: "deepseek", Name: "DeepSeek"},
	{ID: "mistral", Name: "Mistral"},
	{ID: "moonshotai", Name: "Moonshot AI"},
	{ID: "alibaba", Name: "Alibaba"},
}

type Cost struct {
	Input      *float64 `json:"input"`
	Output     *float64 `json:"output"`
	CacheRead  *float64 `json:"cacheRead"`
	CacheWrite *float64 `json:"cacheWrite"`
}

type CostTier struct {
	Type string `json:"type"`
	Size int64  `json:"size"`
	Cost Cost   `json:"cost"`
}

type Model struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Status string     `json:"status,omitempty"`
	Cost   Cost       `json:"cost"`
	Tiers  []CostTier `json:"tiers"`
}

type Provider struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Models []Model `json:"models"`
}

type Catalog struct {
	Providers []Provider `json:"providers"`
}

type upstreamProvider struct {
	Models map[string]upstreamModel `json:"models"`
}

type upstreamModel struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Cost   *upstreamCost `json:"cost"`
}

type upstreamCost struct {
	Input      *float64       `json:"input"`
	Output     *float64       `json:"output"`
	CacheRead  *float64       `json:"cache_read"`
	CacheWrite *float64       `json:"cache_write"`
	Tiers      []upstreamTier `json:"tiers"`
}

type upstreamTier struct {
	Tier struct {
		Type string `json:"type"`
		Size int64  `json:"size"`
	} `json:"tier"`
	Input      *float64 `json:"input"`
	Output     *float64 `json:"output"`
	CacheRead  *float64 `json:"cache_read"`
	CacheWrite *float64 `json:"cache_write"`
}

type Client struct {
	apiURL           string
	httpClient       *http.Client
	maxResponseBytes int64
}

func NewClient(apiURL string, httpClient *http.Client, maxResponseBytes int64) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if maxResponseBytes <= 0 {
		maxResponseBytes = DefaultMaxResponseBytes
	}
	return &Client{
		apiURL:           strings.TrimSpace(apiURL),
		httpClient:       httpClient,
		maxResponseBytes: maxResponseBytes,
	}
}

func (c *Client) Fetch(ctx context.Context) (Catalog, error) {
	if c == nil {
		return Catalog{}, fmt.Errorf("models.dev client is nil")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return Catalog{}, fmt.Errorf("create models.dev request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Catalog{}, fmt.Errorf("fetch models.dev response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Catalog{}, fmt.Errorf("models.dev returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil {
		return Catalog{}, fmt.Errorf("read models.dev response: %w", err)
	}
	if int64(len(body)) > c.maxResponseBytes {
		return Catalog{}, fmt.Errorf("models.dev response exceeds %d bytes", c.maxResponseBytes)
	}

	var upstream map[string]upstreamProvider
	if err := json.Unmarshal(body, &upstream); err != nil {
		return Catalog{}, fmt.Errorf("decode models.dev response: %w", err)
	}

	catalog := Catalog{Providers: make([]Provider, 0, len(officialProviders))}
	for _, official := range officialProviders {
		upstreamProvider, ok := upstream[official.ID]
		if !ok {
			continue
		}
		provider := Provider{
			ID:     official.ID,
			Name:   official.Name,
			Models: make([]Model, 0, len(upstreamProvider.Models)),
		}
		for mapID, upstreamModel := range upstreamProvider.Models {
			if upstreamModel.Cost == nil {
				continue
			}
			modelID := strings.TrimSpace(upstreamModel.ID)
			if modelID == "" {
				modelID = strings.TrimSpace(mapID)
			}
			if modelID == "" {
				continue
			}
			modelName := strings.TrimSpace(upstreamModel.Name)
			if modelName == "" {
				modelName = modelID
			}
			model := Model{
				ID:     modelID,
				Name:   modelName,
				Status: upstreamModel.Status,
				Cost: Cost{
					Input:      upstreamModel.Cost.Input,
					Output:     upstreamModel.Cost.Output,
					CacheRead:  upstreamModel.Cost.CacheRead,
					CacheWrite: upstreamModel.Cost.CacheWrite,
				},
				Tiers: make([]CostTier, 0, len(upstreamModel.Cost.Tiers)),
			}
			for _, upstreamTier := range upstreamModel.Cost.Tiers {
				model.Tiers = append(model.Tiers, CostTier{
					Type: upstreamTier.Tier.Type,
					Size: upstreamTier.Tier.Size,
					Cost: Cost{
						Input:      upstreamTier.Input,
						Output:     upstreamTier.Output,
						CacheRead:  upstreamTier.CacheRead,
						CacheWrite: upstreamTier.CacheWrite,
					},
				})
			}
			provider.Models = append(provider.Models, model)
		}
		if len(provider.Models) == 0 {
			continue
		}
		sort.Slice(provider.Models, func(i, j int) bool {
			return provider.Models[i].ID < provider.Models[j].ID
		})
		catalog.Providers = append(catalog.Providers, provider)
	}
	if len(catalog.Providers) == 0 {
		return Catalog{}, fmt.Errorf("no official model pricing found in models.dev response")
	}
	return catalog, nil
}
