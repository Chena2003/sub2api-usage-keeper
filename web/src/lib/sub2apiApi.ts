import { ApiError, apiPath } from './api'
import type {
  Sub2ApiAccount,
  Sub2ApiEventsResponse,
  Sub2ApiModelUsage,
  Sub2ApiOverview,
  Sub2ApiRanking,
  Sub2ApiRankingDimension,
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

export async function fetchSub2ApiAccounts(days = 7): Promise<Sub2ApiAccount[]> {
  const body = await getJson<AccountsResponse>(`/accounts?days=${days}`)
  return body.accounts
}

export function fetchSub2ApiOverview(days = 7): Promise<Sub2ApiOverview> {
  return getJson<Sub2ApiOverview>(`/overview?days=${days}`)
}

export async function fetchSub2ApiTimeseries(hours = 24): Promise<Sub2ApiTimeseriesPoint[]> {
  const body = await getJson<TimeseriesResponse>(`/timeseries?hours=${hours}`)
  return body.points
}

export async function fetchSub2ApiModels(days = 7, limit = 20): Promise<Sub2ApiModelUsage[]> {
  const body = await getJson<ModelsResponse>(`/models?days=${days}&limit=${limit}`)
  return body.models
}

export async function fetchSub2ApiRankings(
  dimension: Sub2ApiRankingDimension = 'user',
  days = 7,
  limit = 20,
): Promise<Sub2ApiRanking[]> {
  const params = new URLSearchParams({
    dimension,
    days: String(days),
    limit: String(limit),
  })
  const body = await getJson<RankingsResponse>(`/rankings?${params.toString()}`)
  return body.rankings
}

export async function fetchSub2ApiAccountQuotas(days = 7): Promise<Sub2ApiAccount[]> {
  const params = new URLSearchParams({ days: String(days) })
  const body = await getJson<AccountsResponse>(`/account-quotas?${params.toString()}`)
  return body.accounts
}

export function fetchSub2ApiEvents({
  page = 1,
  limit = 100,
}: { page?: number; limit?: number } = {}): Promise<Sub2ApiEventsResponse> {
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  })
  return getJson<Sub2ApiEventsResponse>(`/events?${params.toString()}`)
}
