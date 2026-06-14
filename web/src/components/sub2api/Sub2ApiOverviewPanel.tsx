import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Line, Bar } from 'react-chartjs-2'
import type { ChartData, ChartOptions } from 'chart.js'
import type { Sub2ApiAccount, Sub2ApiOverview, Sub2ApiTimeseriesPoint } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import { Kpi } from '@/components/ui/Kpi'
import { StatusPill } from '@/components/ui/StatusPill'
import styles from './Sub2ApiOverviewPanel.module.scss'
import { RequestHealthTimelineCard } from './RequestHealthTimelineCard'
import type { UsageOverviewPayload } from '@/components/usage/hooks/useUsageData'

type Sub2ApiOverviewPanelProps = {
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  quotaAccounts: Sub2ApiAccount[]
  usage?: UsageOverviewPayload | null
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatCost = (value: number) => `$${value.toFixed(4)}`

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

export function Sub2ApiOverviewPanel({ overview, points, quotaAccounts, usage, loading, error, onRetry }: Sub2ApiOverviewPanelProps) {
  const { t } = useTranslation()
  const hasData = Boolean(overview) || points.length > 0 || quotaAccounts.length > 0
  const quotaRiskCount = quotaAccounts.filter(isQuotaAtRisk).length
  const availableAccounts = quotaAccounts.length - quotaRiskCount

  const fmtHour = (s: string) => {
    const d = new Date(s)
    return Number.isNaN(d.getTime()) ? s : `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }

  const chartLabels = useMemo(() => points.map((p) => fmtHour(p.bucketStart)), [points])

  const tokenChartData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Input',
          data: points.map((p) => p.inputTokens),
          borderColor: 'oklch(62% 0.15 250)',
          backgroundColor: 'oklch(62% 0.15 250 / 0.10)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
        },
        {
          label: 'Output',
          data: points.map((p) => p.outputTokens),
          borderColor: 'oklch(58% 0.16 145)',
          backgroundColor: 'oklch(58% 0.16 145 / 0.10)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
        },
        {
          label: 'Cache',
          data: points.map((p) => p.cacheCreationTokens + p.cacheReadTokens),
          borderColor: 'oklch(72% 0.17 82)',
          backgroundColor: 'oklch(72% 0.17 82 / 0.08)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
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
          backgroundColor: 'oklch(58% 0.16 145 / 0.22)',
          borderColor: 'oklch(58% 0.16 145)',
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
          borderColor: 'oklch(72% 0.17 82)',
          backgroundColor: 'oklch(72% 0.17 82 / 0.08)',
          fill: true, tension: 0.35, pointRadius: 0, pointHitRadius: 8, borderWidth: 2,
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
          borderColor: 'oklch(65% 0.14 200)',
          backgroundColor: 'oklch(65% 0.14 200 / 0.08)',
          fill: true, tension: 0.35, pointRadius: 0, pointHitRadius: 8, borderWidth: 2,
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
          borderColor: 'oklch(55% 0.01 240)',
          backgroundColor: 'oklch(55% 0.01 240 / 0.08)',
          fill: true, tension: 0.35, pointRadius: 0, pointHitRadius: 8, borderWidth: 2,
        },
      ],
    }
  }, [points, chartLabels])

  const chartOptions: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false }, tooltip: { enabled: true, mode: 'index', intersect: false } },
    scales: {
      x: { display: true, grid: { display: false }, ticks: { color: 'oklch(50% 0.018 240)', font: { size: 10, family: 'inherit' }, maxTicksLimit: 8 } },
      y: { display: true, grid: { color: 'oklch(90% 0.008 240 / 0.6)', drawTicks: false }, ticks: { color: 'oklch(50% 0.018 240)', font: { size: 10, family: 'inherit' }, maxTicksLimit: 5 } },
    },
    interaction: { mode: 'index', intersect: false },
  }

  const barOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false }, tooltip: { enabled: true, mode: 'index', intersect: false } },
    scales: {
      x: { display: true, grid: { display: false }, ticks: { color: 'oklch(50% 0.018 240)', font: { size: 10, family: 'inherit' }, maxTicksLimit: 8 } },
      y: { display: true, grid: { color: 'oklch(90% 0.008 240 / 0.6)', drawTicks: false }, beginAtZero: true, ticks: { color: 'oklch(50% 0.018 240)', font: { size: 10, family: 'inherit' }, maxTicksLimit: 5 } },
    },
    interaction: { mode: 'index', intersect: false },
  }

  if (!hasData && !loading) {
    return (
      <Panel eyebrow={t('usage_stats.sub2api_overview_title')} aria-label={t('usage_stats.sub2api_overview_title')}>
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      </Panel>
    )
  }

  return (
    <div className={styles.bentoRoot} aria-label={t('usage_stats.sub2api_overview_title')}>

      {error && <Panel.Error message={error} onRetry={onRetry} />}

      {/* ── KPI row (5 cards) ───────────────────────────────── */}
      <section className={styles.kpiRow} aria-label="Key metrics">
        <article className={styles.kpiCard}>
          <Kpi
            label={t('usage_stats.sub2api_total_requests')}
            value={overview ? formatNumber(overview.totalRequests) : '—'}
            loading={loading && !overview}
            empty={!overview && !loading}
          />
        </article>
        <article className={styles.kpiCard}>
          <Kpi
            label={t('usage_stats.sub2api_total_tokens')}
            value={overview ? formatNumber(overview.totalTokens) : '—'}
            loading={loading && !overview}
            empty={!overview && !loading}
          />
        </article>
        <article className={styles.kpiCard}>
          <Kpi
            label={t('usage_stats.sub2api_total_cost')}
            value={overview ? formatCost(overview.actualCost) : '—'}
            loading={loading && !overview}
            empty={!overview && !loading}
          />
        </article>
        <article className={styles.kpiCard}>
          <Kpi
            label={t('usage_stats.sub2api_active_users')}
            value={overview ? formatNumber(overview.activeUsers) : '—'}
            loading={loading && !overview}
            empty={!overview && !loading}
          />
        </article>
        <article className={styles.kpiCard}>
          <Kpi
            label={t('usage_stats.sub2api_available_accounts')}
            value={`${formatNumber(availableAccounts)} / ${formatNumber(quotaAccounts.length)}`}
            loading={loading && !quotaAccounts.length}
            empty={!overview && !loading}
          />
          {quotaRiskCount > 0 && (
            <div className={styles.kpiPill}>
              <StatusPill variant="warn" label={`${quotaRiskCount} at risk`} />
            </div>
          )}
        </article>
      </section>

      {/* ── Full-width: Token timeline ───────────────────────── */}
      {(points.length > 1 || loading) && (
        <Panel
          eyebrow={t('usage_stats.sub2api_overview_title')}
          title={`${t('usage_stats.sub2api_tokens')} — ${t('usage_stats.input_tokens')} / ${t('usage_stats.output_tokens')} / ${t('usage_stats.cached_tokens')}`}
          as="section"
        >
          {loading && !tokenChartData ? (
            <Panel.Loading rows={4} />
          ) : tokenChartData ? (
            <>
              <div className={styles.chartLegend}>
                <span className={styles.legendDot} style={{ background: 'oklch(62% 0.15 250)' }} />
                <span className={styles.legendLabel}>Input</span>
                <span className={styles.legendDot} style={{ background: 'oklch(58% 0.16 145)' }} />
                <span className={styles.legendLabel}>Output</span>
                <span className={styles.legendDot} style={{ background: 'oklch(72% 0.17 82)' }} />
                <span className={styles.legendLabel}>Cache</span>
              </div>
              <div className={styles.chartAreaFull} aria-label="Token 趋势图">
                <Line data={tokenChartData} options={chartOptions} />
              </div>
            </>
          ) : (
            <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
          )}
        </Panel>
      )}

      {/* ── Full-width: Request Health Timeline ─────────────── */}
      <div className={styles.healthPanel}>
        <RequestHealthTimelineCard usage={usage ?? null} loading={!!loading} />
      </div>

      {/* ── Mini charts row (4 cols) ─────────────────────────── */}
      {points.length > 1 && (
        <div className={styles.miniChartsRow}>
          {requestChartData && (
            <Panel as="section">
              <div className={styles.miniChartHeader}>
                <strong className={styles.miniChartTitle}>{t('usage_stats.sub2api_total_requests')}</strong>
                <span className={styles.miniChartHint}>per hour</span>
              </div>
              <div className={styles.miniChartArea} aria-label="请求量趋势图">
                <Bar data={requestChartData} options={barOptions} />
              </div>
            </Panel>
          )}
          {costChartData && (
            <Panel as="section">
              <div className={styles.miniChartHeader}>
                <strong className={styles.miniChartTitle}>{t('usage_stats.sub2api_total_cost')}</strong>
                <span className={styles.miniChartHint}>per hour</span>
              </div>
              <div className={styles.miniChartArea} aria-label="成本趋势图">
                <Line data={costChartData} options={chartOptions} />
              </div>
            </Panel>
          )}
          {usersChartData && (
            <Panel as="section">
              <div className={styles.miniChartHeader}>
                <strong className={styles.miniChartTitle}>{t('usage_stats.sub2api_active_users')}</strong>
                <span className={styles.miniChartHint}>per hour</span>
              </div>
              <div className={styles.miniChartArea} aria-label="活跃用户趋势图">
                <Line data={usersChartData} options={chartOptions} />
              </div>
            </Panel>
          )}
          {tokensChartData && (
            <Panel as="section">
              <div className={styles.miniChartHeader}>
                <strong className={styles.miniChartTitle}>{t('usage_stats.sub2api_total_tokens')}</strong>
                <span className={styles.miniChartHint}>per hour</span>
              </div>
              <div className={styles.miniChartArea} aria-label="Total tokens chart">
                <Line data={tokensChartData} options={chartOptions} />
              </div>
            </Panel>
          )}
        </div>
      )}

    </div>
  )
}
