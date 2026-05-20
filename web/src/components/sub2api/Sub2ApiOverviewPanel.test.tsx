import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { Sub2ApiOverviewPanel } from './Sub2ApiOverviewPanel'
import type { Sub2ApiAccount } from '@/lib/sub2apiTypes'
import '@/i18n'

const baseAccount: Sub2ApiAccount = {
  id: 1,
  provider: 'anthropic',
  accountType: 'claude',
  displayName: 'Claude account',
  status: 'active',
  schedulable: true,
  credentialKeys: [],
  usage: {
    totalRequests: 0,
    inputTokens: 0,
    outputTokens: 0,
    cacheTokens: 0,
    totalTokens: 0,
    actualCost: 0,
    averageDurationMs: 0,
  },
}

describe('Sub2ApiOverviewPanel', () => {
  it('renders quota risk stats without exposing interpolation placeholders', () => {
    const html = renderToStaticMarkup(
      <Sub2ApiOverviewPanel
        overview={null}
        points={[]}
        models={[]}
        quotaAccounts={[
          {
            ...baseAccount,
            fiveHourWindow: {
              consumed: 8,
              limit: 10,
              ratio: 0.8,
              status: 'warning',
            },
          },
        ]}
      />,
    )

    expect(html).not.toContain('{{count}}')
  })
})
