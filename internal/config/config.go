package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// legacyAuthEnvKeys are security-related environment variables from earlier
// designs. This service has NO built-in authentication, so these are silently
// ignored — we warn operators who set them to avoid a false sense of protection.
var legacyAuthEnvKeys = []string{
	"AUTH_ENABLED",
	"LOGIN_PASSWORD",
	"AUTH_SESSION_TTL",
	"PUBLIC_MODE",
}

const DefaultTimeZone = "Asia/Shanghai"

var (
	DefaultWorkDir      = filepath.Join(".", "data")
	DefaultSQLitePath   = filepath.Join(DefaultWorkDir, "app.db")
	DefaultLogDir       = filepath.Join(DefaultWorkDir, "logs")
	DefaultBackupDir    = filepath.Join(DefaultWorkDir, "backups")
	workDirDatabaseName = filepath.Base(DefaultSQLitePath)
	workDirLogsName     = filepath.Base(DefaultLogDir)
	workDirBackupsName  = filepath.Base(DefaultBackupDir)
)

type Config struct {
	// AppPort 是 Web 服务监听端口。
	AppPort string
	// AppBasePath 是 Web 服务部署子路径，空值表示根路径。
	AppBasePath string
	// TLSEnabled 控制是否以 HTTPS 模式启动 HTTP 服务。
	TLSEnabled bool
	// TLSCertFile 是 HTTPS 证书文件路径。
	TLSCertFile string
	// TLSKeyFile 是 HTTPS 私钥文件路径。
	TLSKeyFile string
	// Sub2APIDatabaseURL 是 Sub2API PostgreSQL 数据库连接地址。
	Sub2APIDatabaseURL string
	// TimeZone 是应用时区名称（IANA），用于 time.Local 以及跨服务的 bucket_date 时区契约。
	TimeZone string
	// WorkDir 是应用工作目录，数据库、日志和备份默认从这里派生。
	WorkDir string
	// SQLitePath 是 SQLite 数据库文件路径。
	SQLitePath string
	// BackupEnabled 控制是否保存 SQLite 数据库备份文件。
	BackupEnabled bool
	// BackupDir 是 SQLite 数据库备份目录。
	BackupDir string
	// BackupInterval 是两次备份写入之间的最小间隔。
	BackupInterval time.Duration
	// BackupRetentionDays 是备份文件保留天数。
	BackupRetentionDays int
	// LogLevel 是应用日志级别。
	LogLevel string
	// LogFileEnabled 控制是否写入持久化日志文件。
	LogFileEnabled bool
	// LogDir 是应用日志文件目录。
	LogDir string
	// LogRetentionDays 是日志保留天数，0 表示不自动清理。
	LogRetentionDays int
}

type LoadOptions struct {
	EnvFile string
}

var executableDir = func() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(executablePath), nil
}

func LoadFromEnv() (*Config, error) {
	return Load(LoadOptions{})
}

func Load(options LoadOptions) (*Config, error) {
	envBaseDir, err := loadDotEnv(options)
	if err != nil {
		return nil, err
	}
	timeZone, err := applyProjectTimeZone()
	if err != nil {
		return nil, err
	}

	backupEnabled, err := getBool("BACKUP_ENABLED", true)
	if err != nil {
		return nil, err
	}

	backupInterval, err := getDuration("BACKUP_INTERVAL", 24*time.Hour)
	if err != nil {
		return nil, err
	}
	if backupInterval <= 0 {
		return nil, fmt.Errorf("BACKUP_INTERVAL must be positive")
	}

	backupRetentionDays, err := getInt("BACKUP_RETENTION_DAYS", 7)
	if err != nil {
		return nil, err
	}
	if backupRetentionDays < 0 {
		return nil, fmt.Errorf("BACKUP_RETENTION_DAYS must be non-negative")
	}

	logFileEnabled, err := getBool("LOG_FILE_ENABLED", true)
	if err != nil {
		return nil, err
	}
	logRetentionDays, err := getInt("LOG_RETENTION_DAYS", 7)
	if err != nil {
		return nil, err
	}
	if logRetentionDays < 0 {
		return nil, fmt.Errorf("LOG_RETENTION_DAYS must be non-negative")
	}

	tlsEnabled, err := getBool("TLS_ENABLED", false)
	if err != nil {
		return nil, err
	}

	appBasePath, err := normalizeBasePath(strings.TrimSpace(os.Getenv("APP_BASE_PATH")))
	if err != nil {
		return nil, fmt.Errorf("APP_BASE_PATH is invalid: %w", err)
	}

	workDir := getString("WORK_DIR", DefaultWorkDir)

	cfg := &Config{
		AppPort:              getString("APP_PORT", "8080"),
		AppBasePath:          appBasePath,
		TLSEnabled:           tlsEnabled,
		TLSCertFile:          strings.TrimSpace(os.Getenv("TLS_CERT_FILE")),
		TLSKeyFile:           strings.TrimSpace(os.Getenv("TLS_KEY_FILE")),
		Sub2APIDatabaseURL:   strings.TrimSpace(os.Getenv("SUB2API_DATABASE_URL")),
		TimeZone:             timeZone,
		WorkDir:              workDir,
		SQLitePath:           filepath.Join(workDir, workDirDatabaseName),
		BackupEnabled:        backupEnabled,
		BackupDir:            filepath.Join(workDir, workDirBackupsName),
		BackupInterval:       backupInterval,
		BackupRetentionDays:  backupRetentionDays,
		LogLevel:             getString("LOG_LEVEL", "info"),
		LogFileEnabled:       logFileEnabled,
		LogDir:               filepath.Join(workDir, workDirLogsName),
		LogRetentionDays:     logRetentionDays,
	}
	if cfg.Sub2APIDatabaseURL == "" {
		return nil, fmt.Errorf("SUB2API_DATABASE_URL is required")
	}
	if cfg.TLSEnabled {
		if cfg.TLSCertFile == "" {
			return nil, fmt.Errorf("TLS_CERT_FILE is required when TLS_ENABLED is true")
		}
		if cfg.TLSKeyFile == "" {
			return nil, fmt.Errorf("TLS_KEY_FILE is required when TLS_ENABLED is true")
		}
	}
	cfg.resolveRelativePaths(envBaseDir)

	warnIgnoredLegacyAuthEnv()

	return cfg, nil
}

// warnIgnoredLegacyAuthEnv emits a single warning when any legacy auth-related
// environment variable is set, since this service has no built-in authentication
// and these variables are ignored.
func warnIgnoredLegacyAuthEnv() {
	var present []string
	for _, key := range legacyAuthEnvKeys {
		if _, ok := os.LookupEnv(key); ok {
			present = append(present, key)
		}
	}
	if len(present) > 0 {
		logrus.Warnf("ignoring legacy auth environment variables %s: this service has no built-in authentication; protect it with a reverse proxy or network access control", strings.Join(present, ", "))
	}
}

func applyProjectTimeZone() (string, error) {
	zoneName := strings.TrimSpace(os.Getenv("TZ"))
	if zoneName == "" {
		zoneName = DefaultTimeZone
		if err := os.Setenv("TZ", zoneName); err != nil {
			return "", fmt.Errorf("set default TZ: %w", err)
		}
	}
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		return "", fmt.Errorf("TZ is invalid: %w", err)
	}
	time.Local = location
	return zoneName, nil
}

func loadDotEnv(options LoadOptions) (string, error) {
	if strings.TrimSpace(options.EnvFile) != "" {
		return loadDotEnvFile(options.EnvFile, true)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	if loaded, err := loadOptionalDotEnv(filepath.Join(cwd, ".env")); err != nil || loaded {
		if loaded {
			return cwd, err
		}
		return "", err
	}

	exeDir, err := executableDir()
	if err != nil {
		return "", fmt.Errorf("get executable directory: %w", err)
	}
	loaded, err := loadOptionalDotEnv(filepath.Join(exeDir, ".env"))
	if loaded {
		return exeDir, err
	}
	return "", err
}

func loadOptionalDotEnv(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat .env: %w", err)
	}
	if err := godotenv.Overload(path); err != nil {
		return false, fmt.Errorf("load .env: %w", err)
	}
	return true, nil
}

func loadDotEnvFile(path string, required bool) (string, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) && !required {
			return "", nil
		}
		return "", fmt.Errorf("stat env file: %w", err)
	}
	if err := godotenv.Overload(path); err != nil {
		return "", fmt.Errorf("load env file: %w", err)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve env file path: %w", err)
	}
	return filepath.Dir(absolutePath), nil
}

func (cfg *Config) resolveRelativePaths(baseDir string) {
	if baseDir == "" {
		return
	}
	cfg.WorkDir = resolveRelativePath(baseDir, cfg.WorkDir)
	cfg.SQLitePath = resolveRelativePath(baseDir, cfg.SQLitePath)
	cfg.LogDir = resolveRelativePath(baseDir, cfg.LogDir)
	cfg.BackupDir = resolveRelativePath(baseDir, cfg.BackupDir)
	cfg.TLSCertFile = resolveRelativePath(baseDir, cfg.TLSCertFile)
	cfg.TLSKeyFile = resolveRelativePath(baseDir, cfg.TLSKeyFile)
}

func resolveRelativePath(baseDir, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(baseDir, value)
}

func normalizeBasePath(value string) (string, error) {
	if value == "" || value == "/" {
		return "", nil
	}
	if !strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("must start with '/'")
	}

	normalized := path.Clean(value)
	if normalized == "." || normalized == "/" {
		return "", nil
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return normalized, nil
}

func getString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return duration, nil
}

func getBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid bool: %w", key, err)
	}
	return parsed, nil
}

func getInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return parsed, nil
}
