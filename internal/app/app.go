package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"sub2api-usage-keeper/internal/api"
	"sub2api-usage-keeper/internal/config"
	"sub2api-usage-keeper/internal/logging"
	"sub2api-usage-keeper/internal/repository"
	"sub2api-usage-keeper/internal/service"
	"sub2api-usage-keeper/internal/sub2api"
	webui "sub2api-usage-keeper/web"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Runner 是 App 后台任务的最小接口，具体语义由字段名和实现方法表达。
type Runner interface {
	Run(ctx context.Context) error
}

type Options struct {
	EnvFile string
}

var openSub2APIRepository = sub2api.Open

type storageCleanupService struct {
	db  *gorm.DB
	now func() time.Time
}

func (s storageCleanupService) CleanupStorage(context.Context) error {
	_, err := repository.CleanupStorage(s.db, s.now())
	return err
}

type App struct {
	Config            *config.Config
	DB                *gorm.DB
	Sub2APIRepository *sub2api.Repository
	Router            *gin.Engine
	Maintenance       *StorageCleanupRunner
	BackupMaintenance *DatabaseBackupRunner
	LogCloser         io.Closer

	backgroundCancel context.CancelFunc
	backgroundWG     sync.WaitGroup
}

func New() (*App, error) {
	return NewWithOptions(Options{})
}

func NewWithOptions(options Options) (*App, error) {
	cfg, err := config.Load(config.LoadOptions{EnvFile: options.EnvFile})
	if err != nil {
		return nil, err
	}

	return NewWithConfig(*cfg)
}

func NewWithConfig(cfg config.Config) (*App, error) {
	logCloser, err := logging.Configure(cfg)
	if err != nil {
		return nil, err
	}

	db, err := repository.OpenDatabase(cfg)
	if err != nil {
		_ = logCloser.Close()
		return nil, err
	}
	sub2apiRepo, err := openSub2APIRepository(cfg.Sub2APIDatabaseURL)
	if err != nil {
		_ = closeGormDB(db)
		_ = logCloser.Close()
		return nil, err
	}
	sub2apiDashboardService := service.NewSub2APIDashboardService(sub2apiRepo)

	var backupMaintenance *DatabaseBackupRunner
	if cfg.BackupEnabled {
		sqlDB, err := db.DB()
		if err != nil {
			_ = sub2apiRepo.Close()
			_ = closeGormDB(db)
			_ = logCloser.Close()
			return nil, err
		}
		backupStore := newDatabaseBackupStore(sqlDB, cfg.BackupDir)
		backupMaintenance = NewDatabaseBackupRunner(backupStore, backupStore, cfg.BackupInterval, cfg.BackupRetentionDays)
	}

	usageService := service.NewUsageService(db)

	return &App{
		Config:            &cfg,
		DB:                db,
		Sub2APIRepository: sub2apiRepo,
		Maintenance:       NewStorageCleanupRunner(storageCleanupService{db: db, now: time.Now}),
		BackupMaintenance: backupMaintenance,
		LogCloser:         logCloser,
		Router: api.NewRouter(
			webui.Static,
			usageService,
			cfg.AppBasePath,
			api.OptionalProviders{Sub2APIDashboard: sub2apiDashboardService},
		),
	}, nil
}

func closeGormDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}

	a.stopBackgroundTasks()

	var closeErr error
	if a.Sub2APIRepository != nil {
		closeErr = errors.Join(closeErr, a.Sub2APIRepository.Close())
		a.Sub2APIRepository = nil
	}
	if a.DB != nil {
		closeErr = errors.Join(closeErr, closeGormDB(a.DB))
		a.DB = nil
	}
	if a.LogCloser != nil {
		closeErr = errors.Join(closeErr, a.LogCloser.Close())
		a.LogCloser = nil
	}
	return closeErr
}

func (a *App) Run() error {
	if a == nil || a.Router == nil || a.Config == nil {
		return fmt.Errorf("application is not initialized")
	}

	ctx := a.startBackgroundContext()
	defer a.stopBackgroundTasks()
	if a.Maintenance != nil {
		a.startBackgroundTask(func() {
			if err := a.Maintenance.Run(ctx); err != nil {
				logrus.Errorf("maintenance cleanup stopped: %v", err)
			}
		})
	}
	if a.BackupMaintenance != nil {
		a.startBackgroundTask(func() {
			if err := a.BackupMaintenance.Run(ctx); err != nil {
				logrus.Errorf("database backup stopped: %v", err)
			}
		})
	}

	server := &http.Server{
		Addr:    ":" + a.Config.AppPort,
		Handler: a.Router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if a.Config.TLSEnabled {
		return server.ListenAndServeTLS(a.Config.TLSCertFile, a.Config.TLSKeyFile)
	}
	return server.ListenAndServe()
}

func (a *App) startBackgroundContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	a.backgroundCancel = cancel
	return ctx
}

func (a *App) startBackgroundTask(run func()) {
	a.backgroundWG.Add(1)
	go func() {
		defer a.backgroundWG.Done()
		run()
	}()
}

func (a *App) stopBackgroundTasks() {
	if a.backgroundCancel != nil {
		a.backgroundCancel()
		a.backgroundCancel = nil
	}
	a.backgroundWG.Wait()
}
