package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testSub2APIDatabaseURL = "postgres://sub2api:pw@sub2api-postgres:5432/sub2api?sslmode=disable"

var configEnvKeys = []string{
	"APP_PORT", "APP_BASE_PATH", "WORK_DIR", "SUB2API_DATABASE_URL", "QUOTA_REFRESH_INTERVAL", "PUBLIC_MODE",
	"SQLITE_PATH", "BACKUP_ENABLED", "BACKUP_DIR", "BACKUP_INTERVAL", "BACKUP_RETENTION_DAYS",
	"REQUEST_TIMEOUT", "LOG_LEVEL", "LOG_FILE_ENABLED", "LOG_DIR", "LOG_RETENTION_DAYS",
	"TZ", "TLS_ENABLED", "TLS_CERT_FILE", "TLS_KEY_FILE",
}

func TestMain(m *testing.M) {
	previousEnv := make(map[string]string, len(configEnvKeys))
	previousPresent := make(map[string]bool, len(configEnvKeys))
	for _, key := range configEnvKeys {
		previousEnv[key], previousPresent[key] = os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	code := m.Run()
	for _, key := range configEnvKeys {
		if previousPresent[key] {
			if err := os.Setenv(key, previousEnv[key]); err != nil {
				panic(err)
			}
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	os.Exit(code)
}

func withIsolatedEnvFiles(t *testing.T) {
	t.Helper()
	previousEnv := make(map[string]string, len(configEnvKeys))
	previousPresent := make(map[string]bool, len(configEnvKeys))
	for _, key := range configEnvKeys {
		previousEnv[key], previousPresent[key] = os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
	t.Cleanup(func() {
		for _, key := range configEnvKeys {
			if previousPresent[key] {
				if err := os.Setenv(key, previousEnv[key]); err != nil {
					t.Fatalf("restore %s: %v", key, err)
				}
				continue
			}
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("unset %s: %v", key, err)
			}
		}
	})
	cwd := t.TempDir()
	exeDir := t.TempDir()
	previousExecutableDir := executableDir
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		executableDir = previousExecutableDir
		if err := os.Chdir(previousWorkingDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	executableDir = func() (string, error) { return exeDir, nil }
}

func setRequiredSub2APIEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SUB2API_DATABASE_URL", testSub2APIDatabaseURL)
}

func TestLoadFromEnvAppliesDefaults(t *testing.T) {
	setRequiredSub2APIEnv(t)

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.AppPort != "8080" {
		t.Fatalf("expected default app port 8080, got %s", cfg.AppPort)
	}
	if cfg.AppBasePath != "" {
		t.Fatalf("expected default app base path to be empty, got %q", cfg.AppBasePath)
	}
	if cfg.Sub2APIDatabaseURL != testSub2APIDatabaseURL {
		t.Fatalf("expected sub2api database URL %q, got %q", testSub2APIDatabaseURL, cfg.Sub2APIDatabaseURL)
	}
	if cfg.QuotaRefreshInterval != 5*time.Minute {
		t.Fatalf("expected default quota refresh interval 5m, got %s", cfg.QuotaRefreshInterval)
	}
	if !cfg.PublicMode {
		t.Fatal("expected public mode to be enabled by default")
	}
	if !cfg.BackupEnabled {
		t.Fatal("expected backup to be enabled by default")
	}
	if cfg.WorkDir != filepath.Join(".", "data") {
		t.Fatalf("expected default work dir ./data, got %s", cfg.WorkDir)
	}
	if cfg.BackupDir != filepath.Join("data", "backups") {
		t.Fatalf("expected default backup dir data/backups, got %s", cfg.BackupDir)
	}
	if cfg.BackupInterval != 24*time.Hour {
		t.Fatalf("expected default backup interval 24h, got %s", cfg.BackupInterval)
	}
	if cfg.BackupRetentionDays != 7 {
		t.Fatalf("expected default backup retention 7 days, got %d", cfg.BackupRetentionDays)
	}
	if cfg.RequestTimeout != 30*time.Second {
		t.Fatalf("expected default request timeout 30s, got %s", cfg.RequestTimeout)
	}
	if cfg.SQLitePath != filepath.Join("data", "app.db") {
		t.Fatalf("expected default sqlite path data/app.db, got %s", cfg.SQLitePath)
	}
	if !cfg.LogFileEnabled {
		t.Fatal("expected log file output to be enabled by default")
	}
	if cfg.LogDir != filepath.Join("data", "logs") {
		t.Fatalf("expected default log dir data/logs, got %s", cfg.LogDir)
	}
	if cfg.LogRetentionDays != 7 {
		t.Fatalf("expected default log retention 7 days, got %d", cfg.LogRetentionDays)
	}
}

func TestLoadReadsSpecifiedEnvFile(t *testing.T) {
	withIsolatedEnvFiles(t)
	envDir := t.TempDir()
	envPath := filepath.Join(envDir, "custom.env")
	content := "SUB2API_DATABASE_URL=postgres://from-file\nAPP_PORT=9091\nWORK_DIR=./custom-data\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(LoadOptions{EnvFile: envPath})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Sub2APIDatabaseURL != "postgres://from-file" || cfg.AppPort != "9091" || cfg.WorkDir != filepath.Join(envDir, "custom-data") || cfg.SQLitePath != filepath.Join(envDir, "custom-data", "app.db") || cfg.LogDir != filepath.Join(envDir, "custom-data", "logs") || cfg.BackupDir != filepath.Join(envDir, "custom-data", "backups") {
		t.Fatalf("expected config values from specified env file, got %+v", cfg)
	}
}

func TestLoadResolvesRelativeEnvFilePathBase(t *testing.T) {
	withIsolatedEnvFiles(t)
	cwd := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWorkingDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.Mkdir("config", 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	content := "SUB2API_DATABASE_URL=postgres://relative-env\nWORK_DIR=./data\n"
	if err := os.WriteFile(filepath.Join("config", "app.env"), []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(LoadOptions{EnvFile: filepath.Join("config", "app.env")})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	envFileAbsolutePath, err := filepath.Abs(filepath.Join("config", "app.env"))
	if err != nil {
		t.Fatalf("resolve env file path: %v", err)
	}
	expectedWorkDir := filepath.Join(filepath.Dir(envFileAbsolutePath), "data")
	if cfg.WorkDir != expectedWorkDir || cfg.SQLitePath != filepath.Join(expectedWorkDir, "app.db") || cfg.LogDir != filepath.Join(expectedWorkDir, "logs") || cfg.BackupDir != filepath.Join(expectedWorkDir, "backups") {
		t.Fatalf("expected paths under %q, got %+v", expectedWorkDir, cfg)
	}
}

func TestLoadIgnoresLegacyPathOverrides(t *testing.T) {
	withIsolatedEnvFiles(t)
	envDir := t.TempDir()
	envPath := filepath.Join(envDir, "legacy.env")
	content := "SUB2API_DATABASE_URL=postgres://legacy\nWORK_DIR=./work\nSQLITE_PATH=./legacy/app.db\nLOG_DIR=./legacy/logs\nBACKUP_DIR=./legacy/backups\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(LoadOptions{EnvFile: envPath})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	expectedWorkDir := filepath.Join(envDir, "work")
	if cfg.WorkDir != expectedWorkDir || cfg.SQLitePath != filepath.Join(expectedWorkDir, "app.db") || cfg.LogDir != filepath.Join(expectedWorkDir, "logs") || cfg.BackupDir != filepath.Join(expectedWorkDir, "backups") {
		t.Fatalf("expected legacy path overrides to be ignored, got %+v", cfg)
	}
}

func TestLoadRejectsMissingSpecifiedEnvFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.env")

	_, err := Load(LoadOptions{EnvFile: missingPath})
	if err == nil || !strings.Contains(err.Error(), "stat env file") {
		t.Fatalf("expected missing specified env file error, got %v", err)
	}
}

func TestLoadFallsBackToExecutableDirEnv(t *testing.T) {
	withIsolatedEnvFiles(t)
	exeDir, err := executableDir()
	if err != nil {
		t.Fatalf("get executable dir: %v", err)
	}
	content := "SUB2API_DATABASE_URL=postgres://from-exe\nWORK_DIR=./data\n"
	if err := os.WriteFile(filepath.Join(exeDir, ".env"), []byte(content), 0o600); err != nil {
		t.Fatalf("write executable env file: %v", err)
	}

	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Sub2APIDatabaseURL != "postgres://from-exe" || cfg.WorkDir != filepath.Join(exeDir, "data") || cfg.SQLitePath != filepath.Join(exeDir, "data", "app.db") || cfg.LogDir != filepath.Join(exeDir, "data", "logs") || cfg.BackupDir != filepath.Join(exeDir, "data", "backups") {
		t.Fatalf("expected config values from executable dir env, got %+v", cfg)
	}
}

func TestDefaultTimeZoneIsLoadable(t *testing.T) {
	location, err := time.LoadLocation(DefaultTimeZone)
	if err != nil {
		t.Fatalf("expected default timezone %s to be loadable: %v", DefaultTimeZone, err)
	}
	if location.String() != DefaultTimeZone {
		t.Fatalf("expected location %s, got %s", DefaultTimeZone, location)
	}
}

func TestLoadFromEnvAppliesDefaultTimeZone(t *testing.T) {
	previousLocal := time.Local
	t.Cleanup(func() { time.Local = previousLocal })
	t.Setenv("TZ", "")
	setRequiredSub2APIEnv(t)

	_, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if time.Local.String() != "Asia/Shanghai" {
		t.Fatalf("expected default local timezone Asia/Shanghai, got %s", time.Local)
	}
}

func TestLoadFromEnvHonorsExplicitTimeZone(t *testing.T) {
	previousLocal := time.Local
	t.Cleanup(func() { time.Local = previousLocal })
	t.Setenv("TZ", "UTC")
	setRequiredSub2APIEnv(t)

	_, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if time.Local.String() != "UTC" {
		t.Fatalf("expected explicit local timezone UTC, got %s", time.Local)
	}
}

func TestLoadFromEnvHonorsExplicitIANATimeZone(t *testing.T) {
	previousLocal := time.Local
	t.Cleanup(func() { time.Local = previousLocal })
	t.Setenv("TZ", "America/New_York")
	setRequiredSub2APIEnv(t)

	_, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if time.Local.String() != "America/New_York" {
		t.Fatalf("expected explicit local timezone America/New_York, got %s", time.Local)
	}
}

func TestLoadFromEnvRejectsInvalidTimeZone(t *testing.T) {
	previousLocal := time.Local
	t.Cleanup(func() { time.Local = previousLocal })
	t.Setenv("TZ", "Not/AZone")
	setRequiredSub2APIEnv(t)

	_, err := LoadFromEnv()
	if err == nil || !strings.Contains(err.Error(), "TZ is invalid") {
		t.Fatalf("expected invalid TZ error, got %v", err)
	}
}

func TestLoadRequiresSub2APIDatabaseURL(t *testing.T) {
	withIsolatedEnvFiles(t)
	t.Setenv("AUTH_ENABLED", "false")

	_, err := Load(LoadOptions{})
	if err == nil || !strings.Contains(err.Error(), "SUB2API_DATABASE_URL is required") {
		t.Fatalf("expected SUB2API_DATABASE_URL required error, got %v", err)
	}
}

func TestLoadSub2APISettings(t *testing.T) {
	t.Setenv("SUB2API_DATABASE_URL", testSub2APIDatabaseURL)
	t.Setenv("QUOTA_REFRESH_INTERVAL", "10m")
	t.Setenv("PUBLIC_MODE", "true")

	cfg, err := Load(LoadOptions{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Sub2APIDatabaseURL != testSub2APIDatabaseURL {
		t.Fatalf("expected Sub2APIDatabaseURL %q, got %q", testSub2APIDatabaseURL, cfg.Sub2APIDatabaseURL)
	}
	if cfg.QuotaRefreshInterval != 10*time.Minute {
		t.Fatalf("expected quota refresh interval 10m, got %s", cfg.QuotaRefreshInterval)
	}
	if !cfg.PublicMode {
		t.Fatal("expected public mode to be true")
	}
}

func TestLoadFromEnvIgnoresRemovedLegacySyncEnvVars(t *testing.T) {
	setRequiredSub2APIEnv(t)
	t.Setenv("USAGE_SYNC_MODE", "invalid")
	t.Setenv("POLL_INTERVAL", "not-a-duration")

	_, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should ignore removed legacy sync env vars, got error: %v", err)
	}
}

func TestLoadFromEnvParsesOverrides(t *testing.T) {
	setRequiredSub2APIEnv(t)
	t.Setenv("WORK_DIR", "/tmp/work")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_BASE_PATH", "/sub2api/")
	t.Setenv("BACKUP_ENABLED", "false")
	t.Setenv("BACKUP_INTERVAL", "2h")
	t.Setenv("BACKUP_RETENTION_DAYS", "7")
	t.Setenv("REQUEST_TIMEOUT", "15s")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FILE_ENABLED", "false")
	t.Setenv("LOG_RETENTION_DAYS", "14")
	t.Setenv("QUOTA_REFRESH_INTERVAL", "30m")
	t.Setenv("PUBLIC_MODE", "false")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.AppPort != "9090" || cfg.AppBasePath != "/sub2api" || cfg.WorkDir != "/tmp/work" || cfg.SQLitePath != filepath.Join("/tmp/work", "app.db") || cfg.BackupEnabled || cfg.BackupDir != filepath.Join("/tmp/work", "backups") || cfg.BackupInterval != 2*time.Hour || cfg.BackupRetentionDays != 7 || cfg.RequestTimeout != 15*time.Second || cfg.LogLevel != "debug" || cfg.LogFileEnabled || cfg.LogDir != filepath.Join("/tmp/work", "logs") || cfg.LogRetentionDays != 14 || cfg.QuotaRefreshInterval != 30*time.Minute || cfg.PublicMode {
		t.Fatalf("unexpected config override result: %+v", cfg)
	}
}

func TestLoadFromEnvRejectsNonPositiveQuotaRefreshInterval(t *testing.T) {
	for _, value := range []string{"0s", "-1m"} {
		t.Run(value, func(t *testing.T) {
			setRequiredSub2APIEnv(t)
			t.Setenv("QUOTA_REFRESH_INTERVAL", value)

			_, err := LoadFromEnv()
			if err == nil || err.Error() != "QUOTA_REFRESH_INTERVAL must be positive" {
				t.Fatalf("expected QUOTA_REFRESH_INTERVAL validation error, got %v", err)
			}
		})
	}
}

func TestLoadFromEnvRejectsNonPositiveBackupInterval(t *testing.T) {
	for _, value := range []string{"0s", "-1h"} {
		t.Run(value, func(t *testing.T) {
			setRequiredSub2APIEnv(t)
			t.Setenv("BACKUP_INTERVAL", value)

			_, err := LoadFromEnv()
			if err == nil || err.Error() != "BACKUP_INTERVAL must be positive" {
				t.Fatalf("expected BACKUP_INTERVAL validation error, got %v", err)
			}
		})
	}
}

func TestLoadFromEnvRejectsNegativeBackupRetentionDays(t *testing.T) {
	setRequiredSub2APIEnv(t)
	t.Setenv("BACKUP_RETENTION_DAYS", "-1")

	_, err := LoadFromEnv()
	if err == nil || err.Error() != "BACKUP_RETENTION_DAYS must be non-negative" {
		t.Fatalf("expected BACKUP_RETENTION_DAYS validation error, got %v", err)
	}
}

func TestLoadFromEnvRejectsNegativeLogRetentionDays(t *testing.T) {
	setRequiredSub2APIEnv(t)
	t.Setenv("LOG_RETENTION_DAYS", "-1")

	_, err := LoadFromEnv()
	if err == nil || err.Error() != "LOG_RETENTION_DAYS must be non-negative" {
		t.Fatalf("expected LOG_RETENTION_DAYS validation error, got %v", err)
	}
}

func TestLoadFromEnvRejectsInvalidBasePath(t *testing.T) {
	setRequiredSub2APIEnv(t)
	t.Setenv("APP_BASE_PATH", "sub2api")

	_, err := LoadFromEnv()
	if err == nil || err.Error() != "APP_BASE_PATH is invalid: must start with '/'" {
		t.Fatalf("expected APP_BASE_PATH validation error, got %v", err)
	}
}

