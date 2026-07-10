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
  currentHours: number
  loading: boolean
  eventsLoading: boolean
  error: string | null
  _refreshVersion: number
  refresh: (opts?: { hours?: number; background?: boolean }) => Promise<void>
  loadHealth: (hours?: number) => Promise<void>
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
  currentHours: 168,
  loading: false,
  eventsLoading: false,
  error: null,
  _refreshVersion: 0,
  refresh: async (opts) => {
    const { rankingDimension, _refreshVersion } = get()
    const version = _refreshVersion + 1
    const hours = opts?.hours ?? 24
    const background = opts?.background ?? false
    set({ _refreshVersion: version, ...(!background && { loading: true }), error: null, currentHours: hours })
    try {
      const [accounts, overview, points, models, rankings, rankingTrend, quotaAccounts] = await Promise.all([
        fetchSub2ApiAccounts(hours),
        fetchSub2ApiOverview(undefined, hours),
        fetchSub2ApiTimeseries(hours),
        fetchSub2ApiModels(hours, 20),
        fetchSub2ApiRankings(rankingDimension, hours),
        fetchSub2ApiRankingTrend(rankingDimension, hours, 12),
        fetchSub2ApiAccountQuotas(hours),
      ])
      if (get()._refreshVersion !== version) return
      set({ accounts, overview, points, models, rankings, rankingTrend, quotaAccounts, rankingDimension, loading: false })
      // Fire events refresh independently so it doesn't block range-dependent cards.
      void get().loadEvents({ page: 1, limit: 100 })
    } catch (error) {
      if (get()._refreshVersion !== version) return
      set({ error: error instanceof Error ? error.message : 'Unknown error', loading: false })
    }
  },
  loadHealth: async (hours?: number) => {
    const h = hours ?? get().currentHours
    try {
      const serviceHealth = await fetchSub2ApiHealth(h)
      set({ serviceHealth })
    } catch (error) {
      console.error('loadHealth failed:', error)
    }
  },
  loadRankings: async (dimension = 'user') => {
    const { currentHours } = get()
    try {
      const [rankings, rankingTrend] = await Promise.all([
        fetchSub2ApiRankings(dimension, currentHours),
        fetchSub2ApiRankingTrend(dimension, currentHours, 12),
      ])
      set({ rankings, rankingTrend, rankingDimension: dimension })
    } catch (error) {
      console.error('loadRankings failed:', error)
    }
  },
  loadEvents: async (params = {}) => {
    set({ eventsLoading: true })
    try {
      const events = await fetchSub2ApiEvents(params)
      set({ events })
    } finally {
      set({ eventsLoading: false })
    }
  },
  loadAccountQuotas: async (days = 7) => {
    const hours = days * 24
    const quotaAccounts = await fetchSub2ApiAccountQuotas(hours)
    set({ quotaAccounts })
  },
}))
