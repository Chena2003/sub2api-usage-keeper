import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Panel } from '@/components/ui/Panel'
import { fetchSub2ApiModelPricing } from '@/lib/sub2apiApi'
import type {
  OfficialModelCostTier,
  OfficialModelPricing,
  OfficialModelPricingResponse,
} from '@/lib/sub2apiTypes'
import styles from './ModelPricingReferenceCard.module.scss'

export type OfficialPricingRow = {
  providerId: string
  providerName: string
  model: OfficialModelPricing
  tier?: OfficialModelCostTier
}

export const formatOfficialPrice = (value: number | null | undefined): string => {
  if (value === null || value === undefined) return '—'
  return `$${new Intl.NumberFormat('en-US', {
    maximumFractionDigits: 6,
    useGrouping: false,
  }).format(value)}`
}

export function buildOfficialPricingRows(
  data: OfficialModelPricingResponse,
  selectedProvider: string,
  search: string,
): OfficialPricingRow[] {
  const query = search.trim().toLowerCase()
  const rows: OfficialPricingRow[] = []

  for (const provider of data.providers) {
    if (selectedProvider !== 'all' && provider.id !== selectedProvider) continue

    for (const model of provider.models) {
      const matchesSearch = query === '' || [provider.name, model.id, model.name]
        .some((value) => value.toLowerCase().includes(query))
      if (!matchesSearch) continue

      rows.push({
        providerId: provider.id,
        providerName: provider.name,
        model,
      })
      for (const tier of model.tiers) {
        rows.push({
          providerId: provider.id,
          providerName: provider.name,
          model,
          tier,
        })
      }
    }
  }

  return rows
}

export function ModelPricingReferenceCard() {
  const { t } = useTranslation()
  const [data, setData] = useState<OfficialModelPricingResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [provider, setProvider] = useState('all')
  const [retryKey, setRetryKey] = useState(0)

  useEffect(() => {
    let active = true

    void fetchSub2ApiModelPricing()
      .then((response) => {
        if (!active) return
        setData(response)
      })
      .catch((requestError: unknown) => {
        if (!active) return
        setError(requestError instanceof Error ? requestError.message : String(requestError))
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [retryKey])

  const rows = useMemo(
    () => data ? buildOfficialPricingRows(data, provider, search) : [],
    [data, provider, search],
  )

  let updatedAt: string | null = null
  if (data?.fetchedAt) {
    const timestamp = new Date(data.fetchedAt)
    updatedAt = Number.isNaN(timestamp.getTime()) ? data.fetchedAt : timestamp.toLocaleString()
  }

  const filters = (
    <div className={styles.filters}>
      <select
        className={styles.providerSelect}
        value={provider}
        onChange={(event) => setProvider(event.target.value)}
        aria-label={t('usage_stats.model_pricing_col_provider')}
        disabled={!data}
      >
        <option value="all">{t('usage_stats.model_pricing_provider_all')}</option>
        {data?.providers.map((item) => (
          <option key={item.id} value={item.id}>{item.name}</option>
        ))}
      </select>
      <input
        className={styles.searchInput}
        type="search"
        placeholder={t('common.search', 'Search…')}
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        aria-label={t('common.search', 'Search…')}
        disabled={!data}
      />
    </div>
  )

  return (
    <Panel
      eyebrow={t('usage_stats.model_pricing_eyebrow')}
      title={t('usage_stats.model_pricing_title')}
      as="section"
      actions={filters}
    >
      <p className={styles.hint}>{t('usage_stats.model_pricing_hint')}</p>

      {loading && data === null && (
        <p className={styles.meta} role="status">{t('usage_stats.model_pricing_loading')}</p>
      )}

      {error && data === null && (
        <div className={styles.error} role="alert">
          <span>{t('usage_stats.model_pricing_error')}</span>
          <button
            className={styles.retryButton}
            type="button"
            onClick={() => {
              setLoading(true)
              setError(null)
              setRetryKey((value) => value + 1)
            }}
          >
            {t('usage_stats.model_pricing_retry')}
          </button>
        </div>
      )}

      {data && (
        <>
          <div className={styles.meta}>
            {data.stale && (
              <span className={styles.stale} role="status">{t('usage_stats.model_pricing_stale')}</span>
            )}
            {updatedAt && (
              <span>{t('usage_stats.model_pricing_updated_at', { time: updatedAt })}</span>
            )}
          </div>
          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>{t('usage_stats.model_pricing_col_provider')}</th>
                  <th>{t('usage_stats.model_pricing_col_model')}</th>
                  <th>{t('usage_stats.model_pricing_col_display_name')}</th>
                  <th>{t('usage_stats.model_pricing_col_input')}</th>
                  <th>{t('usage_stats.model_pricing_col_output')}</th>
                  <th>{t('usage_stats.model_pricing_col_cache_read')}</th>
                  <th>{t('usage_stats.model_pricing_col_cache_write')}</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((row, index) => {
                  const cost = row.tier?.cost ?? row.model.cost
                  return (
                    <tr
                      className={row.tier ? styles.tierRow : undefined}
                      key={`${row.providerId}:${row.model.id}:${row.tier ? `${row.tier.type}:${row.tier.size}:${index}` : 'base'}`}
                    >
                      <td className={styles.providerName}>{row.tier ? '' : row.providerName}</td>
                      <td>
                        {row.tier ? (
                          <span className={styles.tierLabel}>
                            {t('usage_stats.model_pricing_tier_context', {
                              size: new Intl.NumberFormat().format(row.tier.size),
                            })}
                          </span>
                        ) : (
                          <code className={styles.code}>{row.model.id}</code>
                        )}
                      </td>
                      <td className={styles.displayName}>
                        {!row.tier && (
                          <>
                            {row.model.name}
                            {row.model.status === 'deprecated' && (
                              <span className={styles.statusBadge}>
                                {t('usage_stats.model_pricing_deprecated')}
                              </span>
                            )}
                          </>
                        )}
                      </td>
                      <td className={styles.price}>{formatOfficialPrice(cost.input)}</td>
                      <td className={styles.price}>{formatOfficialPrice(cost.output)}</td>
                      <td className={`${styles.price} ${styles.muted}`}>{formatOfficialPrice(cost.cacheRead)}</td>
                      <td className={`${styles.price} ${styles.muted}`}>{formatOfficialPrice(cost.cacheWrite)}</td>
                    </tr>
                  )
                })}
                {rows.length === 0 && (
                  <tr>
                    <td colSpan={7} className={styles.empty}>
                      {t('usage_stats.model_pricing_empty', { query: search })}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </Panel>
  )
}
