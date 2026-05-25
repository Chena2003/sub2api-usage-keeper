import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Doughnut } from 'react-chartjs-2'
import type { ChartData, ChartOptions } from 'chart.js'
import type { Sub2ApiModelUsage } from '@/lib/sub2apiTypes'
import styles from '@/pages/UsagePage.module.scss'
import designStyles from './Sub2ApiDesign.module.scss'

type Sub2ApiAnalysisPanelProps = {
  models: Sub2ApiModelUsage[]
}

const formatNumber = (value: number) => new Intl.NumberFormat().format(value)
const formatCost = (value: number) => `$${value.toFixed(4)}`

const modelTokenTotal = (model: Sub2ApiModelUsage) => (
  model.inputTokens + model.outputTokens + model.cacheCreationTokens + model.cacheReadTokens
)

export function Sub2ApiAnalysisPanel({ models }: Sub2ApiAnalysisPanelProps) {
  const { t } = useTranslation()
  const hasData = models.length > 0
  const topModels = models.slice(0, 5)

  const modelChartData = useMemo((): ChartData<'doughnut'> | null => {
    if (topModels.length === 0) return null
    const colors = ['#34d399', '#3b82f6', '#f59e0b', '#ec4899', '#8b5cf6']
    return {
      labels: topModels.map((m) => m.model || m.requestedModel || 'unknown'),
      datasets: [
        {
          data: topModels.map(modelTokenTotal),
          backgroundColor: colors.slice(0, topModels.length),
          borderWidth: 0,
          hoverOffset: 8,
        },
      ],
    }
  }, [topModels])

  const doughnutOptions: ChartOptions<'doughnut'> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: '62%',
    plugins: { legend: { display: false }, tooltip: { enabled: true } },
  }

  if (!hasData) {
    return <div className={styles.hint}>{t('usage_stats.sub2api_no_data')}</div>
  }

  return (
    <div className={styles.settingsSections}>
      <div className={designStyles.statsGrid} style={{ gridTemplateColumns: '1fr', padding: '24px', background: 'var(--bg-primary)' }}>
        <h3 className={designStyles.sectionTitle} style={{ marginBottom: '16px' }}>{t('usage_stats.sub2api_model', 'Model Usage Analysis')}</h3>
        
        {modelChartData && (
          <div style={{ display: 'flex', gap: '32px', flexWrap: 'wrap', marginBottom: '48px' }}>
            <div style={{ flex: '1', minWidth: '300px' }}>
              <div style={{ height: '260px', position: 'relative' }}>
                <Doughnut data={modelChartData} options={doughnutOptions} />
              </div>
            </div>
            
            <div style={{ flex: '1', minWidth: '300px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <h4 className={styles.chartTitleSmall} style={{ marginBottom: '16px' }}>Top Models (by token volume)</h4>
              <div className={styles.chartDoughnutLegend} style={{ display: 'flex', flexDirection: 'column', gap: '12px', padding: 0 }}>
                {topModels.map((m, i) => {
                  const colors = ['#34d399', '#3b82f6', '#f59e0b', '#ec4899', '#8b5cf6']
                  return (
                    <div key={m.model} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <span className={styles.chartLegendItem}>
                        <span className={styles.chartLegendDot} style={{ background: colors[i] || '#9ca3af' }} />
                        {m.model || m.requestedModel || 'unknown'}
                      </span>
                      <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        {formatNumber(modelTokenTotal(m))} tokens
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        )}

        <div style={{ overflowX: 'auto' }}>
          <table className={styles.dataTable} style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--text-secondary)', fontSize: '12px' }}>
                <th style={{ padding: '12px 16px', fontWeight: 600 }}>Model Name</th>
                <th style={{ padding: '12px 16px', fontWeight: 600, textAlign: 'right' }}>Requests</th>
                <th style={{ padding: '12px 16px', fontWeight: 600, textAlign: 'right' }}>Input Tokens</th>
                <th style={{ padding: '12px 16px', fontWeight: 600, textAlign: 'right' }}>Output Tokens</th>
                <th style={{ padding: '12px 16px', fontWeight: 600, textAlign: 'right' }}>Total Cost</th>
              </tr>
            </thead>
            <tbody>
              {models.map((model) => (
                <tr key={model.model || model.requestedModel} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '13px', transition: 'background 0.2s' }}>
                  <td style={{ padding: '16px', fontWeight: 500, color: 'var(--text-primary)' }}>
                    {model.model || model.requestedModel || 'unknown'}
                  </td>
                  <td style={{ padding: '16px', textAlign: 'right', color: 'var(--text-secondary)' }}>
                    {formatNumber(model.totalRequests)}
                  </td>
                  <td style={{ padding: '16px', textAlign: 'right', color: 'var(--text-secondary)' }}>
                    {formatNumber(model.inputTokens)}
                  </td>
                  <td style={{ padding: '16px', textAlign: 'right', color: 'var(--text-secondary)' }}>
                    {formatNumber(model.outputTokens)}
                  </td>
                  <td style={{ padding: '16px', textAlign: 'right', fontWeight: 600, color: 'var(--text-primary)' }}>
                    {formatCost(model.actualCost)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
