import type { ReactNode } from 'react';
import styles from './Panel.module.scss';

type DotStatus = 'ok' | 'warn' | 'danger';

interface PanelProps {
  eyebrow?: string;
  title?: string;
  actions?: ReactNode;
  dot?: DotStatus;
  padding?: 16 | 20 | 24;
  className?: string;
  as?: 'div' | 'section' | 'article';
  'aria-label'?: string;
  children: ReactNode;
}

interface PanelEmptyProps {
  message?: string;
  children?: ReactNode;
}

interface PanelLoadingProps {
  rows?: number;
}

interface PanelErrorProps {
  message?: string;
  onRetry?: () => void;
  children?: ReactNode;
}

function PanelEmpty({ message = 'No data', children }: PanelEmptyProps) {
  return (
    <div className={styles.empty} role="status">
      <span className={styles.emptyIcon} aria-hidden="true">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M3 6h18M3 12h18M3 18h18" />
        </svg>
      </span>
      {children ?? <span className={styles.emptyMsg}>{message}</span>}
    </div>
  );
}

function PanelLoading({ rows = 3 }: PanelLoadingProps) {
  return (
    <div className={styles.loading} role="status" aria-label="Loading">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className={styles.skeleton} style={{ width: `${70 + (i % 3) * 10}%`, opacity: 1 - i * 0.15 }} />
      ))}
    </div>
  );
}

function PanelError({ message = 'Failed to load', onRetry, children }: PanelErrorProps) {
  return (
    <div className={styles.error} role="alert">
      {children ?? (
        <>
          <span className={styles.errorMsg}>{message}</span>
          {onRetry && (
            <button type="button" className={styles.retryBtn} onClick={onRetry}>
              Retry
            </button>
          )}
        </>
      )}
    </div>
  );
}

export function Panel({
  eyebrow,
  title,
  actions,
  dot,
  padding,
  className,
  as: Tag = 'div',
  children,
  'aria-label': ariaLabel,
}: PanelProps) {
  const paddingVar = padding === 16 ? 'var(--pad-card-tight)' : padding === 24 ? 'var(--pad-card-loose)' : 'var(--pad-card)';

  return (
    <Tag
      className={`${styles.panel} ${className ?? ''}`}
      style={{ '--panel-pad': paddingVar } as React.CSSProperties}
      aria-label={ariaLabel}
    >
      {(eyebrow || title || actions || dot) && (
        <div className={styles.header}>
          <div className={styles.headerText}>
            {eyebrow && (
              <span className={styles.eyebrow} aria-hidden="true">
                {dot && <span className={`${styles.dot} ${styles[`dot_${dot}`]}`} aria-hidden="true" />}
                {eyebrow}
              </span>
            )}
            {title && <h3 className={styles.title}>{title}</h3>}
          </div>
          {actions && <div className={styles.actions}>{actions}</div>}
        </div>
      )}
      <div className={styles.body}>{children}</div>
    </Tag>
  );
}

Panel.Empty = PanelEmpty;
Panel.Loading = PanelLoading;
Panel.Error = PanelError;
