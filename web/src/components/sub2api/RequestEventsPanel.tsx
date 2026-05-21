import { useTranslation } from 'react-i18next'
import type { Sub2ApiEventsResponse } from '@/lib/sub2apiTypes'
import { formatDurationMs } from '@/utils/usage'
import styles from './Sub2ApiDesign.module.scss'

type RequestEventsPanelProps = {
  events: Sub2ApiEventsResponse
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatDateTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value || '-'
  return date.toLocaleString()
}

const getBadgeClass = (status: string) => {
  const s = status.toLowerCase()
  const statusCode = Number(s)
  if (Number.isInteger(statusCode)) {
    if (statusCode >= 200 && statusCode < 400) return styles.badgeSuccess
    if (statusCode >= 400) return styles.badgeError
  }
  if (s === 'success') return styles.badgeSuccess
  if (s.includes('fail') || s.includes('error') || s.includes('timeout')) return styles.badgeError
  return styles.badgeWarning
}

export function RequestEventsPanel({ events }: RequestEventsPanelProps) {
  const { t } = useTranslation()

  return (
    <section className={styles.requestEventsCard}>
      <div className={styles.panelHeader}>
        <span className={styles.eyebrow}>{t('usage_stats.sub2api_events_eyebrow')}</span>
        <h3 className={styles.title}>{t('usage_stats.sub2api_events_title')}</h3>
        <p className={styles.subtitle}>{t('usage_stats.sub2api_events_hint')}</p>
      </div>

      {events.events.length === 0 ? (
        <div className={styles.emptyState}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.eventTable}>
          <div className={styles.eventGrid}>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_timestamp')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_user')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_api_key')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_model')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_total_tokens')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_ttft')} / {t('usage_stats.sub2api_duration_short')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_cost')} / {t('usage_stats.sub2api_status')}</div>
          </div>

          <div>
            {events.events.map((event) => (
              <div key={event.id} className={styles.eventRow}>
                <div className={styles.cellTime}>{formatDateTime(event.createdAt)}</div>
                <div className={styles.cellUser} title={event.user}>{event.user || '-'}</div>
                <div><span className={styles.cellKey}>{event.apiKey || '-'}</span></div>
                <div className={styles.cellModel}>{event.model || event.requestedModel || event.upstreamModel || 'unknown'}</div>
                <div className={styles.cellTokens}>{formatNumber(event.totalTokens)}</div>
                <div className={styles.cellDuration}>
                  <span className={styles.firstTokenDuration}>{formatDurationMs(event.firstTokenDurationMs, { invalidText: '-' })}</span>
                  <span className={styles.durationSeparator}>/</span>
                  <span className={styles.totalDuration}>{formatDurationMs(event.durationMs, { invalidText: '-' })}</span>
                </div>
                <div>
                  <div className={styles.cellCost}>${event.actualCost.toFixed(6)}</div>
                  <span className={`${styles.badge} ${getBadgeClass(event.status)}`}>{event.status}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </section>
  )
}
