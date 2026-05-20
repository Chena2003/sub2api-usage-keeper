import { useTranslation } from 'react-i18next'
import type { Sub2ApiAccount, Sub2ApiModelUsage, Sub2ApiOverview, Sub2ApiTimeseriesPoint } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'

type Sub2ApiOverviewPanelProps = {
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  models: Sub2ApiModelUsage[]
  quotaAccounts: Sub2ApiAccount[]
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatCost = (value: number) => `$${value.toFixed(4)}`

const modelTokenTotal = (model: Sub2ApiModelUsage) => (
  model.inputTokens + model.outputTokens + model.cacheCreationTokens + model.cacheReadTokens
)

const isQuotaAtRisk = (account: Sub2ApiAccount) => {
  const windows = [account.fiveHourWindow, account.weeklyWindow].filter(Boolean)
  return windows.some((window) => {
    const status = window?.status.toLowerCase() ?? ''
    return status.includes('risk') || status.includes('limited') || status.includes('exhaust') || status.includes('warning') || (window?.ratio ?? 0) >= 0.8
  })
}

export function Sub2ApiOverviewPanel({ overview, points, models, quotaAccounts }: Sub2ApiOverviewPanelProps) {
  const { t } = useTranslation()
  const hasData = Boolean(overview) || points.length > 0 || models.length > 0 || quotaAccounts.length > 0
  const topModels = models.slice(0, 5)
  const quotaRiskCount = quotaAccounts.filter(isQuotaAtRisk).length

  if (!hasData) {
    return <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
  }

  return (
    <div className={styles.settingsSections}>
      <section className={styles.statsGrid} aria-label={t('usage_stats.sub2api_overview_title')}>
        <article className={styles.statCard}>
          <span className={styles.statLabel}>{t('usage_stats.sub2api_total_requests')}</span>
          <strong className={styles.statValue}>{formatNumber(overview?.totalRequests ?? 0)}</strong>
        </article>
        <article className={styles.statCard}>
          <span className={styles.statLabel}>{t('usage_stats.sub2api_total_tokens')}</span>
          <strong className={styles.statValue}>{formatNumber(overview?.totalTokens ?? 0)}</strong>
        </article>
        <article className={styles.statCard}>
          <span className={styles.statLabel}>{t('usage_stats.sub2api_total_cost')}</span>
          <strong className={styles.statValue}>{formatCost(overview?.actualCost ?? 0)}</strong>
        </article>
        <article className={styles.statCard}>
          <span className={styles.statLabel}>{t('usage_stats.sub2api_active_users')}</span>
          <strong className={styles.statValue}>{formatNumber(overview?.activeUsers ?? 0)}</strong>
        </article>
        <article className={styles.statCard}>
          <strong className={styles.statValue}>{t('usage_stats.sub2api_quota_risk_count', { count: quotaRiskCount })}</strong>
        </article>
      </section>

      <section className="card">
        <div className={styles.sectionTitleBlock}>
          <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_models_eyebrow')}</span>
          <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_top_models')}</h3>
          <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_top_models_hint')}</p>
        </div>
        {topModels.length === 0 ? (
          <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
        ) : (
          <div className={styles.tableWrapper}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>{t('usage_stats.sub2api_model')}</th>
                  <th>{t('usage_stats.sub2api_requests')}</th>
                  <th>{t('usage_stats.sub2api_tokens')}</th>
                  <th>{t('usage_stats.sub2api_cost')}</th>
                </tr>
              </thead>
              <tbody>
                {topModels.map((model) => (
                  <tr key={`${model.model}-${model.requestedModel}-${model.upstreamModel}`}>
                    <td className={styles.modelCell}>{model.model || model.requestedModel || model.upstreamModel || 'unknown'}</td>
                    <td>{formatNumber(model.totalRequests)}</td>
                    <td>{formatNumber(modelTokenTotal(model))}</td>
                    <td>{formatCost(model.actualCost)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
