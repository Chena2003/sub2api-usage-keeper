import { useTranslation } from 'react-i18next'
import type { Sub2ApiRanking, Sub2ApiRankingDimension } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'

type TokenRankingCardProps = {
  rankings: Sub2ApiRanking[]
  dimension: Sub2ApiRankingDimension
  onDimensionChange: (dimension: Sub2ApiRankingDimension) => void | Promise<void>
}

const DIMENSIONS: Sub2ApiRankingDimension[] = ['user', 'api_key', 'model', 'account']

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatCost = (value: number) => `$${value.toFixed(4)}`

const formatShare = (value: number) => `${(value * 100).toFixed(1)}%`

export function TokenRankingCard({ rankings, dimension, onDimensionChange }: TokenRankingCardProps) {
  const { t } = useTranslation()
  const maxTokens = Math.max(...rankings.map((ranking) => ranking.totalTokens), 1)

  return (
    <section className={styles.rankingSurface}>
      <div className={styles.sectionTitleBlock}>
        <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_ranking_eyebrow')}</span>
        <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_ranking_title')}</h3>
        <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_ranking_hint')}</p>
      </div>

      <div className={styles.refreshSwitcher} role="tablist" aria-label={t('usage_stats.sub2api_ranking_dimension')}>
        {DIMENSIONS.map((option) => (
          <button
            key={option}
            type="button"
            role="tab"
            aria-selected={dimension === option}
            className={`${styles.refreshPill} ${dimension === option ? styles.refreshPillActive : ''}`.trim()}
            onClick={() => void onDimensionChange(option)}
          >
            {t(`usage_stats.sub2api_dimension_${option}`)}
          </button>
        ))}
      </div>

      {rankings.length === 0 ? (
        <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.rankingList}>
          {rankings.map((ranking, index) => (
            <article key={`${ranking.dimension}-${ranking.name}-${index}`} className={styles.rankingItem}>
              <div className={styles.rankingHeader}>
                <span className={styles.rankingRank}>#{index + 1}</span>
                <strong className={styles.rankingName}>{ranking.name || 'unknown'}</strong>
                <span className={styles.rankingShare}>{formatShare(ranking.share)}</span>
              </div>
              <div className={styles.rankingMeterShell} aria-hidden="true">
                <span className={styles.rankingMeterFill} style={{ width: `${Math.max((ranking.totalTokens / maxTokens) * 100, 3)}%` }} />
              </div>
              <div className={styles.rankingMetrics}>
                <span>{t('usage_stats.sub2api_requests')}: {formatNumber(ranking.totalRequests)}</span>
                <span>{t('usage_stats.sub2api_total_tokens')}: {formatNumber(ranking.totalTokens)}</span>
                <span>{t('usage_stats.sub2api_input_tokens')}: {formatNumber(ranking.inputTokens)}</span>
                <span>{t('usage_stats.sub2api_output_tokens')}: {formatNumber(ranking.outputTokens)}</span>
                <span>{t('usage_stats.sub2api_cache_tokens')}: {formatNumber(ranking.cacheTokens)}</span>
                <span>{t('usage_stats.sub2api_cost')}: {formatCost(ranking.actualCost)}</span>
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
