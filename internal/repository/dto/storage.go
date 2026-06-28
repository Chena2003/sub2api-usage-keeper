package dto

// StorageCleanupResult 是仓储层每日清理的结果。
type StorageCleanupResult struct {
	RedisInbox RedisUsageInboxCleanupResult
}

// RedisUsageInboxCleanupResult 是 Redis usage inbox 的清理结果。
type RedisUsageInboxCleanupResult struct {
	ProcessedDeleted int64
	FailedDeleted    int64
}
