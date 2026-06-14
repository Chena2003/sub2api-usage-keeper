import { useTranslation } from 'react-i18next'
import type { Sub2ApiAccount, Sub2ApiQuotaWindow } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import { StatusPill } from '@/components/ui/StatusPill'
import styles from './AccountQuotasCard.module.scss'

type AccountQuotasCardProps = {
  accounts: Sub2ApiAccount[]
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatDateTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString()
}

const formatPercent = (ratio?: number) => {
  if (ratio === undefined || Number.isNaN(ratio)) return '-'
  return `${Math.round(ratio * 100)}%`
}

function quotaVariant(window?: Sub2ApiQuotaWindow): 'ok' | 'warn' | 'danger' | 'neutral' {
  if (!window) return 'neutral'
  const ratio = window.ratio ?? 0
  const status = (window.status ?? '').toLowerCase()
  if (ratio >= 1 || status.includes('exhaust')) return 'danger'
  if (ratio >= 0.8 || status.includes('risk') || status.includes('limit') || status.includes('warn')) return 'warn'
  return 'ok'
}

function QuotaWindow({ label, window }: { label: string; window?: Sub2ApiQuotaWindow }) {
  const { t } = useTranslation()
  const variant = quotaVariant(window)
  const pct = window?.limit ? Math.min(Math.max((window.ratio ?? 0) * 100, 0), 100) : 0

  return (
    <div className={styles.window}>
      <div className={styles.windowTop}>
        <span className={styles.windowLabel}>{label}</span>
        {window ? (
          <StatusPill variant={variant} label={window.status || 'Unknown'} dot={false} />
        ) : (
          <span className={styles.windowEmpty}>{t('usage_stats.sub2api_no_quota_window')}</span>
        )}
      </div>
      {window && (
        <>
          <div className={styles.progressShell}>
            <div
              className={`${styles.progressFill} ${styles[`fill_${variant}`]}`}
              style={{ width: `${pct}%` }}
              aria-label={`${pct.toFixed(0)}% used`}
            />
          </div>
          <div className={styles.windowMeta}>
            <span className={styles.mono}>
              {formatNumber(window.consumed)}
              {window.limit ? ` / ${formatNumber(window.limit)} (${formatPercent(window.ratio)})` : ''}
            </span>
            <span className={styles.windowReset}>{formatDateTime(window.refreshAt)}</span>
          </div>
        </>
      )}
    </div>
  )
}

export function AccountQuotasCard({ accounts, loading, error, onRetry }: AccountQuotasCardProps) {
  const { t } = useTranslation()

  return (
    <Panel
      eyebrow={t('usage_stats.sub2api_quotas_eyebrow')}
      title={t('usage_stats.sub2api_quotas_title')}
      as="section"
    >
      {error ? (
        <Panel.Error message={error} onRetry={onRetry} />
      ) : loading ? (
        <Panel.Loading rows={4} />
      ) : accounts.length === 0 ? (
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      ) : (
        <div className={styles.list}>
          {accounts.map((account) => (
            <article key={account.id} className={styles.row}>
              <div className={styles.rowInfo}>
                <strong className={styles.rowName}>{account.displayName || t('usage_stats.sub2api_account')}</strong>
                <span className={styles.rowProvider}>{account.provider}</span>
              </div>
              <div className={styles.rowMetrics}>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_total_requests')}</span>
                  <span className={styles.mono}>{formatNumber(account.usage.totalRequests)}</span>
                </div>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_total_tokens')}</span>
                  <span className={styles.mono}>{formatNumber(account.usage.totalTokens)}</span>
                </div>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_cost')}</span>
                  <span className={`${styles.mono} ${styles.cost}`}>${account.usage.actualCost.toFixed(4)}</span>
                </div>
              </div>
              <QuotaWindow label={t('usage_stats.sub2api_five_hour_window')} window={account.fiveHourWindow} />
              <QuotaWindow label={t('usage_stats.sub2api_weekly_window')} window={account.weeklyWindow} />
            </article>
          ))}
        </div>
      )}
    </Panel>
  )
}
