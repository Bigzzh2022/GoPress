import { AlertCircle, ArrowRight, Inbox, RotateCcw } from 'lucide-react';

export function LoadingSkeleton({ cards = 3 }: { cards?: number }) {
  return (
    <div className="skeleton-list" aria-label="内容加载中" aria-busy="true">
      {Array.from({ length: cards }, (_, index) => (
        <div className="skeleton-card" key={index}>
          <div className="skeleton skeleton-cover" />
          <div className="skeleton-body">
            <div className="skeleton skeleton-line skeleton-short" />
            <div className="skeleton skeleton-line skeleton-title" />
            <div className="skeleton skeleton-line" />
            <div className="skeleton skeleton-line skeleton-medium" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="feedback-state" role="alert">
      <AlertCircle size={34} aria-hidden="true" />
      <h2>内容暂时无法抵达</h2>
      <p>{message}</p>
      {onRetry && (
        <button className="button button-outline" onClick={onRetry}>
          <RotateCcw size={16} aria-hidden="true" /> 重新加载
        </button>
      )}
    </div>
  );
}

export function EmptyState({
  title = '这里还没有文章',
  description = '换个关键词或筛选条件，也许会有新的发现。',
  action,
}: {
  title?: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}) {
  return (
    <div className="feedback-state empty-state">
      <Inbox size={38} strokeWidth={1.4} aria-hidden="true" />
      <h2>{title}</h2>
      <p>{description}</p>
      {action && (
        <button className="text-link" onClick={action.onClick}>
          {action.label} <ArrowRight size={15} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}
