import { useTranslation } from 'react-i18next'
import type { Sub2ApiEventsResponse } from '@/lib/sub2apiTypes'
import { Panel } from '@/components/ui/Panel'
import { StatusPill } from '@/components/ui/StatusPill'
import styles from './RequestEventsPanel.module.scss'

type RequestEventsPanelProps = {
  events: Sub2ApiEventsResponse
  loading?: boolean
  error?: string | null
  onRetry?: () => void
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

function eventVariant(status: string): 'ok' | 'warn' | 'danger' {
  const s = status.toLowerCase()
  const code = parseInt(status, 10)
  if (s === 'success' || code === 200) return 'ok'
  if (s.includes('fail') || s.includes('error') || s.includes('timeout') || (code >= 400)) return 'danger'
  return 'warn'
}

export function RequestEventsPanel({ events, loading, error, onRetry }: RequestEventsPanelProps) {
  const { t } = useTranslation()

  return (
    <Panel
      eyebrow={t('usage_stats.sub2api_events_eyebrow')}
      title={t('usage_stats.sub2api_events_title')}
      as="section"
    >
      {loading ? (
        <Panel.Loading rows={6} />
      ) : error ? (
        <Panel.Error message={t('usage_stats.sub2api_events_title') + ' — ' + error} onRetry={onRetry} />
      ) : events.events.length === 0 ? (
        <Panel.Empty message={t('usage_stats.sub2api_no_data')} />
      ) : (
        <div className={styles.tableWrap}>
          <table className={styles.table} aria-label="请求事件列表">
            <thead>
              <tr>
                <th scope="col">{t('usage_stats.sub2api_timestamp')}</th>
                <th scope="col">{t('usage_stats.sub2api_user')}</th>
                <th scope="col">{t('usage_stats.sub2api_api_key')}</th>
                <th scope="col">{t('usage_stats.sub2api_model')}</th>
                <th scope="col">{t('usage_stats.sub2api_total_tokens')}</th>
                <th scope="col">TTFT / DUR</th>
                <th scope="col">{t('usage_stats.sub2api_cost')}</th>
                <th scope="col">{t('usage_stats.sub2api_status')}</th>
              </tr>
            </thead>
            <tbody>
              {events.events.map((event) => (
                <tr key={event.id}>
                  <td className={styles.mono}>{formatDateTime(event.createdAt)}</td>
                  <td className={styles.truncate}>{event.user || '-'}</td>
                  <td><code className={styles.code}>{event.apiKey ? event.apiKey.slice(0, 8) + '…' : '-'}</code></td>
                  <td className={styles.truncate}>{event.model || event.requestedModel || event.upstreamModel || 'unknown'}</td>
                  <td className={styles.mono}>{formatNumber(event.totalTokens)}</td>
                  <td className={styles.duration}>
                    <span className={styles.ttft}>{formatDuration(event.firstTokenDurationMs)}</span>
                    <span className={styles.sep}>/</span>
                    <span className={styles.dur}>{formatDuration(event.durationMs)}</span>
                  </td>
                  <td className={`${styles.mono} ${styles.cost}`}>${event.actualCost.toFixed(6)}</td>
                  <td><StatusPill variant={eventVariant(event.status)} label={event.status} dot={false} /></td>
                </tr>
              ))}
            </tbody>
          </table>
          {events.total > events.events.length && (
            <div className={styles.pagination}>
              <span className={styles.mono}>{events.events.length} / {formatNumber(events.total)}</span>
            </div>
          )}
        </div>
      )}
    </Panel>
  )
}
