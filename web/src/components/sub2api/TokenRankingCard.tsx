import { useTranslation } from 'react-i18next'
import type { Sub2ApiRanking, Sub2ApiRankingDimension } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import styles from './TokenRankingCard.module.scss'

type TokenRankingCardProps = {
  rankings: Sub2ApiRanking[]
  dimension: Sub2ApiRankingDimension
  onDimensionChange: (dimension: Sub2ApiRankingDimension) => void | Promise<void>
  loading?: boolean
  error?: string | null
  onRetry?: () => void
}

const DIMENSIONS: Sub2ApiRankingDimension[] = ['user', 'api_key', 'model', 'account']

const RANK_COLORS = [
  'oklch(72% 0.17 82)',   // #1 amber
  'oklch(62% 0.15 250)',  // #2 blue
  'oklch(58% 0.16 145)',  // #3 green
]

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)
const formatCost   = (value: number) => `$${value.toFixed(4)}`

const formatTokensCompact = (n: number) => {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

export function TokenRankingCard({ rankings, dimension, onDimensionChange, loading, error, onRetry }: TokenRankingCardProps) {
  const { t } = useTranslation()
  const maxTokens = Math.max(...rankings.map((r) => r.totalTokens), 1)
  const totalTokens = rankings.reduce((s, r) => s + r.totalTokens, 0)

  const dimSelector = (
    <div className={styles.dimSwitcher} role="tablist" aria-label={t('usage_stats.sub2api_ranking_dimension')}>
      {DIMENSIONS.map((opt) => (
        <button
          key={opt}
          type="button"
          role="tab"
          aria-selected={dimension === opt}
          className={`${styles.dimPill} ${dimension === opt ? styles.dimPillActive : ''}`}
          onClick={() => void onDimensionChange(opt)}
        >
          {t(`usage_stats.sub2api_dimension_${opt}`)}
        </button>
      ))}
    </div>
  )

  return (
    <Panel
      eyebrow={t('usage_stats.sub2api_ranking_eyebrow')}
      title={t('usage_stats.sub2api_ranking_title')}
      actions={dimSelector}
      as="section"
    >
      {error ? (
        <Panel.Error message={error} onRetry={onRetry} />
      ) : loading ? (
        <Panel.Loading rows={5} />
      ) : rankings.length === 0 ? (
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      ) : (
        <div className={styles.list}>
          {rankings.map((r, i) => {
            const rankColor = i < 3 ? RANK_COLORS[i] : 'var(--muted)'
            const pct = totalTokens > 0 ? ((r.totalTokens / totalTokens) * 100).toFixed(1) : '0.0'
            return (
              <article key={`${r.dimension}-${r.name}-${i}`} className={styles.item}>
                <div className={styles.rankBadge} style={{ '--rank-color': rankColor } as React.CSSProperties}>
                  {i + 1}
                </div>
                <div className={styles.body}>
                  <div className={styles.topRow}>
                    <strong className={styles.name}>{r.name || 'unknown'}</strong>
                    <span className={styles.tokenCompact}>{formatTokensCompact(r.totalTokens)}</span>
                  </div>
                  <div className={styles.meterRow}>
                    <div className={styles.meter} aria-hidden="true">
                      <span
                        className={styles.meterFill}
                        style={{
                          width: `${Math.max((r.totalTokens / maxTokens) * 100, 2)}%`,
                          background: rankColor,
                        }}
                      />
                    </div>
                    <span className={styles.pct}>{pct}%</span>
                  </div>
                  <div className={styles.metaRow}>
                    <span className={styles.metaItem}>
                      <span className={styles.metaLabel}>{t('usage_stats.sub2api_requests')}</span>
                      <span className={styles.mono}>{formatNumber(r.totalRequests)}</span>
                    </span>
                    <span className={styles.metaItem}>
                      <span className={styles.metaLabel}>{t('usage_stats.sub2api_cost')}</span>
                      <span className={styles.mono}>{formatCost(r.actualCost)}</span>
                    </span>
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      )}
    </Panel>
  )
}
