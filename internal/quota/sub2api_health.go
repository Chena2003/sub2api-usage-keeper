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

const (
	healthRows           = 7
	healthDefaultColumns = 96
	healthBucketSeconds  = 900 // 15 minutes
)

// BuildSub2APIServiceHealth creates a full grid (rows × columns) spanning the
// requested time window. The window ends at the next local midnight (start of
// tomorrow) so that "today + 6 prior days" are all complete calendar days.
// Data blocks are placed into the matching grid slot; empty slots remain idle
// (Rate=-1). This ensures the frontend grid is always fully occupied.
func BuildSub2APIServiceHealth(blocks []sub2api.HealthBlockRow, hours int) Sub2APIServiceHealth {
	bucketSpan := healthBucketSpan(hours)
	columns := healthDefaultColumns
	totalBlocks := healthRows * columns

	now := time.Now()
	windowEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	windowStart := windowEnd.Add(-time.Duration(totalBlocks) * bucketSpan)

	// Pre-allocate full grid with idle blocks.
	details := make([]Sub2APIServiceHealthBlock, totalBlocks)
	for i := range details {
		start := windowStart.Add(time.Duration(i) * bucketSpan)
		details[i] = Sub2APIServiceHealthBlock{
			StartTime: start,
			EndTime:   start.Add(bucketSpan),
			Rate:      -1,
		}
	}

	// Place data blocks into matching grid slots.
	var totalSuccess, totalFailure int64
	for _, b := range blocks {
		totalSuccess += b.SuccessCount
		totalFailure += b.FailureCount

		idx := int(b.BucketStart.Sub(windowStart) / bucketSpan)
		if idx < 0 || idx >= totalBlocks {
			continue
		}
		// Accumulate into the slot rather than overwrite: when bucketSpan exceeds
		// the SQL bucket granularity (defense-in-depth), multiple SQL buckets can
		// map to one grid slot. With the fixed 168h window bucketSpan is 15min and
		// this is a 1:1 no-op. StartTime/EndTime track the grid slot bounds.
		slot := &details[idx]
		slot.Success += b.SuccessCount
		slot.Failure += b.FailureCount
		total := slot.Success + slot.Failure
		if total > 0 {
			slot.Rate = float64(slot.Success) / float64(total) * 100
		}
	}

	var successRate float64
	total := totalSuccess + totalFailure
	if total > 0 {
		successRate = float64(totalSuccess) / float64(total) * 100
	}

	return Sub2APIServiceHealth{
		TotalSuccess:  totalSuccess,
		TotalFailure:  totalFailure,
		SuccessRate:   successRate,
		Rows:          healthRows,
		Columns:       columns,
		BucketSeconds: int64(bucketSpan / time.Second),
		WindowStart:   windowStart,
		WindowEnd:     windowEnd,
		BlockDetails:  details,
	}
}

// healthBucketSpan picks a bucket duration so that rows×columns blocks cover
// the requested window. For ≤24h use 15-minute buckets (same as CPA short
// range); for longer ranges scale up to fill 672 slots.
func healthBucketSpan(hours int) time.Duration {
	if hours <= 24 {
		return 15 * time.Minute
	}
	totalSeconds := int64(hours) * 3600
	spanSeconds := (totalSeconds + int64(healthRows*healthDefaultColumns) - 1) / int64(healthRows*healthDefaultColumns)
	// Round up to nearest minute for clean display.
	spanMinutes := (spanSeconds + 59) / 60
	if spanMinutes < 15 {
		spanMinutes = 15
	}
	return time.Duration(spanMinutes) * time.Minute
}
