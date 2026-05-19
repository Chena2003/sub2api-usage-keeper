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

function QuotaWindowCell({ window }: { window?: Sub2ApiQuotaWindow }) {
  const { t } = useTranslation()

  if (!window) {
    return <span>{t('usage_stats.sub2api_no_quota_window')}</span>
  }

  return (
    <div>
      <div>{t('usage_stats.sub2api_consumed')}: {formatNumber(window.consumed)}{window.limit !== undefined ? ` / ${formatNumber(window.limit)}` : ''}</div>
      <div>{t('usage_stats.sub2api_status')}: {window.status}</div>
      <div>{t('usage_stats.sub2api_refresh_at')}: {formatDateTime(window.refreshAt)}</div>
    </div>
  )
}

export function AccountQuotasCard({ accounts }: AccountQuotasCardProps) {
  const { t } = useTranslation()

  return (
    <section className="card">
      <div className={styles.sectionTitleBlock}>
        <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_quotas_eyebrow')}</span>
        <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_quotas_title')}</h3>
        <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_quotas_hint')}</p>
      </div>

      {accounts.length === 0 ? (
        <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t('usage_stats.sub2api_account')}</th>
                <th>{t('usage_stats.sub2api_provider')}</th>
                <th>{t('usage_stats.sub2api_plan')}</th>
                <th>{t('usage_stats.sub2api_status')}</th>
                <th>{t('usage_stats.sub2api_five_hour_window')}</th>
                <th>{t('usage_stats.sub2api_weekly_window')}</th>
                <th>{t('usage_stats.sub2api_last_used')}</th>
              </tr>
            </thead>
            <tbody>
              {accounts.map((account) => (
                <tr key={account.id}>
                  <td className={styles.modelCell}>{account.displayName}</td>
                  <td>{account.provider}</td>
                  <td>{account.planType || account.accountType || '-'}</td>
                  <td>{account.status}</td>
                  <td><QuotaWindowCell window={account.fiveHourWindow} /></td>
                  <td><QuotaWindowCell window={account.weeklyWindow} /></td>
                  <td>{formatDateTime(account.lastUsedAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
