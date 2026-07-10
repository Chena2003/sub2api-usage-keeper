import { ApiError, apiPath } from './api'
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
} from './sub2apiTypes'

type AccountsResponse = { accounts: Sub2ApiAccount[] }
type TimeseriesResponse = { points: Sub2ApiTimeseriesPoint[] }
type ModelsResponse = { models: Sub2ApiModelUsage[] }
type RankingsResponse = { rankings: Sub2ApiRanking[] }

export function sub2apiEndpoint(path: string): string {
  const normalized = path.startsWith('/') ? path : `/${path}`
  return apiPath(`/sub2api${normalized}`)
}

async function getJson<T>(path: string): Promise<T> {
  const response = await fetch(sub2apiEndpoint(path), { credentials: 'include' })
  if (!response.ok) {
    throw new ApiError(`Request failed with status ${response.status}`, response.status)
  }
  return response.json() as Promise<T>
}

export async function fetchSub2ApiAccounts(hours = 168): Promise<Sub2ApiAccount[]> {
  const body = await getJson<AccountsResponse>(`/accounts?hours=${hours}`)
  return body.accounts
}

export function fetchSub2ApiOverview(days?: number, hours?: number): Promise<Sub2ApiOverview> {
  const params = new URLSearchParams()
  if (hours !== undefined) {
    params.set('hours', String(hours))
  } else {
    params.set('days', String(days ?? 7))
  }
  return getJson<Sub2ApiOverview>(`/overview?${params.toString()}`)
}

export async function fetchSub2ApiTimeseries(hours = 24): Promise<Sub2ApiTimeseriesPoint[]> {
  const body = await getJson<TimeseriesResponse>(`/timeseries?hours=${hours}`)
  return body.points
}

export async function fetchSub2ApiModels(hours = 168, limit = 20): Promise<Sub2ApiModelUsage[]> {
  const body = await getJson<ModelsResponse>(`/models?hours=${hours}&limit=${limit}`)
  return body.models
}

export async function fetchSub2ApiRankings(
  dimension: Sub2ApiRankingDimension = 'user',
  hours = 168,
  limit = 20,
): Promise<Sub2ApiRanking[]> {
  const params = new URLSearchParams({
    dimension,
    hours: String(hours),
    limit: String(limit),
  })
  const body = await getJson<RankingsResponse>(`/rankings?${params.toString()}`)
  return body.rankings
}

export async function fetchSub2ApiAccountQuotas(hours = 168): Promise<Sub2ApiAccount[]> {
  const params = new URLSearchParams({ hours: String(hours) })
  const body = await getJson<AccountsResponse>(`/account-quotas?${params.toString()}`)
  return body.accounts
}

export function fetchSub2ApiEvents({
  page = 1,
  limit = 100,
  hours,
}: { page?: number; limit?: number; hours?: number } = {}): Promise<Sub2ApiEventsResponse> {
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  })
  if (hours !== undefined) {
    params.set('hours', String(hours))
  }
  return getJson<Sub2ApiEventsResponse>(`/events?${params.toString()}`)
}

export function fetchSub2ApiHealth(hours = 24): Promise<Sub2ApiServiceHealth> {
  return getJson<Sub2ApiServiceHealth>(`/health?hours=${hours}`)
}

export async function fetchSub2ApiRankingTrend(
  dimension: Sub2ApiRankingDimension = 'user',
  hours = 168,
  limit = 12,
): Promise<Sub2ApiRankingTrendResponse> {
  return getJson<Sub2ApiRankingTrendResponse>(`/rankings-trend?dimension=${dimension}&hours=${hours}&limit=${limit}`)
}
