import { create } from 'zustand'
import { fetchSub2ApiAccounts, fetchSub2ApiModels, fetchSub2ApiOverview, fetchSub2ApiTimeseries } from '../lib/sub2apiApi'
import type { Sub2ApiAccount, Sub2ApiModelUsage, Sub2ApiOverview, Sub2ApiTimeseriesPoint } from '../lib/sub2apiTypes'

type Sub2ApiDashboardState = {
  accounts: Sub2ApiAccount[]
  overview: Sub2ApiOverview | null
  points: Sub2ApiTimeseriesPoint[]
  models: Sub2ApiModelUsage[]
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
}

export const useSub2ApiDashboardStore = create<Sub2ApiDashboardState>((set) => ({
  accounts: [],
  overview: null,
  points: [],
  models: [],
  loading: false,
  error: null,
  refresh: async () => {
    set({ loading: true, error: null })
    try {
      const [accounts, overview, points, models] = await Promise.all([
        fetchSub2ApiAccounts(7),
        fetchSub2ApiOverview(7),
        fetchSub2ApiTimeseries(24),
        fetchSub2ApiModels(7, 20),
      ])
      set({ accounts, overview, points, models, loading: false })
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Unknown error', loading: false })
    }
  },
}))
