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

  return (
    <section className="card">
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
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t('usage_stats.sub2api_rank')}</th>
                <th>{t('usage_stats.sub2api_name')}</th>
                <th>{t('usage_stats.sub2api_requests')}</th>
                <th>{t('usage_stats.sub2api_input_tokens')}</th>
                <th>{t('usage_stats.sub2api_output_tokens')}</th>
                <th>{t('usage_stats.sub2api_cache_tokens')}</th>
                <th>{t('usage_stats.sub2api_total_tokens')}</th>
                <th>{t('usage_stats.sub2api_cost')}</th>
                <th>{t('usage_stats.sub2api_share')}</th>
              </tr>
            </thead>
            <tbody>
              {rankings.map((ranking, index) => (
                <tr key={`${ranking.dimension}-${ranking.name}-${index}`}>
                  <td>{index + 1}</td>
                  <td className={styles.modelCell}>{ranking.name || 'unknown'}</td>
                  <td>{formatNumber(ranking.totalRequests)}</td>
                  <td>{formatNumber(ranking.inputTokens)}</td>
                  <td>{formatNumber(ranking.outputTokens)}</td>
                  <td>{formatNumber(ranking.cacheTokens)}</td>
                  <td>{formatNumber(ranking.totalTokens)}</td>
                  <td>{formatCost(ranking.actualCost)}</td>
                  <td>{formatShare(ranking.share)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
