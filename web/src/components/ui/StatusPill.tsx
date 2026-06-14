// StatusPill.tsx — 状态 pill（design.md §3.2：非颜色线索必须提供）
import styles from './StatusPill.module.scss';

type PillVariant = 'ok' | 'warn' | 'danger' | 'neutral' | 'info';

interface StatusPillProps {
  variant: PillVariant;
  label: string;
  dot?: boolean;
}

export function StatusPill({ variant, label, dot = true }: StatusPillProps) {
  return (
    <span className={`${styles.pill} ${styles[variant]}`} role="status">
      {dot && <span className={styles.dot} aria-hidden="true" />}
      {label}
    </span>
  );
}
