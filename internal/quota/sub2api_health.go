package quota

import (
	"time"

	"sub2api-usage-keeper/internal/sub2api"
)

type Sub2APIServiceHealthBlock struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Success   int64     `json:"success"`
	Failure   int64     `json:"failure"`
	Rate      float64   `json:"rate"` // -1 means idle/no data
}

type Sub2APIServiceHealth struct {
	TotalSuccess  int64                       `json:"total_success"`
	TotalFailure  int64                       `json:"total_failure"`
	SuccessRate   float64                     `json:"success_rate"`
	Rows          int                         `json:"rows"`
	Columns       int                         `json:"columns"`
	BucketSeconds int64                       `json:"bucket_seconds"`
	WindowStart   time.Time                   `json:"window_start"`
	WindowEnd     time.Time                   `json:"window_end"`
	BlockDetails  []Sub2APIServiceHealthBlock `json:"block_details"`
}

func BuildSub2APIServiceHealth(blocks []sub2api.HealthBlockRow) Sub2APIServiceHealth {
	if len(blocks) == 0 {
		return Sub2APIServiceHealth{
			Rows:         7,
			Columns:      0,
			BlockDetails: []Sub2APIServiceHealthBlock{},
		}
	}

	var totalSuccess, totalFailure int64
	for _, b := range blocks {
		totalSuccess += b.SuccessCount
		totalFailure += b.FailureCount
	}

	var successRate float64
	total := totalSuccess + totalFailure
	if total > 0 {
		successRate = float64(totalSuccess) / float64(total) * 100
	}

	details := make([]Sub2APIServiceHealthBlock, len(blocks))
	for i, b := range blocks {
		rate := float64(-1)
		if b.TotalCount > 0 {
			rate = float64(b.SuccessCount) / float64(b.TotalCount) * 100
		}
		details[i] = Sub2APIServiceHealthBlock{
			StartTime: b.BucketStart,
			EndTime:   b.BucketEnd,
			Success:   b.SuccessCount,
			Failure:   b.FailureCount,
			Rate:      rate,
		}
	}

	return Sub2APIServiceHealth{
		TotalSuccess:  totalSuccess,
		TotalFailure:  totalFailure,
		SuccessRate:   successRate,
		Rows:          7,
		Columns:       len(blocks),
		BucketSeconds: 900,
		WindowStart:   blocks[0].BucketStart,
		WindowEnd:     blocks[len(blocks)-1].BucketEnd,
		BlockDetails:  details,
	}
}
