import { readFileSync } from 'node:fs'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import type { OfficialModelPricingResponse } from '@/lib/sub2apiTypes'
import {
  ModelPricingReferenceCard,
  buildOfficialPricingRows,
  formatOfficialPrice,
} from './ModelPricingReferenceCard'
import '@/i18n'

const response: OfficialModelPricingResponse = {
  providers: [
    {
      id: 'anthropic',
      name: 'Anthropic',
      models: [
        {
          id: 'claude-sonnet-test',
          name: 'Claude Sonnet Test',
          cost: { input: 3, output: 15, cacheRead: 0.3, cacheWrite: 3.75 },
          tiers: [
            {
              type: 'context',
              size: 200000,
              cost: { input: 6, output: 22.5, cacheRead: 0.6, cacheWrite: 7.5 },
            },
          ],
        },
        {
          id: 'claude-haiku-test',
          name: 'Claude Haiku Test',
          cost: { input: 1, output: 5, cacheRead: 0.1, cacheWrite: 1.25 },
          tiers: [],
        },
      ],
    },
    {
      id: 'openai',
      name: 'OpenAI',
      models: [
        {
          id: 'gpt-test',
          name: 'GPT Test',
          cost: { input: 2, output: 8, cacheRead: 0.5, cacheWrite: null },
          tiers: [],
        },
      ],
    },
  ],
  fetchedAt: '2026-08-01T00:00:00Z',
  stale: false,
}

describe('ModelPricingReferenceCard helpers', () => {
  it('formats missing, zero, and decimal prices distinctly', () => {
    expect(formatOfficialPrice(null)).toBe('—')
    expect(formatOfficialPrice(undefined)).toBe('—')
    expect(formatOfficialPrice(0)).toBe('$0')
    expect(formatOfficialPrice(3.75)).toBe('$3.75')
  })

  it('filters by official provider before searching model metadata', () => {
    const anthropicSonnet = buildOfficialPricingRows(response, 'anthropic', 'sonnet')
    expect(anthropicSonnet).toHaveLength(2)
    expect(anthropicSonnet.every((row) => row.providerId === 'anthropic')).toBe(true)
    expect(anthropicSonnet.every((row) => row.model.id === 'claude-sonnet-test')).toBe(true)

    const allRows = buildOfficialPricingRows(response, 'all', '')
    expect(allRows).toHaveLength(4)
    expect(new Set(allRows.map((row) => row.providerId))).toEqual(new Set(['anthropic', 'openai']))
  })
})

describe('ModelPricingReferenceCard', () => {
  it('starts in the loading state without rendering legacy hardcoded prices', () => {
    const html = renderToStaticMarkup(<ModelPricingReferenceCard />)

    expect(html).toContain('Loading official model pricing')
    expect(html).not.toContain('claude-3-5-haiku-20241022')
  })

  it('does not keep the legacy hardcoded pricing constant', () => {
    const source = readFileSync(new URL('./ModelPricingReferenceCard.tsx', import.meta.url), 'utf8')
    expect(source).not.toContain('PRICING_DATA')
  })
})
