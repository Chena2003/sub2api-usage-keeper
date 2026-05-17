import { ApiError, apiPath } from './api'
import type { Sub2ApiAccount, Sub2ApiModelUsage, Sub2ApiOverview, Sub2ApiTimeseriesPoint } from './sub2apiTypes'

type AccountsResponse = { accounts: Sub2ApiAccount[] }
type TimeseriesResponse = { points: Sub2ApiTimeseriesPoint[] }
type ModelsResponse = { models: Sub2ApiModelUsage[] }

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
