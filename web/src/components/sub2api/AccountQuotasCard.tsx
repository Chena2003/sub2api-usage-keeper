import { useTranslation } from 'react-i18next'
import type { Sub2ApiAccount, Sub2ApiQuotaWindow } from '@/lib/sub2apiTypes'
import styles from './Sub2ApiDesign.module.scss'

type AccountQuotasCardProps = {
  accounts: Sub2ApiAccount[]
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

const getWindowStatusClass = (status?: string) => {
  const s = status?.toLowerCase() || ''
  if (s.includes('exhaust') || s.includes('limit') || s.includes('risk')) return styles.statusExhausted
  if (s.includes('active') || s.includes('ok')) return styles.statusActive
  return styles.statusUnknown
}

function QuotaWindowInline({ label, window }: { label: string; window?: Sub2ApiQuotaWindow }) {
  const { t } = useTranslation()
  const statusClass = getWindowStatusClass(window?.status)
  const isDanger = (window?.ratio ?? 0) >= 0.8 || window?.status?.includes('exhaust')

  return (
    <div className={styles.windowInline}>
      <div className={styles.windowInlineTop}>
        <span className={styles.windowName}>{label}</span>
        <span className={`${styles.windowStatus} ${statusClass}`}>{window?.status || 'Unknown'}</span>
      </div>
      
      {window ? (
        <>
          <div className={styles.progressBarInline}>
            <div 
              className={`${styles.progressFill} ${isDanger ? styles.progressFillDanger : ''}`}
              style={{ width: `${window.limit ? Math.min(Math.max((window.ratio ?? 0) * 100, 0), 100) : 0}%` }}
            />
          </div>
          <div className={styles.windowDetailsInline}>
            <span>
              {t('usage_stats.used', 'Used')}: {formatNumber(window.consumed)}
              {window.limit ? ` | ${t('usage_stats.remaining', 'Remaining')}: ${formatNumber(Math.max(0, window.limit - window.consumed))} / ${formatNumber(window.limit)}` : ''} 
              {window.limit ? ` (${formatPercent(window.ratio)})` : ''}
            </span>
            <span style={{ color: 'var(--info-color)' }}>{formatDateTime(window.refreshAt)}</span>
          </div>
        </>
      ) : (
        <div style={{ color: 'var(--text-tertiary)', fontSize: '10px' }}>{t('usage_stats.sub2api_no_quota_window')}</div>
      )}
    </div>
  )
}

export function AccountQuotasCard({ accounts }: AccountQuotasCardProps) {
  const { t } = useTranslation()

  return (
    <section className={styles.quotaSurface}>
      <div className={styles.panelHeader}>
        <span className={styles.eyebrow}>{t('usage_stats.sub2api_quotas_eyebrow')}</span>
        <h3 className={styles.title}>{t('usage_stats.sub2api_quotas_title')}</h3>
        <p className={styles.subtitle}>{t('usage_stats.sub2api_quotas_hint')}</p>
      </div>

      {accounts.length === 0 ? (
        <div style={{ color: 'var(--text-tertiary)' }}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.quotaList}>
          {accounts.map((account) => (
            <article key={account.id} className={styles.quotaRow}>
              <div className={styles.quotaInfo}>
                <h4 className={styles.quotaTitle}>{account.displayName || t('usage_stats.sub2api_account')}</h4>
                <span className={styles.quotaProvider}>{account.provider}</span>
              </div>

              <div className={styles.quotaMetricsInline}>
                <div className={styles.metricBlockInline}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_total_requests')}</span>
                  <span className={styles.metricValue}>{formatNumber(account.usage.totalRequests)}</span>
                </div>
                <div className={styles.metricBlockInline}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_total_tokens')}</span>
                  <span className={styles.metricValue}>{formatNumber(account.usage.totalTokens)}</span>
                </div>
                <div className={styles.metricBlockInline}>
                  <span className={styles.metricLabel}>{t('usage_stats.sub2api_cost')}</span>
                  <span className={`${styles.metricValue} ${styles.costValue}`}>${account.usage.actualCost.toFixed(4)}</span>
                </div>
              </div>

              <QuotaWindowInline label={t('usage_stats.sub2api_five_hour_window')} window={account.fiveHourWindow} />
              <QuotaWindowInline label={t('usage_stats.sub2api_weekly_window')} window={account.weeklyWindow} />
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
