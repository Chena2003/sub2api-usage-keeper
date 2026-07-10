import { afterEach, describe, expect, it, vi } from 'vitest'
import { fetchSub2ApiAccountQuotas, fetchSub2ApiEvents, fetchSub2ApiRankings, sub2apiEndpoint } from './sub2apiApi'

describe('sub2apiEndpoint', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
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

describe('Sub2API data fetchers', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('fetches rankings with TimeRangeParams and credentials', async () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ rankings: [] }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(fetchSub2ApiRankings()).resolves.toEqual([])

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/sub2api/rankings?dimension=user&limit=20', {
      credentials: 'include',
    })
  })

  it('fetches rankings with since/until params', async () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ rankings: [] }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(fetchSub2ApiRankings('user', { since: '2026-07-10T00:00:00.000Z', until: '2026-07-11T00:00:00.000Z' })).resolves.toEqual([])

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/sub2api/rankings?dimension=user&limit=20&since=2026-07-10T00%3A00%3A00.000Z&until=2026-07-11T00%3A00%3A00.000Z', {
      credentials: 'include',
    })
  })

  it('fetches account quotas with TimeRangeParams', async () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ accounts: [] }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(fetchSub2ApiAccountQuotas({ hours: 336 })).resolves.toEqual([])

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/sub2api/account-quotas?hours=336', {
      credentials: 'include',
    })
  })

  it('fetches events with pagination and time range', async () => {
    vi.stubGlobal('window', { __APP_BASE_PATH__: undefined })
    const response = { events: [], total: 0, page: 2, limit: 50 }
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(response),
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(fetchSub2ApiEvents({ page: 2, limit: 50, hours: 24 })).resolves.toEqual(response)

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/sub2api/events?page=2&limit=50&hours=24', {
      credentials: 'include',
    })
  })
})
