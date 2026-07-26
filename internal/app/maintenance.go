package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type StorageCleanupSyncer interface {
	CleanupStorage(ctx context.Context) error
}

type StorageCleanupRunner struct {
	syncer StorageCleanupSyncer
	now    func() time.Time
	sleep  func(context.Context, time.Duration) bool

	mu      sync.Mutex
	running bool
}

func NewStorageCleanupRunner(syncer StorageCleanupSyncer) *StorageCleanupRunner {
	return &StorageCleanupRunner{
		syncer: syncer,
		now:    time.Now,
		sleep:  maintenanceSleepContext,
	}
}

// Run 每天按项目本地时区 03:00 执行一次统一存储清理，失败只记录日志，不终止后台任务。
func (r *StorageCleanupRunner) Run(ctx context.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	logrus.Info("storage cleanup task started")
	r.setRunning(true)
	defer r.setRunning(false)

	for {
		now := r.now()
		delay := nextDailyCleanupAt(now).Sub(now)
		if delay < 0 {
			delay = 0
		}
		if !r.sleep(ctx, delay) {
			return nil
		}
		if err := r.syncer.CleanupStorage(ctx); err != nil {
			logrus.WithError(err).Error("storage cleanup failed")
		}
	}
}

// nextDailyCleanupAt 用 now 所在时区计算下一次 03:00，因此 TZ 同时控制业务日期边界和清理触发时间。
// 生产环境下 now 来自 time.Now()，其 Location 就是 time.Local，行为与此前一致。
func nextDailyCleanupAt(now time.Time) time.Time {
	loc := now.Location()
	cleanupAt := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, loc)
	if !now.Before(cleanupAt) {
		cleanupAt = cleanupAt.AddDate(0, 0, 1)
	}
	return cleanupAt
}

func (r *StorageCleanupRunner) validate() error {
	if r == nil {
		return fmt.Errorf("storage cleanup runner is nil")
	}
	if r.syncer == nil {
		return fmt.Errorf("storage cleanup syncer is nil")
	}
	if r.now == nil {
		r.now = time.Now
	}
	if r.sleep == nil {
		r.sleep = maintenanceSleepContext
	}
	return nil
}

func maintenanceSleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (r *StorageCleanupRunner) setRunning(running bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = running
}
