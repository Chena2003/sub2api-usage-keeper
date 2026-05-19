import { create } from 'zustand'
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
  Sub2ApiEventsResponse,
  Sub2ApiModelUsage,
  Sub2ApiOverview,
  Sub2ApiRanking,
  Sub2ApiRankingDimension,
  Sub2ApiTimeseriesPoint,
} from '../lib/sub2apiTypes'

type Sub2ApiDashboardState = {
  accounts: Sub2ApiAccount[]
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  models: Sub2ApiModelUsage[]
  rankings: Sub2ApiRanking[]
  events: Sub2ApiEventsResponse
  quotaAccounts: Sub2ApiAccount[]
  rankingDimension: Sub2ApiRankingDimension
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
  loadRankings: (dimension?: Sub2ApiRankingDimension) => Promise<void>
  loadEvents: (params?: { page?: number; limit?: number }) => Promise<void>
  loadAccountQuotas: (days?: number) => Promise<void>
}

export const useSub2ApiDashboardStore = create<Sub2ApiDashboardState>((set) => ({
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
  refresh: async () => {
    set({ loading: true, error: null })
    try {
      const [accounts, overview, points, models, rankings, events, quotaAccounts] = await Promise.all([
        fetchSub2ApiAccounts(7),
        fetchSub2ApiOverview(7),
        fetchSub2ApiTimeseries(24),
        fetchSub2ApiModels(7, 20),
        fetchSub2ApiRankings('user'),
        fetchSub2ApiEvents({ page: 1, limit: 100 }),
        fetchSub2ApiAccountQuotas(7),
      ])
      set({ accounts, overview, points, models, rankings, events, quotaAccounts, rankingDimension: 'user', loading: false })
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Unknown error', loading: false })
    }
  },
  loadRankings: async (dimension = 'user') => {
    const rankings = await fetchSub2ApiRankings(dimension)
    set({ rankings, rankingDimension: dimension })
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
