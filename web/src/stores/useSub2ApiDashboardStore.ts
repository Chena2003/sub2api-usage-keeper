import { create } from 'zustand'
import {
  fetchSub2ApiAccounts,
  fetchSub2ApiEvents,
  fetchSub2ApiHealth,
  fetchSub2ApiModels,
  fetchSub2ApiOverview,
  fetchSub2ApiRankings,
  fetchSub2ApiRankingTrend,
  fetchSub2ApiTimeseries,
  type TimeRangeParams,
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

type RefreshOpts = { hours?: number; since?: string; until?: string; background?: boolean }

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
  currentSince: string | undefined
  currentUntil: string | undefined
  loading: boolean
  eventsLoading: boolean
  error: string | null
  _refreshVersion: number
  _healthVersion: number
  _rankingVersion: number
  refresh: (opts?: RefreshOpts) => Promise<void>
  loadHealth: (tr?: TimeRangeParams) => Promise<void>
  loadRankings: (dimension?: Sub2ApiRankingDimension) => Promise<void>
  loadEvents: (params?: { page?: number; limit?: number; hours?: number; since?: string; until?: string }) => Promise<void>
}

function buildTimeRange(state: Sub2ApiDashboardState): TimeRangeParams {
  if (state.currentSince) {
    return { since: state.currentSince, until: state.currentUntil }
  }
  return { hours: state.currentHours }
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
  currentSince: undefined,
  currentUntil: undefined,
  loading: false,
  eventsLoading: false,
  error: null,
  _refreshVersion: 0,
  _healthVersion: 0,
  _rankingVersion: 0,
  refresh: async (opts) => {
    const { rankingDimension, _refreshVersion } = get()
    const version = _refreshVersion + 1
    const hours = opts?.hours ?? 24
    const background = opts?.background ?? false
    const tr: TimeRangeParams = opts?.since
      ? { since: opts.since, until: opts.until }
      : { hours }
    set({
      _refreshVersion: version,
      ...(!background && { loading: true }),
      error: null,
      currentHours: hours,
      currentSince: opts?.since,
      currentUntil: opts?.until,
    })
    try {
      const [accounts, overview, points, models, rankings, rankingTrend] = await Promise.all([
        fetchSub2ApiAccounts(tr),
        fetchSub2ApiOverview(tr),
        fetchSub2ApiTimeseries(tr),
        fetchSub2ApiModels(tr, 20),
        fetchSub2ApiRankings(rankingDimension, tr),
        fetchSub2ApiRankingTrend(rankingDimension, tr, 12),
      ])
      if (get()._refreshVersion !== version) return
      set({
        accounts,
        quotaAccounts: accounts,
        overview,
        points,
        models,
        rankings,
        rankingTrend,
        rankingDimension,
        loading: false,
      })
      void get().loadEvents({ page: 1, limit: 100, ...tr })
    } catch (error) {
      if (get()._refreshVersion !== version) return
      set({ error: error instanceof Error ? error.message : 'Unknown error', loading: false })
    }
  },
  loadHealth: async (tr?: TimeRangeParams) => {
    const version = get()._healthVersion + 1
    const params: TimeRangeParams = tr ?? buildTimeRange(get())
    set({ _healthVersion: version })
    try {
      const serviceHealth = await fetchSub2ApiHealth(params)
      if (get()._healthVersion !== version) return
      set({ serviceHealth })
    } catch (error) {
      if (get()._healthVersion !== version) return
      console.error('loadHealth failed:', error)
    }
  },
  loadRankings: async (dimension = 'user') => {
    const version = get()._rankingVersion + 1
    const tr = buildTimeRange(get())
    set({ _rankingVersion: version, rankingDimension: dimension })
    try {
      const [rankings, rankingTrend] = await Promise.all([
        fetchSub2ApiRankings(dimension, tr),
        fetchSub2ApiRankingTrend(dimension, tr, 12),
      ])
      if (get()._rankingVersion !== version) return
      set({ rankings, rankingTrend, rankingDimension: dimension })
    } catch (error) {
      if (get()._rankingVersion !== version) return
      console.error('loadRankings failed:', error)
    }
  },
  loadEvents: async (params = {}) => {
    set({ eventsLoading: true })
    try {
      const { page, limit, ...tr } = params
      const timeRange: TimeRangeParams = {}
      if (tr.hours) timeRange.hours = tr.hours
      if (tr.since) timeRange.since = tr.since
      if (tr.until) timeRange.until = tr.until
      if (!timeRange.hours && !timeRange.since) {
        const state = get()
        if (state.currentSince) {
          timeRange.since = state.currentSince
          if (state.currentUntil) timeRange.until = state.currentUntil
        } else {
          timeRange.hours = state.currentHours
        }
      }
      const events = await fetchSub2ApiEvents({ page, limit, ...timeRange })
      set({ events })
    } catch (error) {
      console.error('loadEvents failed:', error)
      set({ events: { events: [], total: 0, page: 1, limit: 100 } })
    } finally {
      set({ eventsLoading: false })
    }
  },
}))
