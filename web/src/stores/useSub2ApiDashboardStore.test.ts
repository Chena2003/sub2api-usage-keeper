import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSub2ApiDashboardStore } from './useSub2ApiDashboardStore'
import {
  fetchSub2ApiAccountQuotas,
  fetchSub2ApiAccounts,
  fetchSub2ApiEvents,
  fetchSub2ApiModels,
  fetchSub2ApiOverview,
  fetchSub2ApiRankings,
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
  fetchSub2ApiAccountQuotas: vi.fn(),
  fetchSub2ApiAccounts: vi.fn(),
  fetchSub2ApiEvents: vi.fn(),
  fetchSub2ApiModels: vi.fn(),
  fetchSub2ApiOverview: vi.fn(),
  fetchSub2ApiRankings: vi.fn(),
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

const quotaAccount: Sub2ApiAccount = {
  ...account,
  fiveHourWindow: { consumed: 10, limit: 100, remaining: 90, ratio: 0.1, status: 'ok' },
  weeklyWindow: { consumed: 20, limit: 200, remaining: 180, ratio: 0.1, status: 'ok' },
}

const mockedFetchSub2ApiAccounts = vi.mocked(fetchSub2ApiAccounts)
const mockedFetchSub2ApiOverview = vi.mocked(fetchSub2ApiOverview)
const mockedFetchSub2ApiTimeseries = vi.mocked(fetchSub2ApiTimeseries)
const mockedFetchSub2ApiModels = vi.mocked(fetchSub2ApiModels)
const mockedFetchSub2ApiRankings = vi.mocked(fetchSub2ApiRankings)
const mockedFetchSub2ApiEvents = vi.mocked(fetchSub2ApiEvents)
const mockedFetchSub2ApiAccountQuotas = vi.mocked(fetchSub2ApiAccountQuotas)

function resetStore() {
  useSub2ApiDashboardStore.setState({
    accounts: [],
    overview: null,
    points: [],
    models: [],
    rankings: [],
    events: { events: [], total: 0, page: 1, limit: 100 },
    quotaAccounts: [],
    rankingDimension: 'user',
    loading: false,
    error: null,
  })
}

describe('useSub2ApiDashboardStore', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetStore()
  })

  it('refreshes all dashboard datasets including rankings, events, and quotas', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)
    mockedFetchSub2ApiAccountQuotas.mockResolvedValue([quotaAccount])

    await useSub2ApiDashboardStore.getState().refresh()

    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('user')
    expect(mockedFetchSub2ApiEvents).toHaveBeenCalledWith({ page: 1, limit: 100 })
    expect(mockedFetchSub2ApiAccountQuotas).toHaveBeenCalledWith(7)
    expect(useSub2ApiDashboardStore.getState()).toMatchObject({
      accounts: [account],
      overview,
      points: [point],
      models: [model],
      rankings: [ranking],
      events: eventsResponse,
      quotaAccounts: [quotaAccount],
      loading: false,
      error: null,
    })
  })

  it('loads rankings for the selected dimension and updates state', async () => {
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])

    await useSub2ApiDashboardStore.getState().loadRankings('api_key')

    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('api_key')
    expect(useSub2ApiDashboardStore.getState().rankingDimension).toBe('api_key')
    expect(useSub2ApiDashboardStore.getState().rankings).toEqual([ranking])
  })

  it('preserves the selected ranking dimension during refresh', async () => {
    mockedFetchSub2ApiAccounts.mockResolvedValue([account])
    mockedFetchSub2ApiOverview.mockResolvedValue(overview)
    mockedFetchSub2ApiTimeseries.mockResolvedValue([point])
    mockedFetchSub2ApiModels.mockResolvedValue([model])
    mockedFetchSub2ApiRankings.mockResolvedValue([ranking])
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)
    mockedFetchSub2ApiAccountQuotas.mockResolvedValue([quotaAccount])
    useSub2ApiDashboardStore.setState({ rankingDimension: 'api_key' })

    await useSub2ApiDashboardStore.getState().refresh()

    expect(mockedFetchSub2ApiRankings).toHaveBeenCalledWith('api_key')
    expect(mockedFetchSub2ApiRankings).not.toHaveBeenCalledWith('user')
    expect(useSub2ApiDashboardStore.getState().rankingDimension).toBe('api_key')
  })

  it('loads events with requested pagination and updates state', async () => {
    mockedFetchSub2ApiEvents.mockResolvedValue(eventsResponse)

    await useSub2ApiDashboardStore.getState().loadEvents({ page: 2, limit: 50 })

    expect(mockedFetchSub2ApiEvents).toHaveBeenCalledWith({ page: 2, limit: 50 })
    expect(useSub2ApiDashboardStore.getState().events).toEqual(eventsResponse)
  })

  it('loads quota accounts for the requested day window and updates state', async () => {
    mockedFetchSub2ApiAccountQuotas.mockResolvedValue([quotaAccount])

    await useSub2ApiDashboardStore.getState().loadAccountQuotas(14)

    expect(mockedFetchSub2ApiAccountQuotas).toHaveBeenCalledWith(14)
    expect(useSub2ApiDashboardStore.getState().quotaAccounts).toEqual([quotaAccount])
  })
})
