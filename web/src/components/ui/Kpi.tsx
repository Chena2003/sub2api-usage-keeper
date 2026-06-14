// Kpi.tsx — KPI 数字展示单元（design.md §3.3-3.4, §6.1）
import type { ReactNode } from 'react';
import styles from './Kpi.module.scss';

interface KpiProps {
  label: string;
  value: ReactNode;
  unit?: string;
  footnote?: string;
  delta?: { value: string; up: boolean };
  loading?: boolean;
  empty?: boolean;
}

export function Kpi({ label, value, unit, footnote, delta, loading, empty }: KpiProps) {
  const displayValue = loading ? null : empty ? '—' : value;

  return (
    <div className={styles.kpi}>
      <span className={styles.label}>{label}</span>
      <div className={styles.valueRow}>
        {loading ? (
          <span className={styles.skeleton} aria-hidden="true" />
        ) : (
          <strong className={styles.value}>{displayValue}</strong>
        )}
        {unit && !loading && !empty && <span className={styles.unit}>{unit}</span>}
        {delta && !loading && !empty && (
          <span className={`${styles.delta} ${delta.up ? styles.deltaUp : styles.deltaDown}`} aria-label={`${delta.up ? 'up' : 'down'} ${delta.value}`}>
            {delta.up ? '↑' : '↓'} {delta.value}
          </span>
        )}
      </div>
      {footnote && !loading && (
        <span className={styles.footnote}>{empty ? '—' : footnote}</span>
      )}
    </div>
  );
}
