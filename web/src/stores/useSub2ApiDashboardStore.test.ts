import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSub2ApiDashboardStore } from './useSub2ApiDashboardStore'
import {
  fetchSub2ApiAccounts,
  fetchSub2ApiEvents,
  fetchSub2ApiHealth,
  fetchSub2ApiModels,
  fetchSub2ApiOverview,
  fetchSub2ApiRankings,
  fetchSub2ApiRankingTrend,
  fetchSub2ApiTimeseries,
} from '../lib/sub2apiApi'
import type {
  Sub2ApiAccount,
  Sub2ApiEvent,
  Sub2ApiEventsResponse,
  Sub2ApiModelUsage,
  Sub2ApiOverview,
  Sub2ApiRanking,
  Sub2ApiTimeseriesPoint,
} from '../lib/sub2apiTypes'

vi.mock('../lib/sub2apiApi', () => ({
  fetchSub2ApiAccounts: vi.fn(),
  fetchSub2ApiEvents: vi.fn(),
  fetchSub2ApiHealth: vi.fn(),
  fetchSub2ApiModels: vi.fn(),
  fetchSub2ApiOverview: vi.fn(),
  fetchSub2ApiRankings: vi.fn(),
  fetchSub2ApiRankingTrend: vi.fn(),
  fetchSub2ApiTimeseries: vi.fn(),
}))

const account: Sub2ApiAccount = {
  id: 1,
  provider: 'anthropic',
  accountType: 'subscription',
  displayName: 'Team account',
  status: 'active',
  schedulable: true,
  credentialKeys: [],
  usage: {
    totalRequests: 1,
    inputTokens: 10,
    outputTokens: 5,
    cacheTokens: 2,
    totalTokens: 17,
    actualCost: 0.12,
    averageDurationMs: 100,
  },
}

const overview: Sub2ApiOverview = {
  accountCount: 1,
  activeAccountCount: 1,
  totalRequests: 2,
  inputTokens: 20,
  outputTokens: 10,
  cacheTokens: 4,
  totalTokens: 34,
  actualCost: 0.24,
  accountCost: 0,
  activeUsers: 1,
}

const point: Sub2ApiTimeseriesPoint = {
  bucketStart: '2026-05-19T00:00:00Z',
  totalRequests: 2,
  inputTokens: 20,
  outputTokens: 10,
  cacheCreationTokens: 1,
  cacheReadTokens: 3,
  totalCost: 0.24,
  actualCost: 0.24,
  accountCost: 0,
  totalDurationMS: 200,
  activeUsers: 1,
}

const model: Sub2ApiModelUsage = {
  model: 'claude-sonnet-4-6',
  requestedModel: 'claude-sonnet-4-6',
  upstreamModel: 'claude-sonnet-4-6',
  totalRequests: 2,
  inputTokens: 20,
  outputTokens: 10,
  cacheCreationTokens: 1,
  cacheReadTokens: 3,
  totalCost: 0.24,
  actualCost: 0.24,
  averageDurationMS: 100,
}

const ranking: Sub2ApiRanking = {
  dimension: 'user',
  name: 'masked-user',
  totalRequests: 2,
  inputTokens: 20,
  outputTokens: 10,
  cacheTokens: 4,
  totalTokens: 34,
  actualCost: 0.24,
  share: 1,
}

const event: Sub2ApiEvent = {
  id: 101,
  createdAt: '2026-05-19T00:00:00Z',
  user: 'masked-user',
  apiKey: 'key-abcd',
  model: 'claude-sonnet-4-6',
  requestedModel: 'claude-sonnet-4-6',
  upstreamModel: 'claude-sonnet-4-6',
  accountId: 1,
  accountName: 'Team account',
  status: 'success',
  inputTokens: 20,
  outputTokens: 10,
  cacheTokens: 4,
  totalTokens: 34,
  actualCost: 0.24,
  durationMs: 100,
}

const eventsResponse: Sub2ApiEventsResponse = {
  events: [event],
  total: 1,
  page: 1,
  limit: 100,
}

const mockedFetchSub2ApiAccounts = vi.mocked(fetchSub2ApiAccounts)
const mockedFetchSub2ApiOverview = vi.mocked(fetchSub2ApiOverview)
const mockedFetchSub2ApiTimeseries = vi.mocked(fetchSub2ApiTimeseries)
const mockedFetchSub2ApiModels = vi.mocked(fetchSub2ApiModels)
const mockedFetchSub2ApiRankings = vi.mocked(fetchSub2ApiRankings)
const mockedFetchSub2ApiRankingTrend = vi.mocked(fetchSub2ApiRankingTrend)
const mockedFetchSub2ApiEvents = vi.mocked(fetchSub2ApiEvents)
const mockedFetchSub2ApiHealth = vi.mocked(fetchSub2ApiHealth)

function resetStore() {
  useSub2ApiDashboardStore.setState({
    accounts: [],
    overview: null,
    points: [],
    models: [],
    rankings: [],
    events: { events: [], total: 0, page: 1, limit: 100 },
    quotaAccounts: [],
    serviceHealth: null,
    rankingDimension: 'user',
    currentHours: 168,
    currentSince: undefined,
    currentUntil: undefined,
    loading: false,
    eventsLoading: false,
    error: null,
    _refreshVersion: 0,
    _healthVersion: 0,
    _rankingVersion: 0,
  })
}

describe('useSub2ApiDashboardStore', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetStore()
  })

  it('refreshes all dashboard datasets with TimeRangeParams', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiRankingTrend.mockResolvedValue({ points: [], granularity: 'day' })
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)

    await useSub2ApiDashboardStore.getState().refresh({ hours: 24 })

    expect(mockedFetchSub2ApiAccounts).toHaveBeenCalledWith({ hours: 24 })
    expect(mockedFetchSub2ApiOverview).toHaveBeenCalledWith({ hours: 24 })
    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('user', { hours: 24 })
    expect(mockedFetchSub2ApiEvents).toHaveBeenCalledWith({ page: 1, limit: 100, hours: 24 })
    expect(mockedFetchSub2ApiHealth).not.toHaveBeenCalled()
    // quotaAccounts should be the same as accounts (deduplication)
    expect(useSub2ApiDashboardStore.getState()).toMatchObject({
      accounts: [account],
      quotaAccounts: [account],
      overview,
      points: [point],
      models: [model],
      rankings: [ranking],
      events: eventsResponse,
      loading: false,
      error: null,
    })
  })

  it('refreshes with since/until time range', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiRankingTrend.mockResolvedValue({ points: [], granularity: 'day' })
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)

    const since = '2026-07-10T00:00:00.000Z'
    const until = '2026-07-11T00:00:00.000Z'
    await useSub2ApiDashboardStore.getState().refresh({ since, until })

    expect(mockedFetchSub2ApiAccounts).toHaveBeenCalledWith({ since, until })
    expect(mockedFetchSub2ApiOverview).toHaveBeenCalledWith({ since, until })
    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('user', { since, until })
    expect(mockedFetchSub2ApiEvents).toHaveBeenCalledWith({ page: 1, limit: 100, since, until })
    expect(useSub2ApiDashboardStore.getState().currentSince).toBe(since)
    expect(useSub2ApiDashboardStore.getState().currentUntil).toBe(until)
  })

  it('loads rankings for the selected dimension and updates state', async () => {
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiRankingTrend.mockResolvedValue({ points: [], granularity: 'day' })

    await useSub2ApiDashboardStore.getState().loadRankings('api_key')

    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('api_key', { hours: 168 })
    expect(useSub2ApiDashboardStore.getState().rankingDimension).toBe('api_key')
    expect(useSub2ApiDashboardStore.getState().rankings).toEqual([ranking])
  })

  it('preserves the selected ranking dimension during refresh', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiRankingTrend.mockResolvedValue({ points: [], granularity: 'day' })
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)
    useSub2ApiDashboardStore.setState({ rankingDimension: 'api_key' })

    await useSub2ApiDashboardStore.getState().refresh()

    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('api_key', { hours: 24 })
    expect(useSub2ApiDashboardStore.getState().rankingDimension).toBe('api_key')
  })

  it('loads events with requested pagination and updates state', async () => {
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)

    await useSub2ApiDashboardStore.getState().loadEvents({ page: 2, limit: 50 })

    expect(mockedFetchSub2ApiEvents).toHaveBeenCalledWith({ page: 2, limit: 50, hours: 168 })
    expect(useSub2ApiDashboardStore.getState().events).toEqual(eventsResponse)
  })

  it('loadHealth fetches health data without affecting loading state', async () => {
    const healthData = { total_success: 100, total_failure: 5, success_rate: 95.24, rows: 7, columns: 96, bucket_seconds: 900, window_start: '', window_end: '', block_details: [] }
    mockedFetchSub2ApiHealth.mockResolvedValue(healthData as never)

    await useSub2ApiDashboardStore.getState().loadHealth()

    expect(mockedFetchSub2ApiHealth).toHaveBeenCalledWith({ hours: 168 })
    expect(useSub2ApiDashboardStore.getState().serviceHealth).toEqual(healthData)
    expect(useSub2ApiDashboardStore.getState().loading).toBe(false)
  })

  it('loadHealth uses provided time range instead of currentHours', async () => {
    const healthData = { total_success: 50, total_failure: 2, success_rate: 96.15, rows: 3, columns: 48, bucket_seconds: 900, window_start: '', window_end: '', block_details: [] }
    mockedFetchSub2ApiHealth.mockResolvedValue(healthData as never)

    await useSub2ApiDashboardStore.getState().loadHealth({ hours: 8 })

    expect(mockedFetchSub2ApiHealth).toHaveBeenCalledWith({ hours: 8 })
    expect(useSub2ApiDashboardStore.getState().serviceHealth).toEqual(healthData)
  })

  it('loadHealth uses since/until from currentSince when available', async () => {
    const healthData = { total_success: 50, total_failure: 2, success_rate: 96.15, rows: 3, columns: 48, bucket_seconds: 900, window_start: '', window_end: '', block_details: [] }
    mockedFetchSub2ApiHealth.mockResolvedValue(healthData as never)
    const since = '2026-07-10T00:00:00.000Z'
    const until = '2026-07-11T00:00:00.000Z'
    useSub2ApiDashboardStore.setState({ currentSince: since, currentUntil: until })

    await useSub2ApiDashboardStore.getState().loadHealth()

    expect(mockedFetchSub2ApiHealth).toHaveBeenCalledWith({ since, until })
  })

  it('loadEvents catches errors and resets events state', async () => {
    mockedFetchSub2ApiEvents.mockRejectedValue(new Error('network error'))

    await useSub2ApiDashboardStore.getState().loadEvents({ page: 1, limit: 100 })

    expect(useSub2ApiDashboardStore.getState().events).toEqual({ events: [], total: 0, page: 1, limit: 100 })
    expect(useSub2ApiDashboardStore.getState().eventsLoading).toBe(false)
  })

  it('background refresh does not flip loading to true', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiRankingTrend.mockResolvedValue({ points: [], granularity: 'day' })
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)

    const refreshPromise = useSub2ApiDashboardStore.getState().refresh({ background: true })
    expect(useSub2ApiDashboardStore.getState().loading).toBe(false)

    await refreshPromise
    expect(useSub2ApiDashboardStore.getState().loading).toBe(false)
    expect(useSub2ApiDashboardStore.getState().overview).toEqual(overview)
  })
})
