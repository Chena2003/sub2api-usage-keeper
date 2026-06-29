import { create } from 'zustand'
import {
  fetchSub2ApiAccountQuotas,
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
  Sub2ApiEventsResponse,
  Sub2ApiModelUsage,
  Sub2ApiOverview,
  Sub2ApiRanking,
  Sub2ApiRankingDimension,
  Sub2ApiRankingTrendResponse,
  Sub2ApiServiceHealth,
  Sub2ApiTimeseriesPoint,
} from '../lib/sub2apiTypes'

type Sub2ApiDashboardState = {
  accounts: Sub2ApiAccount[]
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  models: Sub2ApiModelUsage[]
  rankings: Sub2ApiRanking[]
  rankingTrend: Sub2ApiRankingTrendResponse | null
  events: Sub2ApiEventsResponse
  quotaAccounts: Sub2ApiAccount[]
  serviceHealth: Sub2ApiServiceHealth | null
  rankingDimension: Sub2ApiRankingDimension
  currentDays: number
  loading: boolean
  error: string | null
  _refreshVersion: number
  refresh: (opts?: { hours?: number }) => Promise<void>
  loadRankings: (dimension?: Sub2ApiRankingDimension) => Promise<void>
  loadEvents: (params?: { page?: number; limit?: number }) => Promise<void>
  loadAccountQuotas: (days?: number) => Promise<void>
}

export const useSub2ApiDashboardStore = create<Sub2ApiDashboardState>((set, get) => ({
  accounts: [],
  overview: null,
  points: [],
  models: [],
  rankings: [],
  rankingTrend: null,
  events: { events: [], total: 0, page: 1, limit: 100 },
  quotaAccounts: [],
  serviceHealth: null,
  rankingDimension: 'user',
  currentDays: 7,
  loading: false,
  error: null,
  _refreshVersion: 0,
  refresh: async (opts) => {
    const { rankingDimension, _refreshVersion } = get()
    const version = _refreshVersion + 1
    const hours = opts?.hours ?? 24
    const days = Math.max(1, Math.ceil(hours / 24))
    set({ _refreshVersion: version, loading: true, error: null, currentDays: days })
    try {
      const [accounts, overview, points, models, rankings, rankingTrend, events, quotaAccounts, serviceHealth] = await Promise.all([
        fetchSub2ApiAccounts(days),
        fetchSub2ApiOverview(undefined, hours),
        fetchSub2ApiTimeseries(hours),
        fetchSub2ApiModels(days, 20),
        fetchSub2ApiRankings(rankingDimension, days),
        fetchSub2ApiRankingTrend(rankingDimension, days, 12),
        fetchSub2ApiEvents({ page: 1, limit: 100 }),
        fetchSub2ApiAccountQuotas(days),
        fetchSub2ApiHealth(168),
      ])
      if (get()._refreshVersion !== version) return
      set({ accounts, overview, points, models, rankings, rankingTrend, events, quotaAccounts, serviceHealth, rankingDimension, loading: false })
    } catch (error) {
      if (get()._refreshVersion !== version) return
      set({ error: error instanceof Error ? error.message : 'Unknown error', loading: false })
    }
  },
  loadRankings: async (dimension = 'user') => {
    const { currentDays } = get()
    const [rankings, rankingTrend] = await Promise.all([
      fetchSub2ApiRankings(dimension, currentDays),
      fetchSub2ApiRankingTrend(dimension, currentDays, 12),
    ])
    set({ rankings, rankingTrend, rankingDimension: dimension })
  },
  loadEvents: async (params = {}) => {
    const events = await fetchSub2ApiEvents(params)
    set({ events })
  },
  loadAccountQuotas: async (days = 7) => {
    const quotaAccounts = await fetchSub2ApiAccountQuotas(days)
    set({ quotaAccounts })
  },
}))
