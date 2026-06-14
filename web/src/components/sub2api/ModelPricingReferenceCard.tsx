import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Panel } from '@/components/ui/Panel'
import styles from './ModelPricingReferenceCard.module.scss'

type PricingRow = {
  model: string
  displayName: string
  input: number
  output: number
  cacheHit: number
  cacheCreate: number
}

// Prices in $/1M tokens — hardcoded reference data
const PRICING_DATA: PricingRow[] = [
  { model: 'claude-3-5-haiku-20241022', displayName: 'Claude 3.5 Haiku', input: 0.80, output: 4, cacheHit: 0.08, cacheCreate: 1 },
  { model: 'claude-3-5-sonnet-20241022', displayName: 'Claude 3.5 Sonnet', input: 3, output: 15, cacheHit: 0.30, cacheCreate: 3.75 },
  { model: 'claude-haiku-4-5-20251001', displayName: 'Claude Haiku 4.5', input: 1, output: 5, cacheHit: 0.10, cacheCreate: 1.25 },
  { model: 'claude-sonnet-4-20250514', displayName: 'Claude Sonnet 4', input: 3, output: 15, cacheHit: 0.30, cacheCreate: 3.75 },
  { model: 'claude-sonnet-4-5-20250929', displayName: 'Claude Sonnet 4.5', input: 3, output: 15, cacheHit: 0.30, cacheCreate: 3.75 },
  { model: 'claude-sonnet-4-6-20260217', displayName: 'Claude Sonnet 4.6', input: 3, output: 15, cacheHit: 0.30, cacheCreate: 3.75 },
  { model: 'claude-opus-4-20250514', displayName: 'Claude Opus 4', input: 15, output: 75, cacheHit: 1.50, cacheCreate: 18.75 },
  { model: 'claude-opus-4-5-20251101', displayName: 'Claude Opus 4.5', input: 5, output: 25, cacheHit: 0.50, cacheCreate: 6.25 },
  { model: 'claude-opus-4-6-20260206', displayName: 'Claude Opus 4.6', input: 5, output: 25, cacheHit: 0.50, cacheCreate: 6.25 },
  { model: 'claude-opus-4-7', displayName: 'Claude Opus 4.7', input: 5, output: 25, cacheHit: 0.50, cacheCreate: 6.25 },
  { model: 'claude-opus-4-8', displayName: 'Claude Opus 4.8', input: 5, output: 25, cacheHit: 0.50, cacheCreate: 6.25 },
  { model: 'gpt-4.1', displayName: 'GPT-4.1', input: 2, output: 8, cacheHit: 0.50, cacheCreate: 0 },
  { model: 'gpt-4.1-mini', displayName: 'GPT-4.1 Mini', input: 0.40, output: 1.60, cacheHit: 0.10, cacheCreate: 0 },
  { model: 'gpt-4.1-nano', displayName: 'GPT-4.1 Nano', input: 0.10, output: 0.40, cacheHit: 0.025, cacheCreate: 0 },
  { model: 'gpt-5', displayName: 'GPT-5', input: 1.25, output: 10, cacheHit: 0.125, cacheCreate: 0 },
  { model: 'gpt-5-mini', displayName: 'GPT-5 Mini', input: 0.25, output: 2, cacheHit: 0.025, cacheCreate: 0 },
  { model: 'o3', displayName: 'OpenAI o3', input: 2, output: 8, cacheHit: 0.50, cacheCreate: 0 },
  { model: 'o3-mini', displayName: 'OpenAI o3-mini', input: 0.55, output: 2.20, cacheHit: 0.55, cacheCreate: 0 },
  { model: 'o4-mini', displayName: 'OpenAI o4-mini', input: 1.10, output: 4.40, cacheHit: 0.275, cacheCreate: 0 },
  { model: 'gemini-2.5-flash', displayName: 'Gemini 2.5 Flash', input: 0.30, output: 2.5, cacheHit: 0.03, cacheCreate: 0 },
  { model: 'gemini-2.5-pro', displayName: 'Gemini 2.5 Pro', input: 1.25, output: 10, cacheHit: 0.125, cacheCreate: 0 },
  { model: 'deepseek-chat', displayName: 'DeepSeek Chat', input: 0.27, output: 1.10, cacheHit: 0.07, cacheCreate: 0 },
  { model: 'deepseek-v3', displayName: 'DeepSeek V3', input: 0.28, output: 1.11, cacheHit: 0.028, cacheCreate: 0 },
  { model: 'deepseek-reasoner', displayName: 'DeepSeek Reasoner', input: 0.55, output: 2.19, cacheHit: 0.14, cacheCreate: 0 },
  { model: 'grok-3', displayName: 'Grok 3', input: 3, output: 15, cacheHit: 0.75, cacheCreate: 0 },
  { model: 'grok-4', displayName: 'Grok 4', input: 3, output: 15, cacheHit: 0.75, cacheCreate: 0 },
  { model: 'qwen3-235b-a22b', displayName: 'Qwen3 235B-A22B', input: 0.70, output: 8.40, cacheHit: 0, cacheCreate: 0 },
  { model: 'mistral-large-3-2512', displayName: 'Mistral Large 3', input: 0.50, output: 1.50, cacheHit: 0.05, cacheCreate: 0 },
  { model: 'kimi-k2-0905', displayName: 'Kimi K2', input: 0.55, output: 2.20, cacheHit: 0.10, cacheCreate: 0 },
]

const fmt = (v: number) => v === 0 ? '—' : `$${v % 1 === 0 ? v.toFixed(0) : v}`

export function ModelPricingReferenceCard() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return PRICING_DATA
    return PRICING_DATA.filter((r) =>
      r.model.toLowerCase().includes(q) || r.displayName.toLowerCase().includes(q)
    )
  }, [search])

  return (
    <Panel
      eyebrow="Settings"
      title={t('usage_stats.model_pricing_title', 'Model Pricing Reference')}
      as="section"
      actions={
        <input
          className={styles.searchInput}
          type="search"
          placeholder={t('common.search', 'Search…')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Filter models"
        />
      }
    >
      <p className={styles.hint}>
        {t('usage_stats.model_pricing_hint', 'Reference pricing per 1M tokens. Prices may vary — check provider docs for the latest rates.')}
      </p>
      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Model</th>
              <th>Display Name</th>
              <th>Input</th>
              <th>Output</th>
              <th>Cache Hit</th>
              <th>Cache Create</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((row) => (
              <tr key={row.model}>
                <td><code className={styles.code}>{row.model}</code></td>
                <td className={styles.displayName}>{row.displayName}</td>
                <td className={styles.price}>{fmt(row.input)}</td>
                <td className={styles.price}>{fmt(row.output)}</td>
                <td className={`${styles.price} ${styles.muted}`}>{fmt(row.cacheHit)}</td>
                <td className={`${styles.price} ${styles.muted}`}>{fmt(row.cacheCreate)}</td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={6} className={styles.empty}>No models match "{search}"</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </Panel>
  )
}
