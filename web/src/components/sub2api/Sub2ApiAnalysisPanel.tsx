import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Doughnut, Bar, Line } from 'react-chartjs-2'
import type { ChartData, ChartOptions } from 'chart.js'
import type { Sub2ApiModelUsage, Sub2ApiTimeseriesPoint } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import styles from './Sub2ApiAnalysisPanel.module.scss'

type Sub2ApiAnalysisPanelProps = {
  models: Sub2ApiModelUsage[]
  points?: Sub2ApiTimeseriesPoint[]
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)
const formatCost = (value: number) => `$${value.toFixed(4)}`

const modelTokenTotal = (model: Sub2ApiModelUsage) => (
  model.inputTokens + model.outputTokens + model.cacheCreationTokens + model.cacheReadTokens
)

// Distinct OKLCH palette — spaced across hue wheel for clear doughnut segments
const PALETTE = [
  'oklch(58% 0.16 145)',   // green  (--ok)
  'oklch(62% 0.20 260)',   // blue
  'oklch(72% 0.17 82)',    // amber  (--warn)
  'oklch(62% 0.20 330)',   // pink/purple
  'oklch(55% 0.18 28)',    // red-orange (--danger toned)
  'oklch(65% 0.15 200)',   // cyan
  'oklch(60% 0.15 300)',   // violet
  'oklch(68% 0.14 60)',    // yellow
]

export function Sub2ApiAnalysisPanel({ models, points = [], loading, error, onRetry }: Sub2ApiAnalysisPanelProps) {
  const { t } = useTranslation()
  const topModels = models.slice(0, 8)

  const fmtHour = (s: string) => {
    const d = new Date(s)
    return Number.isNaN(d.getTime()) ? s : `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }

  const chartLabels = useMemo(() => points.map((p) => fmtHour(p.bucketStart)), [points])

  const modelChartData = useMemo((): ChartData<'doughnut'> | null => {
    if (topModels.length === 0) return null
    return {
      labels: topModels.map((m) => m.model || m.requestedModel || 'unknown'),
      datasets: [{
        data: topModels.map(modelTokenTotal),
        backgroundColor: PALETTE.slice(0, topModels.length),
        borderWidth: 0,
        hoverOffset: 10,
      }],
    }
  }, [topModels])

  const costChartData = useMemo((): ChartData<'bar'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [{
        label: t('usage_stats.sub2api_total_cost'),
        data: points.map((p) => p.actualCost),
        backgroundColor: 'oklch(72% 0.17 82 / 0.25)',
        borderColor: 'oklch(72% 0.17 82)',
        borderWidth: 1,
        borderRadius: 4,
      }],
    }
  }, [points, chartLabels, t])

  const tokenTypeData = useMemo((): ChartData<'line'> | null => {
    if (points.length === 0) return null
    return {
      labels: chartLabels,
      datasets: [
        {
          label: 'Input',
          data: points.map((p) => p.inputTokens),
          borderColor: 'oklch(62% 0.20 260)',
          backgroundColor: 'oklch(62% 0.20 260 / 0.08)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
        },
        {
          label: 'Output',
          data: points.map((p) => p.outputTokens),
          borderColor: 'oklch(58% 0.16 145)',
          backgroundColor: 'oklch(58% 0.16 145 / 0.08)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
        },
        {
          label: 'Cache',
          data: points.map((p) => p.cacheCreationTokens + p.cacheReadTokens),
          borderColor: 'oklch(65% 0.15 200)',
          backgroundColor: 'oklch(65% 0.15 200 / 0.06)',
          fill: true, tension: 0.25, pointRadius: 0, pointHitRadius: 8, borderWidth: 1.5,
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

  const doughnutOptions: ChartOptions<'doughnut'> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: '62%',
    plugins: { legend: { display: false }, tooltip: { enabled: true } },
  }

  return (
    <div className={styles.analysisRoot}>

      {error && <Panel.Error message={error} onRetry={onRetry} />}

      {/* ── Row 1: Model Mix — doughnut + legend, full width ── */}
      <Panel
        eyebrow={t('usage_stats.sub2api_models_eyebrow')}
        title="Model Mix"
        as="section"
      >
        {loading ? (
          <Panel.Loading rows={5} />
        ) : models.length === 0 ? (
          <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
        ) : (
          <div className={styles.mixRow}>
            {modelChartData && (
              <div className={styles.doughnutWrap} aria-label="模型分布饼图">
                <Doughnut data={modelChartData} options={doughnutOptions} />
              </div>
            )}
            <div className={styles.mixLegend}>
              {topModels.map((m, i) => {
                const total = modelTokenTotal(m)
                const totalAll = topModels.reduce((s, x) => s + modelTokenTotal(x), 0)
                const pct = totalAll > 0 ? ((total / totalAll) * 100).toFixed(1) : '0.0'
                return (
                  <div key={m.model || m.requestedModel} className={styles.mixLegendItem}>
                    <span className={styles.mixLegendDot} style={{ background: PALETTE[i % PALETTE.length] }} />
                    <span className={styles.mixLegendName}>{m.model || m.requestedModel || 'unknown'}</span>
                    <div className={styles.mixMeter}>
                      <span className={styles.mixMeterFill} style={{ width: `${Math.max(Number(pct), 2)}%`, background: PALETTE[i % PALETTE.length] }} />
                    </div>
                    <span className={styles.mixLegendPct}>{pct}%</span>
                    <span className={styles.mixLegendTokens}>{formatNumber(total)}</span>
                    <span className={styles.mixLegendCost}>{formatCost(m.actualCost)}</span>
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </Panel>

      {/* ── Row 2: Token type breakdown (full width) ────────── */}
      {(points.length > 1 || loading) && (
        <Panel
          eyebrow="Tokens"
          title="Input / Output / Cache Breakdown"
          as="section"
        >
          {loading && !tokenTypeData ? (
            <Panel.Loading rows={4} />
          ) : tokenTypeData ? (
            <>
              <div className={styles.chartLegend}>
                <span className={styles.legendDotInline} style={{ background: 'oklch(62% 0.20 260)' }} />
                <span className={styles.legendLabelInline}>Input</span>
                <span className={styles.legendDotInline} style={{ background: 'oklch(58% 0.16 145)' }} />
                <span className={styles.legendLabelInline}>Output</span>
                <span className={styles.legendDotInline} style={{ background: 'oklch(65% 0.15 200)' }} />
                <span className={styles.legendLabelInline}>Cache</span>
              </div>
              <div className={styles.chartAreaFull} aria-label="Token 趋势折线图">
                <Line data={tokenTypeData} options={chartOptions} />
              </div>
            </>
          ) : (
            <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
          )}
        </Panel>
      )}

      {/* ── Row 3: Cost over time (full width) ──────────────── */}
      {(points.length > 1 || loading) && (
        <Panel
          eyebrow="Cost"
          title="Cost per Hour"
          as="section"
        >
          {loading && !costChartData ? (
            <Panel.Loading rows={3} />
          ) : costChartData ? (
            <div className={styles.chartAreaMid} aria-label="成本柱状图">
              <Bar data={costChartData} options={barOptions} />
            </div>
          ) : (
            <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
          )}
        </Panel>
      )}

      {/* ── Model detail table ───────────────────────────────── */}
      <Panel
        eyebrow={t('usage_stats.sub2api_models_eyebrow')}
        title="Model Detail"
        as="section"
      >
        {loading ? (
          <Panel.Loading rows={5} />
        ) : models.length === 0 ? (
          <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
        ) : (
          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Requests</th>
                  <th>Input Tokens</th>
                  <th>Output Tokens</th>
                  <th>Cache Tokens</th>
                  <th>Total Tokens</th>
                  <th>Cost</th>
                </tr>
              </thead>
              <tbody>
                {models.map((model) => (
                  <tr key={model.model || model.requestedModel}>
                    <td className={styles.tdName}>{model.model || model.requestedModel || 'unknown'}</td>
                    <td className={styles.mono}>{formatNumber(model.totalRequests)}</td>
                    <td className={styles.mono}>{formatNumber(model.inputTokens)}</td>
                    <td className={styles.mono}>{formatNumber(model.outputTokens)}</td>
                    <td className={styles.mono}>{formatNumber(model.cacheCreationTokens + model.cacheReadTokens)}</td>
                    <td className={styles.mono}>{formatNumber(modelTokenTotal(model))}</td>
                    <td className={`${styles.mono} ${styles.cost}`}>{formatCost(model.actualCost)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>

    </div>
  )
}
