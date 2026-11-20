import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { blogApi } from '../api/blog';

export function useDocumentTitle(title: string) {
  const settingsQuery = useQuery({
    queryKey: ['public', 'settings'],
    queryFn: blogApi.settings,
    staleTime: 5 * 60 * 1000,
  });
  const siteName = settingsQuery.data?.site_name || 'GoPress';
  useEffect(() => {
    const previous = document.title;
    document.title = title ? `${title} · ${siteName}` : `${siteName} · 代码与思考`;
    return () => {
      document.title = previous;
    };
  }, [title, siteName]);
}
