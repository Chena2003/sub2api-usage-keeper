export type Sub2ApiRankingDimension = 'user' | 'api_key' | 'model' | 'account'

export type Sub2ApiQuotaWindow = {
  consumed: number
  limit?: number
  remaining?: number
  ratio?: number
  windowStart?: string
  windowEnd?: string
  refreshAt?: string
  status: string
}

export type Sub2ApiRanking = {
  dimension: Sub2ApiRankingDimension
  name: string
  totalRequests: number
  inputTokens: number
  outputTokens: number
  cacheTokens: number
  totalTokens: number
  actualCost: number
  share: number
}

export type Sub2ApiEvent = {
  id: number
  createdAt: string
  user: string
  apiKey: string
  model: string
  requestedModel: string
  upstreamModel: string
  accountId: number
  accountName: string
  status: string
  inputTokens: number
  outputTokens: number
  cacheTokens: number
  totalTokens: number
  actualCost: number
  durationMs: number
  firstTokenDurationMs?: number
}

export type Sub2ApiEventsResponse = {
  events: Sub2ApiEvent[]
  total: number
  page: number
  limit: number
}

export type Sub2ApiAccountUsage = {
  totalRequests: number
  inputTokens: number
  outputTokens: number
  cacheTokens: number
  totalTokens: number
  actualCost: number
  averageDurationMs: number
}

export type Sub2ApiAccount = {
  id: number
  provider: string
  accountType: string
  displayName: string
  planType?: string
  status: string
  schedulable: boolean
  sessionWindowStatus?: string
  resetAt?: string
  expiresAt?: string
  lastUsedAt?: string
  credentialKeys: string[]
  usage: Sub2ApiAccountUsage
  fiveHourWindow?: Sub2ApiQuotaWindow
  weeklyWindow?: Sub2ApiQuotaWindow
}

export type Sub2ApiOverview = {
  accountCount: number
  activeAccountCount: number
  totalRequests: number
  inputTokens: number
  outputTokens: number
  cacheTokens: number
  totalTokens: number
  actualCost: number
  accountCost: number
  activeUsers: number
}

export type Sub2ApiTimeseriesPoint = {
  bucketStart: string
  totalRequests: number
  inputTokens: number
  outputTokens: number
  cacheCreationTokens: number
  cacheReadTokens: number
  totalCost: number
  actualCost: number
  accountCost: number
  totalDurationMS: number
  activeUsers: number
}

export type Sub2ApiModelUsage = {
  model: string
  requestedModel: string
  upstreamModel: string
  totalRequests: number
  inputTokens: number
  outputTokens: number
  cacheCreationTokens: number
  cacheReadTokens: number
  totalCost: number
  actualCost: number
  averageDurationMS: number
}
