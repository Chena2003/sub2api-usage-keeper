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
