package quota

import (
	"testing"
	"time"

	"sub2api-usage-keeper/internal/sub2api"
)

func TestBuildSub2APIServiceHealth_WindowBounds(t *testing.T) {
	result := BuildSub2APIServiceHealth(nil, 168)

	if result.Rows != healthRows {
		t.Fatalf("Rows = %d, want %d", result.Rows, healthRows)
	}
	if result.Columns != healthDefaultColumns {
		t.Fatalf("Columns = %d, want %d", result.Columns, healthDefaultColumns)
	}
	if len(result.BlockDetails) != healthRows*healthDefaultColumns {
		t.Fatalf("BlockDetails length = %d, want %d", len(result.BlockDetails), healthRows*healthDefaultColumns)
	}
	if result.BucketSeconds != 900 {
		t.Fatalf("BucketSeconds = %d, want 900", result.BucketSeconds)
	}

	// WindowEnd must be start of tomorrow in local time (midnight).
	if result.WindowEnd.Hour() != 0 || result.WindowEnd.Minute() != 0 || result.WindowEnd.Second() != 0 {
		t.Fatalf("WindowEnd not at midnight: %v", result.WindowEnd)
	}
	if result.WindowEnd.Location() != time.Local {
		t.Fatalf("WindowEnd location = %v, want %v", result.WindowEnd.Location(), time.Local)
	}

	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	if !result.WindowEnd.Equal(tomorrow) {
		t.Fatalf("WindowEnd = %v, want %v", result.WindowEnd, tomorrow)
	}

	expectedDuration := time.Duration(healthRows*healthDefaultColumns) * 15 * time.Minute
	actualDuration := result.WindowEnd.Sub(result.WindowStart)
	if actualDuration != expectedDuration {
		t.Fatalf("window span = %v, want %v", actualDuration, expectedDuration)
	}
}

func TestBuildSub2APIServiceHealth_BlockPlacement(t *testing.T) {
	now := time.Now()
	windowEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	totalBlocks := healthRows * healthDefaultColumns
	windowStart := windowEnd.Add(-time.Duration(totalBlocks) * 15 * time.Minute)

	blockAtStart := sub2api.HealthBlockRow{
		BucketStart:  windowStart,
		BucketEnd:    windowStart.Add(15 * time.Minute),
		SuccessCount: 10,
		FailureCount: 2,
		TotalCount:   12,
	}

	blockBeforeWindow := sub2api.HealthBlockRow{
		BucketStart:  windowStart.Add(-15 * time.Minute),
		BucketEnd:    windowStart,
		SuccessCount: 5,
		FailureCount: 1,
		TotalCount:   6,
	}

	result := BuildSub2APIServiceHealth([]sub2api.HealthBlockRow{blockAtStart, blockBeforeWindow}, 168)

	// Block at windowStart should land at index 0.
	if result.BlockDetails[0].Success != 10 || result.BlockDetails[0].Failure != 2 {
		t.Fatalf("idx 0: got success=%d failure=%d, want 10/2", result.BlockDetails[0].Success, result.BlockDetails[0].Failure)
	}

	// Block before window should be dropped — totals only include in-window blocks.
	// But totalSuccess/totalFailure counts ALL blocks (even out-of-window) because
	// the original logic sums before checking idx. Verify the sum includes both.
	if result.TotalSuccess != 15 {
		t.Fatalf("TotalSuccess = %d, want 15", result.TotalSuccess)
	}
	if result.TotalFailure != 3 {
		t.Fatalf("TotalFailure = %d, want 3", result.TotalFailure)
	}
}

func TestBuildSub2APIServiceHealth_IdleSlots(t *testing.T) {
	result := BuildSub2APIServiceHealth(nil, 168)

	for i, block := range result.BlockDetails {
		if block.Rate != -1 {
			t.Fatalf("BlockDetails[%d].Rate = %f, want -1 (idle)", i, block.Rate)
		}
	}
}

// TestBuildSub2APIServiceHealth_FixedWindowIgnoresShortHours proves the grid is
// always 168h / 672 15-min blocks even when the page range (hours) is small.
func TestBuildSub2APIServiceHealth_FixedWindowIgnoresShortHours(t *testing.T) {
	for _, hours := range []int{8, 24, 168} {
		result := BuildSub2APIServiceHealth(nil, hours)
		if len(result.BlockDetails) != healthRows*healthDefaultColumns {
			t.Fatalf("hours=%d: BlockDetails length = %d, want %d", hours, len(result.BlockDetails), healthRows*healthDefaultColumns)
		}
		if result.BucketSeconds != 900 {
			t.Fatalf("hours=%d: BucketSeconds = %d, want 900", hours, result.BucketSeconds)
		}
		expectedDuration := time.Duration(healthRows*healthDefaultColumns) * 15 * time.Minute
		if got := result.WindowEnd.Sub(result.WindowStart); got != expectedDuration {
			t.Fatalf("hours=%d: window span = %v, want %v (168h)", hours, got, expectedDuration)
		}
	}
}

// TestBuildSub2APIServiceHealth_AccumulatesColocatedBlocks proves that when two
// SQL buckets land in the same grid slot their counts accumulate (defense-in-depth
// for bucketSpan > SQL granularity) rather than the second overwriting the first.
func TestBuildSub2APIServiceHealth_AccumulatesColocatedBlocks(t *testing.T) {
	now := time.Now()
	windowEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	totalBlocks := healthRows * healthDefaultColumns
	windowStart := windowEnd.Add(-time.Duration(totalBlocks) * 15 * time.Minute)

	// Two blocks both mapping to index 0 (same 15-min grid slot).
	blockA := sub2api.HealthBlockRow{
		BucketStart:  windowStart,
		BucketEnd:    windowStart.Add(15 * time.Minute),
		SuccessCount: 8,
		FailureCount: 2,
		TotalCount:   10,
	}
	blockB := sub2api.HealthBlockRow{
		BucketStart:  windowStart.Add(5 * time.Minute),
		BucketEnd:    windowStart.Add(15 * time.Minute),
		SuccessCount: 2,
		FailureCount: 3,
		TotalCount:   5,
	}

	result := BuildSub2APIServiceHealth([]sub2api.HealthBlockRow{blockA, blockB}, 168)

	if result.BlockDetails[0].Success != 10 || result.BlockDetails[0].Failure != 5 {
		t.Fatalf("idx 0: got success=%d failure=%d, want 10/5", result.BlockDetails[0].Success, result.BlockDetails[0].Failure)
	}
	// rate = 10 / 15 * 100
	if got := result.BlockDetails[0].Rate; got < 66.6 || got > 66.7 {
		t.Fatalf("idx 0 rate = %f, want ~66.67", got)
	}
}
