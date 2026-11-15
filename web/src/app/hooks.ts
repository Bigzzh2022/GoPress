import { useEffect } from 'react';

export function useDocumentTitle(title: string) {
  useEffect(() => {
    const previous = document.title;
    document.title = title ? `${title} · GoPress` : 'GoPress · 代码与思考';
    return () => {
      document.title = previous;
    };
  }, [title]);
}
