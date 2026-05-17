import { afterEach, describe, expect, it, vi } from 'vitest'
import { sub2apiEndpoint } from './sub2apiApi'

describe('sub2apiEndpoint', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('builds root path endpoints', () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    expect(sub2apiEndpoint('/accounts')).toBe('/api/v1/sub2api/accounts')
  })

  it('preserves query strings', () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    expect(sub2apiEndpoint('/overview?days=7')).toBe('/api/v1/sub2api/overview?days=7')
  })

  it('uses configured app base path', () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: '/dashboard' })
    expect(sub2apiEndpoint('/models?days=7&limit=20')).toBe('/dashboard/api/v1/sub2api/models?days=7&limit=20')
  })
})
