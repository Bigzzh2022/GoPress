import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationProps {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
}

export function Pagination({ page, totalPages, onChange }: PaginationProps) {
  if (totalPages <= 1) return null;
  const pages = Array.from({ length: totalPages }, (_, index) => index + 1)
    .filter((item) => item === 1 || item === totalPages || Math.abs(item - page) <= 1);

  return (
    <nav className="pagination" aria-label="文章分页">
      <button disabled={page <= 1} onClick={() => onChange(page - 1)} aria-label="上一页">
        <ChevronLeft size={18} aria-hidden="true" />
      </button>
      {pages.map((item, index) => (
        <span key={item} className="pagination-item">
          {index > 0 && item - pages[index - 1] > 1 && <span className="pagination-gap">…</span>}
          <button
            className={item === page ? 'active' : ''}
            onClick={() => onChange(item)}
            aria-current={item === page ? 'page' : undefined}
            aria-label={`第 ${item} 页`}
          >
            {String(item).padStart(2, '0')}
          </button>
        </span>
      ))}
      <button disabled={page >= totalPages} onClick={() => onChange(page + 1)} aria-label="下一页">
        <ChevronRight size={18} aria-hidden="true" />
      </button>
    </nav>
  );
}
