import { useTranslation } from 'react-i18next'
import type { Sub2ApiEventsResponse } from '@/lib/sub2apiTypes'
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

const formatDuration = (ms?: number) => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const getBadgeClass = (status: string) => {
  const s = status.toLowerCase()
  if (s === 'success' || s === '200') return styles.badgeSuccess
  if (s.includes('fail') || s.includes('error') || s.includes('timeout') || s >= '400') return styles.badgeError
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
        <div style={{ color: '#94a3b8' }}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div style={{ marginTop: '24px' }}>
          <div className={styles.eventGrid}>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_timestamp')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_user')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_api_key')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_model')}</div>
            <div className={styles.gridHeader}>{t('usage_stats.sub2api_total_tokens')}</div>
            <div className={styles.gridHeader}>TTFT / DUR</div>
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
                  <span style={{ color: '#38bdf8' }}>{formatDuration(event.firstTokenDurationMs)}</span>
                  <span style={{ margin: '0 4px', color: '#64748b' }}>/</span>
                  <span style={{ color: '#2dd4bf' }}>{formatDuration(event.durationMs)}</span>
                </div>
                <div>
                  <div className={styles.cellCost} style={{ marginBottom: '4px' }}>${event.actualCost.toFixed(6)}</div>
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
