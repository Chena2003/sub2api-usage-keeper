import { useTranslation } from 'react-i18next'
import type { Sub2ApiAccount, Sub2ApiQuotaWindow } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'

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

const windowTone = (window?: Sub2ApiQuotaWindow) => {
  const status = window?.status?.toLowerCase() ?? ''
  if (status.includes('exhaust') || status.includes('limit') || status.includes('risk')) return styles.quotaMeterDanger
  if (status.includes('warning') || (window?.ratio ?? 0) >= 0.65) return styles.quotaMeterWarning
  return styles.quotaMeterOk
}

function QuotaWindowCard({ label, window }: { label: string; window?: Sub2ApiQuotaWindow }) {
  const { t } = useTranslation()
  const meterClass = windowTone(window)

  return (
    <div className={styles.quotaWindowCard}>
      <div className={styles.quotaWindowHeader}>
        <span className={styles.quotaWindowLabel}>{label}</span>
        <span className={`${styles.quotaWindowStatus} ${meterClass}`.trim()}>{window?.status || t('usage_stats.sub2api_no_quota_window')}</span>
      </div>
      {window ? (
        <>
          <div className={styles.quotaMeterShell} aria-hidden="true">
            <span
              className={`${styles.quotaMeterFill} ${meterClass}`.trim()}
              style={{ width: `${Math.min(Math.max((window.ratio ?? 0) * 100, 0), 100)}%` }}
            />
          </div>
          <div className={styles.quotaWindowStats}>
            <span>
              {t('usage_stats.sub2api_consumed')}: {formatNumber(window.consumed)}{window.limit !== undefined ? ` / ${formatNumber(window.limit)}` : ''}
            </span>
            <span>
              {t('usage_stats.sub2api_ratio')}: {formatPercent(window.ratio)}
            </span>
            <span>
              {t('usage_stats.sub2api_refresh_at')}: {formatDateTime(window.refreshAt)}
            </span>
          </div>
        </>
      ) : (
        <div className={styles.quotaWindowEmpty}>{t('usage_stats.sub2api_no_quota_window')}</div>
      )}
    </div>
  )
}

function AccountIdentity({ account }: { account: Sub2ApiAccount }) {
  const { t } = useTranslation()
  const title = account.displayName.trim() || t('usage_stats.sub2api_account')
  const subtitle = [account.provider, account.planType || account.accountType || ''].filter(Boolean).join(' · ')

  return (
    <div className={styles.accountIdentity}>
      <div className={styles.accountIdentityTitleRow}>
        <strong className={styles.accountIdentityTitle}>{title}</strong>
        <span className={styles.accountIdentityBadge}>{account.status}</span>
      </div>
      <div className={styles.accountIdentityMeta}>{subtitle || t('usage_stats.sub2api_account')}</div>
      <div className={styles.accountIdentityMetrics}>
        <span>{t('usage_stats.sub2api_total_requests')}: {formatNumber(account.usage.totalRequests)}</span>
        <span>{t('usage_stats.sub2api_total_tokens')}: {formatNumber(account.usage.totalTokens)}</span>
        <span>{t('usage_stats.sub2api_cost')}: ${account.usage.actualCost.toFixed(4)}</span>
      </div>
    </div>
  )
}

export function AccountQuotasCard({ accounts }: AccountQuotasCardProps) {
  const { t } = useTranslation()

  return (
    <section className={styles.quotaSurface}>
      <div className={styles.sectionTitleBlock}>
        <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_quotas_eyebrow')}</span>
        <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_quotas_title')}</h3>
        <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_quotas_hint')}</p>
      </div>

      {accounts.length === 0 ? (
        <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.quotaAccountList}>
          {accounts.map((account) => (
            <article key={account.id} className={styles.quotaAccountCard}>
              <AccountIdentity account={account} />
              <div className={styles.quotaWindowGrid}>
                <QuotaWindowCard label={t('usage_stats.sub2api_five_hour_window')} window={account.fiveHourWindow} />
                <QuotaWindowCard label={t('usage_stats.sub2api_weekly_window')} window={account.weeklyWindow} />
              </div>
              <div className={styles.quotaAccountFooter}>
                <span>{t('usage_stats.sub2api_last_used')}: {formatDateTime(account.lastUsedAt)}</span>
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
