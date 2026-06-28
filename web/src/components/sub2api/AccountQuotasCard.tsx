import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Sub2ApiAccount, Sub2ApiQuotaWindow } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import { StatusPill } from '@/components/ui/StatusPill'
import { formatCompactNumber } from '@/utils/usage'
import styles from './AccountQuotasCard.module.scss'

type AccountQuotasCardProps = {
  accounts: Sub2ApiAccount[]
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}

type PillVariant = 'ok' | 'warn' | 'danger' | 'neutral' | 'info'

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatDuration = (ms: number) => {
  if (!ms || ms <= 0) return '—'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(ms < 10000 ? 2 : 1)}s`
}

// utilization 0-100 -> used semantics: <80 ok, 80-100 warn, >=100 danger
function utilizationTone(util: number): 'ok' | 'warn' | 'danger' {
  if (util >= 100) return 'danger'
  if (util >= 80) return 'warn'
  return 'ok'
}

const STATUS_VARIANT: Record<string, PillVariant> = {
  active: 'ok',
  paused: 'neutral',
  inactive: 'neutral',
  rate_limited: 'warn',
  temp_unschedulable: 'warn',
  overloaded: 'danger',
  error: 'danger',
}

function useResetCountdown() {
  // mount-time captured via lazy initializer to stay out of the render path
  const [mountTime] = useState(() => Date.now())
  return (resetAt?: string): string => {
    if (!resetAt) return ''
    const target = Date.parse(resetAt)
    if (Number.isNaN(target)) return ''
    const diffMs = target - mountTime
    if (diffMs <= 0) return ''
    const totalMinutes = Math.floor(diffMs / 60000)
    const days = Math.floor(totalMinutes / (24 * 60))
    const hours = Math.floor((totalMinutes % (24 * 60)) / 60)
    const minutes = totalMinutes % 60
    if (days > 0) return `${days}d ${hours}h`
    if (hours > 0) return `${hours}h ${minutes}m`
    return `${minutes}m`
  }
}

function AccountStatus({ account, countdown }: { account: Sub2ApiAccount; countdown: (resetAt?: string) => string }) {
  const { t } = useTranslation()
  const detail = account.statusDetail || account.status || 'active'
  const variant = STATUS_VARIANT[detail] ?? 'neutral'
  const labelKey = `usage_stats.sub2api_status_${detail}`
  const label = t(labelKey, { defaultValue: detail })
  const reset = countdown(account.statusResetAt)

  return (
    <span className={styles.statusGroup}>
      <StatusPill variant={variant} label={label} dot />
      {reset && <span className={styles.statusReset}>{reset}</span>}
      {account.hasError && (
        <span className={styles.errorDot} role="img" aria-label={t('usage_stats.sub2api_status_error', { defaultValue: 'error' })} />
      )}
    </span>
  )
}

function Metric({ label, value, tone }: { label: string; value: string; tone?: 'cost' }) {
  return (
    <div className={styles.metric}>
      <span className={styles.metricLabel}>{label}</span>
      <span className={`${styles.metricValue} ${tone === 'cost' ? styles.metricCost : ''}`}>{value}</span>
    </div>
  )
}

function QuotaBar({ label, window, countdown }: { label: string; window?: Sub2ApiQuotaWindow; countdown: (resetAt?: string) => string }) {
  const { t } = useTranslation()
  const util = window?.utilization
  const hasUtil = util != null && Number.isFinite(util)

  if (!window || !hasUtil) {
    return (
      <div className={styles.quotaBlock}>
        <div className={styles.quotaHeader}>
          <span className={styles.quotaLabel}>{label}</span>
        </div>
        <div className={styles.quotaEmpty}>{t('usage_stats.sub2api_no_quota_window')}</div>
      </div>
    )
  }

  const tone = utilizationTone(util)
  const width = `${Math.min(Math.max(util, 0), 100)}%`
  const percentLabel = util > 999 ? '>999%' : `${Math.round(util)}%`
  const reset = countdown(window.refreshAt)

  return (
    <div className={styles.quotaBlock}>
      <div className={styles.quotaHeader}>
        <span className={styles.quotaLabel}>{label}</span>
        <span className={styles.quotaValueGroup}>
          {reset && <span className={styles.quotaReset}>{reset}</span>}
          <strong className={`${styles.quotaPercent} ${styles[`tone_${tone}`]}`}>{percentLabel}</strong>
        </span>
      </div>
      <div className={styles.quotaTrack}>
        <span
          className={`${styles.quotaFill} ${styles[`fill_${tone}`]}`}
          style={{ width }}
          aria-label={`${percentLabel} ${t('usage_stats.sub2api_consumed', { defaultValue: 'used' })}`}
        />
      </div>
      {window.consumed > 0 && (
        <div className={styles.quotaMeta}>
          <span className={styles.quotaConsumed}>{formatCompactNumber(window.consumed)}</span>
        </div>
      )}
    </div>
  )
}

export function AccountQuotasCard({ accounts, loading, error, onRetry }: AccountQuotasCardProps) {
  const { t } = useTranslation()
  const countdown = useResetCountdown()

  const countLabel = t('usage_stats.sub2api_quotas_count', {
    count: accounts.length,
    defaultValue: `${accounts.length} accounts`,
  })

  return (
    <Panel
      eyebrow={t('usage_stats.sub2api_quotas_eyebrow')}
      title={t('usage_stats.sub2api_quotas_title')}
      as="section"
      actions={accounts.length > 0 ? <span className={styles.countBadge}>{countLabel}</span> : undefined}
    >
      <p className={styles.subtitle}>{t('usage_stats.sub2api_quotas_hint')}</p>

      {error ? (
        <Panel.Error message={error} onRetry={onRetry} />
      ) : loading ? (
        <Panel.Loading rows={4} />
      ) : accounts.length === 0 ? (
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      ) : (
        <div className={styles.list}>
          <div className={styles.tableHeader} aria-hidden="true">
            <span>{t('usage_stats.sub2api_account')}</span>
            <span className={styles.headerMetrics}>
              <span>{t('usage_stats.sub2api_total_requests')}</span>
              <span>{t('usage_stats.sub2api_total_tokens')}</span>
              <span>{t('usage_stats.sub2api_cost')}</span>
              <span>{t('usage_stats.sub2api_avg_duration', { defaultValue: 'Avg Duration' })}</span>
            </span>
            <span className={styles.headerQuota}>
              {t('usage_stats.sub2api_five_hour_window')} · {t('usage_stats.sub2api_weekly_window')}
            </span>
          </div>

          {accounts.map((account) => (
            <article key={account.id} className={styles.row}>
              <div className={styles.identity}>
                <div className={styles.nameRow}>
                  <strong className={styles.name}>{account.displayName || t('usage_stats.sub2api_account')}</strong>
                </div>
                <div className={styles.badges}>
                  <span className={styles.providerBadge}>{account.provider}</span>
                  {account.planType && <span className={styles.planBadge}>{account.planType}</span>}
                </div>
                <AccountStatus account={account} countdown={countdown} />
              </div>

              <div className={styles.metrics}>
                <Metric label={t('usage_stats.sub2api_total_requests')} value={formatNumber(account.usage.totalRequests)} />
                <Metric label={t('usage_stats.sub2api_total_tokens')} value={formatCompactNumber(account.usage.totalTokens)} />
                <Metric label={t('usage_stats.sub2api_cost')} value={`$${account.usage.actualCost.toFixed(2)}`} tone="cost" />
                <Metric label={t('usage_stats.sub2api_avg_duration', { defaultValue: 'Avg Duration' })} value={formatDuration(account.usage.averageDurationMs)} />
              </div>

              <div className={styles.quotas}>
                <QuotaBar label={t('usage_stats.sub2api_five_hour_window')} window={account.fiveHourWindow} countdown={countdown} />
                <QuotaBar label={t('usage_stats.sub2api_weekly_window')} window={account.weeklyWindow} countdown={countdown} />
              </div>
            </article>
          ))}
        </div>
      )}
    </Panel>
  )
}
