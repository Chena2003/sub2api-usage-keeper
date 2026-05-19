import { useTranslation } from 'react-i18next'
import type { Sub2ApiEventsResponse } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'

type RequestEventsPanelProps = {
  events: Sub2ApiEventsResponse
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)

const formatDateTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value || '-'
  return date.toLocaleString()
}

export function RequestEventsPanel({ events }: RequestEventsPanelProps) {
  const { t } = useTranslation()

  return (
    <section className="card">
      <div className={styles.sectionTitleBlock}>
        <span className={styles.sectionEyebrow}>{t('usage_stats.sub2api_events_eyebrow')}</span>
        <h3 className={styles.sectionTitle}>{t('usage_stats.sub2api_events_title')}</h3>
        <p className={styles.sectionSubtitle}>{t('usage_stats.sub2api_events_hint')}</p>
      </div>

      {events.events.length === 0 ? (
        <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t('usage_stats.sub2api_timestamp')}</th>
                <th>{t('usage_stats.sub2api_user')}</th>
                <th>{t('usage_stats.sub2api_api_key')}</th>
                <th>{t('usage_stats.sub2api_model')}</th>
                <th>{t('usage_stats.sub2api_total_tokens')}</th>
                <th>{t('usage_stats.sub2api_status')}</th>
              </tr>
            </thead>
            <tbody>
              {events.events.map((event) => (
                <tr key={event.id}>
                  <td>{formatDateTime(event.createdAt)}</td>
                  <td className={styles.modelCell}>{event.user || '-'}</td>
                  <td className={styles.modelCell}>{event.apiKey || '-'}</td>
                  <td className={styles.modelCell}>{event.model || event.requestedModel || event.upstreamModel || 'unknown'}</td>
                  <td>{formatNumber(event.totalTokens)}</td>
                  <td>{event.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
