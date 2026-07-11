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

export type TimeRangeParams = { hours?: number; since?: string; until?: string }

function appendTimeRangeParams(params: URLSearchParams, tr: TimeRangeParams): void {
  if (tr.since) params.set('since', tr.since)
  if (tr.until) params.set('until', tr.until)
  if (tr.hours !== undefined) params.set('hours', String(tr.hours))
}

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

export async function fetchSub2ApiAccounts(tr: TimeRangeParams = {}): Promise<Sub2ApiAccount[]> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  const body = await getJson<AccountsResponse>(`/accounts?${params.toString()}`)
  return body.accounts ?? []
}

export function fetchSub2ApiOverview(tr: TimeRangeParams = {}): Promise<Sub2ApiOverview> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  return getJson<Sub2ApiOverview>(`/overview?${params.toString()}`)
}

export async function fetchSub2ApiTimeseries(tr: TimeRangeParams = {}): Promise<Sub2ApiTimeseriesPoint[]> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  const body = await getJson<TimeseriesResponse>(`/timeseries?${params.toString()}`)
  return body.points ?? []
}

export async function fetchSub2ApiModels(tr: TimeRangeParams = {}, limit = 20): Promise<Sub2ApiModelUsage[]> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  params.set('limit', String(limit))
  const body = await getJson<ModelsResponse>(`/models?${params.toString()}`)
  return body.models ?? []
}

export async function fetchSub2ApiRankings(
  dimension: Sub2ApiRankingDimension = 'user',
  tr: TimeRangeParams = {},
  limit = 20,
): Promise<Sub2ApiRanking[]> {
  const params = new URLSearchParams({
    dimension,
    limit: String(limit),
  })
  appendTimeRangeParams(params, tr)
  const body = await getJson<RankingsResponse>(`/rankings?${params.toString()}`)
  return body.rankings ?? []
}

export async function fetchSub2ApiAccountQuotas(tr: TimeRangeParams = {}): Promise<Sub2ApiAccount[]> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  const body = await getJson<AccountsResponse>(`/account-quotas?${params.toString()}`)
  return body.accounts ?? []
}

export function fetchSub2ApiEvents({
  page = 1,
  limit = 100,
  hours,
  since,
  until,
}: { page?: number; limit?: number; hours?: number; since?: string; until?: string } = {}): Promise<Sub2ApiEventsResponse> {
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  })
  if (hours !== undefined) {
    params.set('hours', String(hours))
  }
  if (since) params.set('since', since)
  if (until) params.set('until', until)
  return getJson<Sub2ApiEventsResponse>(`/events?${params.toString()}`)
}

export function fetchSub2ApiHealth(tr: TimeRangeParams = {}): Promise<Sub2ApiServiceHealth> {
  const params = new URLSearchParams()
  appendTimeRangeParams(params, tr)
  return getJson<Sub2ApiServiceHealth>(`/health?${params.toString()}`)
}

export async function fetchSub2ApiRankingTrend(
  dimension: Sub2ApiRankingDimension = 'user',
  tr: TimeRangeParams = {},
  limit = 12,
): Promise<Sub2ApiRankingTrendResponse> {
  const params = new URLSearchParams({
    dimension,
    limit: String(limit),
  })
  appendTimeRangeParams(params, tr)
  return getJson<Sub2ApiRankingTrendResponse>(`/rankings-trend?${params.toString()}`)
}
