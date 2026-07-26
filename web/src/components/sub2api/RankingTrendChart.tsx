import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { ChartData, ChartOptions } from 'chart.js'
import { Line } from 'react-chartjs-2'
import { useThemeStore } from '@/stores'
import type { Sub2ApiRankingTrendResponse } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import styles from './RankingTrendChart.module.scss'

const TREND_COLORS = [
  '#8b8680', '#8b5cf6', '#22c55e', '#f97316', '#f59e0b',
  '#06b6d4', '#ef4444', '#6366f1', '#ec4899', '#a855f7',
  '#14b8a6', '#84cc16',
]

// FNV-1a hash, deterministic across renders regardless of insertion order.
function hashStringToIndex(value: string, modulo: number): number {
  let hash = 0x811c9dc5
  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i)
    hash = Math.imul(hash, 0x01000193)
  }
  return Math.abs(hash) % modulo
}

// Assigns each name a color derived from a hash of the name itself (not its
// position), so a series keeps its color across refreshes even if the
// backend ranking order changes. Names are processed in a fixed (sorted)
// order so the outcome doesn't depend on the order `names` is passed in.
// On a hash collision, the next free color in `colors` is used instead, so
// colors stay distinct within a single render.
export function assignSeriesColors(names: string[], colors: string[] = TREND_COLORS): Map<string, string> {
  const used = new Set<number>()
  const assignments = new Map<string, string>()
  for (const name of [...names].sort()) {
    let index = hashStringToIndex(name, colors.length)
    if (used.has(index)) {
      let candidate = (index + 1) % colors.length
      while (used.has(candidate) && candidate !== index) {
        candidate = (candidate + 1) % colors.length
      }
      index = candidate
    }
    used.add(index)
    assignments.set(name, colors[index])
  }
  return assignments
}

type Props = {
  data: Sub2ApiRankingTrendResponse | null
  loading?: boolean
}

const formatTokensCompact = (n: number): string => {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

export function RankingTrendChart({ data, loading }: Props) {
  const { t } = useTranslation()
  const theme = useThemeStore((s) => s.theme)
  const isDark = theme === 'dark'

  const { chartData, chartOptions } = useMemo(() => {
    if (!data || data.points.length === 0) {
      return { chartData: null, chartOptions: null }
    }

    // Group points by name
    const byName = new Map<string, Map<string, number>>()
    const allBuckets = new Set<string>()
    for (const pt of data.points) {
      allBuckets.add(pt.bucket)
      let map = byName.get(pt.name)
      if (!map) {
        map = new Map()
        byName.set(pt.name, map)
      }
      map.set(pt.bucket, pt.tokens)
    }

    const labels = [...allBuckets].sort()
    const colorByName = assignSeriesColors([...byName.keys()])
    const datasets = [...byName.entries()].map(([name, values]) => ({
      label: name,
      data: labels.map((b) => values.get(b) ?? 0),
      borderColor: colorByName.get(name)!,
      backgroundColor: `${colorByName.get(name)}18`,
      fill: false,
      tension: 0.3,
      pointRadius: labels.length > 30 ? 0 : 3,
      pointHoverRadius: 4,
      borderWidth: 2,
    }))

    const gridColor = isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(17, 24, 39, 0.06)'
    const axisBorderColor = isDark ? 'rgba(255, 255, 255, 0.10)' : 'rgba(17, 24, 39, 0.10)'
    const tickColor = isDark ? 'rgba(255, 255, 255, 0.72)' : 'rgba(17, 24, 39, 0.72)'
    const tooltipBg = isDark ? 'rgba(17, 24, 39, 0.92)' : 'rgba(255, 255, 255, 0.98)'
    const tooltipTitle = isDark ? '#ffffff' : '#111827'
    const tooltipBody = isDark ? 'rgba(255, 255, 255, 0.86)' : '#374151'
    const tooltipBorder = isDark ? 'rgba(255, 255, 255, 0.10)' : 'rgba(17, 24, 39, 0.10)'

    const displayLabels = labels.map((l) => {
      if (data.granularity === 'hour') {
        const parts = l.split(' ')
        return parts[1] ?? l
      }
      const parts = l.split('-')
      return parts.length === 3 ? `${parts[1]}-${parts[2]}` : l
    })

    const options: ChartOptions<'line'> = {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: 'index', intersect: false },
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: tooltipBg,
          titleColor: tooltipTitle,
          bodyColor: tooltipBody,
          borderColor: tooltipBorder,
          borderWidth: 1,
          padding: 10,
          displayColors: true,
          usePointStyle: true,
          itemSort: (a, b) => (b.parsed?.y ?? 0) - (a.parsed?.y ?? 0),
          callbacks: {
            label: (ctx) => {
              const label = ctx.dataset.label ?? ''
              return `${label}: ${formatTokensCompact(ctx.parsed.y ?? 0)}`
            },
          },
        },
      },
      scales: {
        x: {
          grid: { color: gridColor, drawTicks: false },
          border: { color: axisBorderColor },
          ticks: {
            color: tickColor,
            font: { size: 11 },
            maxRotation: 0,
            autoSkip: true,
            maxTicksLimit: 12,
          },
        },
        y: {
          beginAtZero: true,
          grid: { color: gridColor },
          border: { color: axisBorderColor },
          ticks: {
            color: tickColor,
            font: { size: 11 },
            callback: (value) => formatTokensCompact(Number(value)),
          },
        },
      },
      elements: {
        line: { tension: 0.3, borderWidth: 2 },
        point: { borderWidth: 2 },
      },
    }

    const cd: ChartData<'line'> = { labels: displayLabels, datasets }
    return { chartData: cd, chartOptions: options }
  }, [data, isDark])

  return (
    <Panel
      eyebrow={t('usage_stats.sub2api_ranking_trend_eyebrow')}
      title={t('usage_stats.sub2api_ranking_trend_title')}
      as="section"
    >
      {loading ? (
        <Panel.Loading rows={3} />
      ) : !chartData || chartData.datasets.length === 0 ? (
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      ) : (
        <div className={styles.wrapper}>
          <div className={styles.legend}>
            {chartData.datasets.map((ds, i) => (
              <div key={`${ds.label}-${i}`} className={styles.legendItem}>
                <span className={styles.legendDot} style={{ backgroundColor: ds.borderColor as string }} />
                <span className={styles.legendLabel}>{ds.label}</span>
              </div>
            ))}
          </div>
          <div className={styles.chartArea}>
            <Line data={chartData} options={chartOptions!} />
          </div>
        </div>
      )}
    </Panel>
  )
}
