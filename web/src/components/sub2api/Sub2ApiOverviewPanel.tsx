import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Line, Bar } from 'react-chartjs-2'
import type { ChartData, ChartOptions } from 'chart.js'
import type { Sub2ApiAccount, Sub2ApiModelUsage, Sub2ApiOverview, Sub2ApiTimeseriesPoint } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'
import designStyles from './Sub2ApiDesign.module.scss'
import { RequestHealthTimelineCard } from './RequestHealthTimelineCard'
import type { UsageOverviewPayload } from '@/components/usage/hooks/useUsageData'

type Sub2ApiOverviewPanelProps = {
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  models: Sub2ApiModelUsage[]
  quotaAccounts: Sub2ApiAccount[]
  usage?: UsageOverviewPayload | null
  loading?: boolean
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatCost = (value: number) => `$${value.toFixed(4)}`

const modelTokenTotal = (model: Sub2ApiModelUsage) => (
  model.inputTokens + model.outputTokens + model.cacheCreationTokens + model.cacheReadTokens
)

const pointTokenTotal = (point: Sub2ApiTimeseriesPoint) => (
  point.inputTokens + point.outputTokens + point.cacheCreationTokens + point.cacheReadTokens
)

const isQuotaAtRisk = (account: Sub2ApiAccount) => {
  const windows = [account.fiveHourWindow, account.weeklyWindow].filter(Boolean)
  return windows.some((window) => {
    const status = window?.status.toLowerCase() ?? ''
    return status.includes('risk') || status.includes('limited') || status.includes('exhaust') || status.includes('warning') || (window?.ratio ?? 0) >= 0.8
  })
}

export function Sub2ApiOverviewPanel({ overview, points, models, quotaAccounts, usage, loading }: Sub2ApiOverviewPanelProps) {
  const { t } = useTranslation()
  const hasData = Boolean(overview) || points.length > 0 || models.length > 0 || quotaAccounts.length > 0
  const topModels = models.slice(0, 5)
  const maxModelTokens = Math.max(...topModels.map(modelTokenTotal), 1)
  const quotaRiskCount = quotaAccounts.filter(isQuotaAtRisk).length
  const availableAccounts = quotaAccounts.length - quotaRiskCount

  const fmtHour = (s: string) => {
    const d = new Date(s)
    return Number.isNaN(d.getTime()) ? s : `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  const fmtDate = (s: string) => {
    const d = new Date(s)
    return Number.isNaN(d.getTime()) ? s : `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:00`
  }

  // Chart data transforms
  const chartLabels = useMemo(() => points.map((p) => fmtHour(p.bucketStart)), [points])

  const tokenChartData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Input',
          data: points.map((p) => p.inputTokens),
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59, 130, 246, 0.12)',
          fill: true,
          tension: 0.25,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 1.5,
        },
        {
          label: 'Output',
          data: points.map((p) => p.outputTokens),
          borderColor: '#22c55e',
          backgroundColor: 'rgba(34, 197, 94, 0.12)',
          fill: true,
          tension: 0.25,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 1.5,
        },
        {
          label: 'Cache',
          data: points.map((p) => p.cacheCreationTokens + p.cacheReadTokens),
          borderColor: '#f59e0b',
          backgroundColor: 'rgba(245, 158, 11, 0.10)',
          fill: true,
          tension: 0.25,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 1.5,
        },
      ],
    }
  }, [points, chartLabels])

  const requestChartData = useMemo((): ChartData<'bar'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Requests',
          data: points.map((p) => p.totalRequests),
          backgroundColor: 'rgba(16, 185, 129, 0.25)',
          borderColor: '#10b981',
          borderWidth: 1,
          borderRadius: 4,
        },
      ],
    }
  }, [points, chartLabels])

  const costChartData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Cost',
          data: points.map((p) => p.actualCost),
          borderColor: '#f59e0b',
          backgroundColor: 'rgba(245, 158, 11, 0.10)',
          fill: true,
          tension: 0.35,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 2,
        },
      ],
    }
  }, [points, chartLabels])

  const usersChartData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Active Users',
          data: points.map((p) => p.activeUsers),
          borderColor: '#14b8a6',
          backgroundColor: 'rgba(20, 184, 166, 0.10)',
          fill: true,
          tension: 0.35,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 2,
        },
      ],
    }
  }, [points, chartLabels])

  const tokensChartData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Total Tokens',
          data: points.map(pointTokenTotal),
          borderColor: '#64748b',
          backgroundColor: 'rgba(100, 116, 139, 0.10)',
          fill: true,
          tension: 0.35,
          pointRadius: 0,
          pointHitRadius: 8,
          borderWidth: 2,
        },
      ],
    }
  }, [points, chartLabels])

  const chartOptions: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false }, tooltip: { enabled: true, mode: 'index', intersect: false } },
    scales: {
      x: { display: true, grid: { display: false }, ticks: { color: '#9ca3af', font: { size: 10, family: 'inherit' }, maxTicksLimit: 8 } },
      y: { display: true, grid: { color: 'rgba(0,0,0,0.04)', drawTicks: false }, ticks: { color: '#9ca3af', font: { size: 10, family: 'inherit' }, maxTicksLimit: 5 } },
    },
    interaction: { mode: 'index', intersect: false },
  }

  const barOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false }, tooltip: { enabled: true, mode: 'index', intersect: false } },
    scales: {
      x: { display: true, grid: { display: false }, ticks: { color: '#9ca3af', font: { size: 10, family: 'inherit' }, maxTicksLimit: 8 } },
      y: { display: true, grid: { color: 'rgba(0,0,0,0.04)', drawTicks: false }, beginAtZero: true, ticks: { color: '#9ca3af', font: { size: 10, family: 'inherit' }, maxTicksLimit: 5 } },
    },
    interaction: { mode: 'index', intersect: false },
  }

  if (!hasData) {
    return <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
  }

  return (
    <div className={styles.settingsSections}>
      <section className={designStyles.statsGrid} aria-label={t('usage_stats.sub2api_overview_title')}>
        <article className={designStyles.statCard}>
          <span className={designStyles.statLabel}>{t('usage_stats.sub2api_total_requests')}</span>
          <strong className={designStyles.statValue}>{formatNumber(overview?.totalRequests ?? 0)}</strong>
        </article>
        <article className={designStyles.statCard}>
          <span className={designStyles.statLabel}>{t('usage_stats.sub2api_total_tokens')}</span>
          <strong className={designStyles.statValue}>{formatNumber(overview?.totalTokens ?? 0)}</strong>
        </article>
        <article className={designStyles.statCard}>
          <span className={designStyles.statLabel}>{t('usage_stats.sub2api_total_cost')}</span>
          <strong className={designStyles.statValue}>{formatCost(overview?.actualCost ?? 0)}</strong>
        </article>
        <article className={designStyles.statCard}>
          <span className={designStyles.statLabel}>{t('usage_stats.sub2api_active_users')}</span>
          <strong className={designStyles.statValue}>{formatNumber(overview?.activeUsers ?? 0)}</strong>
        </article>
        <article className={designStyles.statCard}>
          <span className={designStyles.statLabel}>{t('usage_stats.sub2api_available_accounts')}</span>
          <strong className={designStyles.statValue}>{formatNumber(availableAccounts)} / {formatNumber(quotaAccounts.length)}</strong>
        </article>
      </section>

      {points.length > 1 && tokenChartData && (
        <section className={designStyles.overviewSurface} aria-label="Token usage chart">
          <div className={designStyles.sectionTitleBlock}>
            <span className={designStyles.sectionEyebrow}>{t('usage_stats.sub2api_overview_title')}</span>
            <h3 className={designStyles.sectionTitle}>{t('usage_stats.sub2api_tokens')} — {t('usage_stats.input_tokens')} / {t('usage_stats.output_tokens')} / {t('usage_stats.cached_tokens')}</h3>
            <p className={designStyles.sectionSubtitle}>Token consumption over time</p>
          </div>
          <div className={styles.overviewChartStacked}>
            <div className={styles.chartLegendRow}>
              <span className={styles.chartLegendDot} style={{ background: '#3b82f6' }} /> Input
              <span className={styles.chartLegendDot} style={{ background: '#22c55e', marginLeft: 14 }} /> Output
              <span className={styles.chartLegendDot} style={{ background: '#f59e0b', marginLeft: 14 }} /> Cache
            </div>
            <div className={styles.overviewChartArea}>
              <Line data={tokenChartData} options={chartOptions} />
            </div>
          </div>
          
          <div style={{ marginTop: '24px' }}>
            <RequestHealthTimelineCard usage={usage ?? null} loading={!!loading} />
          </div>
        </section>
      )}

      {points.length > 1 && (requestChartData || costChartData || usersChartData || tokensChartData) && (
        <div className={styles.overviewChartGrid}>
          {requestChartData && (
            <section className={styles.overviewChartCard} aria-label="Request volume chart">
              <div className={styles.chartCardHeader}>
                <strong className={styles.chartCardTitle}>{t('usage_stats.sub2api_total_requests')}</strong>
                <span className={styles.chartCardHint}>per hour</span>
              </div>
              <div className={styles.overviewChartAreaShort}>
                <Bar data={requestChartData} options={barOptions} />
              </div>
            </section>
          )}
          {costChartData && (
            <section className={styles.overviewChartCard} aria-label="Cost chart">
              <div className={styles.chartCardHeader}>
                <strong className={styles.chartCardTitle}>{t('usage_stats.sub2api_total_cost')}</strong>
                <span className={styles.chartCardHint}>per hour</span>
              </div>
              <div className={styles.overviewChartAreaShort}>
                <Line data={costChartData} options={chartOptions} />
              </div>
            </section>
          )}
          {usersChartData && (
            <section className={styles.overviewChartCard} aria-label="Active users chart">
              <div className={styles.chartCardHeader}>
                <strong className={styles.chartCardTitle}>{t('usage_stats.sub2api_active_users')}</strong>
                <span className={styles.chartCardHint}>per hour</span>
              </div>
              <div className={styles.overviewChartAreaShort}>
                <Line data={usersChartData} options={chartOptions} />
              </div>
            </section>
          )}
          {tokensChartData && (
            <section className={styles.overviewChartCard} aria-label="Total tokens chart">
              <div className={styles.chartCardHeader}>
                <strong className={styles.chartCardTitle}>{t('usage_stats.sub2api_total_tokens')}</strong>
                <span className={styles.chartCardHint}>per hour</span>
              </div>
              <div className={styles.overviewChartAreaShort}>
                <Line data={tokensChartData} options={chartOptions} />
              </div>
            </section>
          )}

        </div>
      )}

      <section className={styles.overviewSurface}>
        <div className={styles.sectionTitleBlock}>
          <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_models_eyebrow')}</span>
          <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_top_models')}</h3>
          <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_top_models_hint')}</p>
        </div>
        {topModels.length === 0 ? (
          <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
        ) : (
          <div className={styles.modelSummaryList}>
            {topModels.map((model) => {
              const totalTokens = modelTokenTotal(model)
              return (
                <article key={`${model.model}-${model.requestedModel}-${model.upstreamModel}`} className={styles.modelSummaryItem}>
                  <div className={styles.modelSummaryHeader}>
                    <strong className={styles.modelSummaryName}>{model.model || model.requestedModel || model.upstreamModel || 'unknown'}</strong>
                    <span className={styles.modelSummaryCost}>{formatCost(model.actualCost)}</span>
                  </div>
                  <div className={styles.rankingMeterShell} aria-hidden="true">
                    <span className={styles.rankingMeterFill} style={{ width: `${Math.max((totalTokens / maxModelTokens) * 100, 3)}%` }} />
                  </div>
                  <div className={styles.rankingMetrics}>
                    <span>{t('usage_stats.sub2api_requests')}: {formatNumber(model.totalRequests)}</span>
                    <span>{t('usage_stats.sub2api_tokens')}: {formatNumber(totalTokens)}</span>
                  </div>
                </article>
              )
            })}
          </div>
        )}
      </section>
    </div>
  )
}
