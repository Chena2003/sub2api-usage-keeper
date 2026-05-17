import { useEffect } from 'react'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useSub2ApiDashboardStore } from '../stores/useSub2ApiDashboardStore'
import styles from './Sub2ApiDashboardPage.module.scss'

function formatNumber(value: number): string {
  return new Intl.NumberFormat('zh-CN').format(value)
}

function formatCost(value: number): string {
  return `$${value.toFixed(4)}`
}

function formatDateTime(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN')
}

function modelTokens(model: { inputTokens: number; outputTokens: number; cacheCreationTokens: number; cacheReadTokens: number }): number {
  return model.inputTokens + model.outputTokens + model.cacheCreationTokens + model.cacheReadTokens
}

export function Sub2ApiDashboardPage() {
  const { accounts, overview, points, models, loading, error, refresh } = useSub2ApiDashboardStore()

  useEffect(() => {
    void refresh()
  }, [refresh])

  return (
    <main className={styles.page}>
      <header className={styles.hero}>
        <div>
          <p className={styles.kicker}>Sub2API Resource Pool</p>
          <h1>账号额度与调用统计</h1>
          <p>展示所有上游账号的脱敏状态、总体消耗和近 24 小时调用趋势。</p>
        </div>
        <button type="button" onClick={() => void refresh()} disabled={loading}>
          {loading ? '刷新中' : '刷新'}
        </button>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <section className={styles.cards} aria-label="Sub2API overview">
        <article><span>总账号</span><strong>{formatNumber(overview?.accountCount ?? 0)}</strong></article>
        <article><span>可用账号</span><strong>{formatNumber(overview?.activeAccountCount ?? 0)}</strong></article>
        <article><span>7 天请求</span><strong>{formatNumber(overview?.totalRequests ?? 0)}</strong></article>
        <article><span>7 天 Tokens</span><strong>{formatNumber(overview?.totalTokens ?? 0)}</strong></article>
        <article><span>实际成本</span><strong>{formatCost(overview?.actualCost ?? 0)}</strong></article>
      </section>

      <section className={styles.panel}>
        <div className={styles.panelHeader}>
          <h2>近 24 小时请求趋势</h2>
        </div>
        <div className={styles.chart}>
          <ResponsiveContainer width="100%" height={280}>
            <AreaChart data={points}>
              <defs>
                <linearGradient id="sub2apiRequests" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#7c3aed" stopOpacity={0.45} />
                  <stop offset="95%" stopColor="#7c3aed" stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(148, 163, 184, 0.18)" />
              <XAxis dataKey="bucketStart" tickFormatter={(value) => new Date(String(value)).getHours().toString().padStart(2, '0')} stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" />
              <Tooltip />
              <Area type="monotone" dataKey="totalRequests" stroke="#7c3aed" fill="url(#sub2apiRequests)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </section>

      <section className={styles.panel}>
        <div className={styles.panelHeader}>
          <h2>上游账号</h2>
          <span>{accounts.length} accounts</span>
        </div>
        <div className={styles.tableWrap}>
          <table>
            <thead>
              <tr>
                <th>账号</th>
                <th>Provider</th>
                <th>Plan</th>
                <th>Status</th>
                <th>7 天请求</th>
                <th>7 天 Tokens</th>
                <th>Reset</th>
              </tr>
            </thead>
            <tbody>
              {accounts.map((account) => (
                <tr key={account.id}>
                  <td>{account.displayName}</td>
                  <td>{account.provider}</td>
                  <td>{account.planType || '-'}</td>
                  <td><span className={styles.status}>{account.status}</span></td>
                  <td>{formatNumber(account.usage.totalRequests)}</td>
                  <td>{formatNumber(account.usage.totalTokens)}</td>
                  <td>{formatDateTime(account.resetAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className={styles.panel}>
        <div className={styles.panelHeader}>
          <h2>模型排行</h2>
        </div>
        <div className={styles.modelGrid}>
          {models.map((model) => (
            <article key={`${model.model}-${model.requestedModel}-${model.upstreamModel}`}>
              <strong>{model.model || model.requestedModel || model.upstreamModel || 'unknown'}</strong>
              <span>{formatNumber(model.totalRequests)} requests</span>
              <span>{formatNumber(modelTokens(model))} tokens</span>
            </article>
          ))}
        </div>
      </section>
    </main>
  )
}
