package service

import (
	"context"
	"errors"

	repodto "sub2api-usage-keeper/internal/repository/dto"
	servicedto "sub2api-usage-keeper/internal/service/dto"
)

var ErrInvalidID = errors.New("invalid id")

type UsageProvider interface {
	GetUsageWithFilter(context.Context, servicedto.UsageFilter) (*repodto.StatisticsSnapshot, error)
	GetUsageOverview(context.Context, servicedto.UsageFilter) (*servicedto.UsageOverviewSnapshot, error)
	ListUsageEvents(context.Context, servicedto.UsageFilter) (*servicedto.UsageEventsPage, error)
	ListUsageEventFilterOptions(context.Context, servicedto.UsageFilter) (*servicedto.UsageEventFilterOptions, error)
	GetAnalysis(context.Context, servicedto.UsageFilter) (*servicedto.AnalysisSnapshot, error)
}
