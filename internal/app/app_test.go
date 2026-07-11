package app

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/config"
	"sub2api-usage-keeper/internal/sub2api"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAppCloseClosesDatabase(t *testing.T) {
	installSub2APITestRepository(t)
	app, err := NewWithConfig(testAppConfig(t))
	if err != nil {
		t.Fatalf("NewWithConfig returned error: %v", err)
	}
	sqlDB, err := app.DB.DB()
	if err != nil {
		t.Fatalf("load sql db: %v", err)
	}

	if err := app.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if err := sqlDB.Ping(); err == nil {
		t.Fatal("expected database ping to fail after app close")
	}
}

func TestNewWithConfigBuildsSub2APIDashboardAndRouter(t *testing.T) {
	installSub2APITestRepository(t)
	app, err := NewWithConfig(testAppConfig(t))
	if err != nil {
		t.Fatalf("NewWithConfig returned error: %v", err)
	}
	defer app.Close()
	if app.Sub2APIRepository == nil {
		t.Fatal("expected sub2api repository to be initialized")
	}
	if app.Router == nil {
		t.Fatal("expected router to be initialized")
	}
	if app.LogCloser == nil {
		t.Fatal("expected log closer to be initialized")
	}
	if app.BackupMaintenance == nil {
		t.Fatal("expected database backup runner to be initialized")
	}
	if app.Maintenance == nil {
		t.Fatal("expected maintenance runner to be initialized")
	}
}

func TestNewWithConfigSkipsBackupRunnerWhenDisabled(t *testing.T) {
	installSub2APITestRepository(t)
	cfg := testAppConfig(t)
	cfg.BackupEnabled = false
	app, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("NewWithConfig returned error: %v", err)
	}
	defer app.Close()
	if app.BackupMaintenance != nil {
		t.Fatal("expected database backup runner to be skipped when backups are disabled")
	}
}

func TestNewWithConfigRequiresSub2APIRepository(t *testing.T) {
	previous := openSub2APIRepository
	openSub2APIRepository = func(string) (*sub2api.Repository, error) {
		return nil, fmt.Errorf("open sub2api postgres: dial failed")
	}
	t.Cleanup(func() { openSub2APIRepository = previous })

	cfg := testAppConfig(t)
	cfg.BackupEnabled = false
	_, err := NewWithConfig(cfg)
	if err == nil {
		t.Fatalf("expected error because Sub2API repository cannot be opened")
	}
	if !strings.Contains(err.Error(), "open sub2api postgres") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunStartsMaintenanceOnly(t *testing.T) {
	cfg := testAppConfig(t)
	cfg.AppPort = "invalid-port"
	maintenanceStarted := make(chan struct{})
	backupStarted := make(chan struct{})
	maintenance := NewStorageCleanupRunner(&maintenanceSyncStub{})
	maintenance.sleep = func(context.Context, time.Duration) bool {
		close(maintenanceStarted)
		return false
	}
	backupRunner := NewDatabaseBackupRunner(&databaseBackupWriterStub{}, nil, time.Second, 0)
	backupRunner.sleep = func(context.Context, time.Duration) bool {
		close(backupStarted)
		return false
	}
	app := &App{
		Config:            &cfg,
		Router:            gin.New(),
		Maintenance:       maintenance,
		BackupMaintenance: backupRunner,
	}

	if err := app.Run(); err == nil {
		t.Fatal("expected Run to return an error for invalid port")
	}
	select {
	case <-maintenanceStarted:
	case <-time.After(time.Second):
		t.Fatal("expected maintenance runner to start")
	}
	select {
	case <-backupStarted:
	case <-time.After(time.Second):
		t.Fatal("expected database backup runner to start")
	}
}

func TestRunCancelsBackgroundTasksWhenRouterStops(t *testing.T) {
	cfg := testAppConfig(t)
	cfg.AppPort = "invalid-port"
	backupStarted := make(chan struct{})
	backupCanceled := make(chan struct{})
	backupRunner := NewDatabaseBackupRunner(&databaseBackupWriterStub{}, nil, time.Second, 0)
	backupRunner.sleep = func(ctx context.Context, _ time.Duration) bool {
		close(backupStarted)
		<-ctx.Done()
		close(backupCanceled)
		return false
	}
	app := &App{
		Config:            &cfg,
		Router:            gin.New(),
		BackupMaintenance: backupRunner,
	}

	if err := app.Run(); err == nil {
		t.Fatal("expected Run to return an error for invalid port")
	}
	select {
	case <-backupStarted:
	case <-time.After(time.Second):
		t.Fatal("expected database backup runner to start")
	}
	select {
	case <-backupCanceled:
	case <-time.After(time.Second):
		t.Fatal("expected database backup runner context to be canceled")
	}
}

func installSub2APITestRepository(t *testing.T) {
	t.Helper()
	previous := openSub2APIRepository
	openSub2APIRepository = func(string) (*sub2api.Repository, error) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return sub2api.NewRepository(db), nil
	}
	t.Cleanup(func() { openSub2APIRepository = previous })
}

func captureAppInfoLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previousOutput := logrus.StandardLogger().Out
	previousFormatter := logrus.StandardLogger().Formatter
	previousLevel := logrus.GetLevel()
	logrus.SetOutput(&logs)
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)
	t.Cleanup(func() {
		logrus.SetOutput(previousOutput)
		logrus.SetFormatter(previousFormatter)
		logrus.SetLevel(previousLevel)
	})
	return &logs
}

func testAppConfig(t *testing.T) config.Config {
	t.Helper()
	return config.Config{
		AppPort:              "8080",
		Sub2APIDatabaseURL:   "postgres://sub2api:pw@sub2api-postgres:5432/sub2api?sslmode=disable",
		SQLitePath:           t.TempDir() + "/app.db",
		BackupEnabled:        true,
		BackupDir:            t.TempDir() + "/backups",
		BackupInterval:       24 * time.Hour,
		BackupRetentionDays:  7,
		LogLevel:             "info",
		LogFileEnabled:       false,
		LogRetentionDays:     7,
	}
}
