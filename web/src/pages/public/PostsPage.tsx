import { useEffect, useState, type FormEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, SlidersHorizontal, X } from 'lucide-react';
import { useSearchParams } from 'react-router-dom';
import { blogApi } from '../../api/blog';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { EmptyState, ErrorState, LoadingSkeleton } from '../../components/Feedback';
import { Pagination } from '../../components/Pagination';
import { PostCard } from '../../components/PostCard';

export default function PostsPage() {
  useDocumentTitle('全部文章');
  const [searchParams, setSearchParams] = useSearchParams();
  const keyword = searchParams.get('keyword') || '';
  const category = searchParams.get('category') || '';
  const tag = searchParams.get('tag') || '';
  const page = Math.max(1, Number(searchParams.get('page')) || 1);
  const [searchInput, setSearchInput] = useState(keyword);
  useEffect(() => setSearchInput(keyword), [keyword]);

  const categoriesQuery = useQuery({ queryKey: ['public', 'categories'], queryFn: blogApi.categories });
  const tagsQuery = useQuery({ queryKey: ['public', 'tags'], queryFn: blogApi.tags });
  const postsQuery = useQuery({
    queryKey: ['public', 'posts', { page, page_size: 6, keyword, category, tag }],
    queryFn: () => blogApi.posts({ page, page_size: 6, keyword, category, tag }),
  });

  const updateFilters = (changes: Record<string, string>) => {
    const next = new URLSearchParams(searchParams);
    Object.entries(changes).forEach(([key, value]) => value ? next.set(key, value) : next.delete(key));
    if (!('page' in changes)) next.delete('page');
    setSearchParams(next);
  };
  const submitSearch = (event: FormEvent) => {
    event.preventDefault();
    updateFilters({ keyword: searchInput.trim() });
  };
  const reset = () => {
    setSearchInput('');
    setSearchParams({});
  };
  const activeCategory = categoriesQuery.data?.find((item) => item.slug === category)?.name;
  const activeTag = tagsQuery.data?.find((item) => item.slug === tag)?.name;

  return (
    <div className="posts-page page-container">
      <header className="archive-header">
        <p className="eyebrow">THE ARCHIVE · NOTES ON BUILDING SOFTWARE</p>
        <h1>所有文章<span>。</span></h1>
        <p>从一行代码到一个完整系统，这里是我的学习、实践与复盘。</p>
      </header>

      <section className="filter-panel" aria-label="文章筛选">
        <form className="search-form" onSubmit={submitSearch} role="search">
          <Search size={20} aria-hidden="true" />
          <label htmlFor="post-search" className="sr-only">搜索文章</label>
          <input id="post-search" value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="搜索标题、摘要或内容…" autoComplete="off" />
          {searchInput && <button type="button" className="search-clear" onClick={() => setSearchInput('')} aria-label="清空搜索框"><X size={16} /></button>}
          <button className="button button-dark" type="submit">搜索</button>
        </form>
        <div className="filter-row">
          <span className="filter-label"><SlidersHorizontal size={15} /> 分类</span>
          <div className="filter-chips">
            <button className={!category ? 'active' : ''} onClick={() => updateFilters({ category: '' })}>全部</button>
            {categoriesQuery.data?.map((item) => <button key={item.id} className={category === item.slug ? 'active' : ''} onClick={() => updateFilters({ category: item.slug })}>{item.name}<small>{item.post_count}</small></button>)}
          </div>
        </div>
        {!!tagsQuery.data?.length && (
          <div className="filter-row">
            <span className="filter-label">标签</span>
            <div className="filter-chips tag-chips">
              {tagsQuery.data.map((item) => <button key={item.id} className={tag === item.slug ? 'active' : ''} onClick={() => updateFilters({ tag: tag === item.slug ? '' : item.slug })}>#{item.name}</button>)}
            </div>
          </div>
        )}
      </section>

      <div className="results-bar" aria-live="polite">
        <span>{postsQuery.data ? <>共 <strong>{postsQuery.data.total}</strong> 篇文章</> : '正在整理文章…'}</span>
        <div>{keyword && <button onClick={() => updateFilters({ keyword: '' })}>关键词：{keyword}<X size={13} /></button>}{activeCategory && <button onClick={() => updateFilters({ category: '' })}>{activeCategory}<X size={13} /></button>}{activeTag && <button onClick={() => updateFilters({ tag: '' })}>#{activeTag}<X size={13} /></button>}</div>
      </div>

      {postsQuery.isLoading ? <LoadingSkeleton cards={6} /> : postsQuery.isError ? (
        <ErrorState message={getErrorMessage(postsQuery.error)} onRetry={() => postsQuery.refetch()} />
      ) : postsQuery.data?.items.length ? (
        <div className="post-grid archive-grid">{postsQuery.data.items.map((post) => <PostCard key={post.id} post={post} />)}</div>
      ) : <EmptyState action={{ label: '清除全部筛选', onClick: reset }} />}
      {postsQuery.data && <Pagination page={page} totalPages={postsQuery.data.total_pages} onChange={(next) => { updateFilters({ page: next > 1 ? String(next) : '' }); window.scrollTo({ top: 380, behavior: 'smooth' }); }} />}
    </div>
  );
}
