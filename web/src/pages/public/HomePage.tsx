import { useQuery } from '@tanstack/react-query';
import { ArrowDown, ArrowRight, Asterisk, MoveRight } from 'lucide-react';
import { Link, useSearchParams } from 'react-router-dom';
import { blogApi } from '../../api/blog';
import { getErrorMessage } from '../../api/client';
import { EmptyState, ErrorState, LoadingSkeleton } from '../../components/Feedback';
import { Pagination } from '../../components/Pagination';
import { PostCard } from '../../components/PostCard';
import { useDocumentTitle } from '../../app/hooks';

export default function HomePage() {
  useDocumentTitle('');
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Math.max(1, Number(searchParams.get('page')) || 1);
  const postsQuery = useQuery({
    queryKey: ['public', 'posts', { page, page_size: 7 }],
    queryFn: () => blogApi.posts({ page, page_size: 7 }),
  });
  const categoriesQuery = useQuery({ queryKey: ['public', 'categories'], queryFn: blogApi.categories });
  const settingsQuery = useQuery({
    queryKey: ['public', 'settings'],
    queryFn: blogApi.settings,
    staleTime: 5 * 60 * 1000,
  });
  const siteDescription = settingsQuery.data?.site_description
    || '关于 Go、分布式系统与工程实践的独立博客。不追逐喧嚣，只记录值得反复阅读的技术思考。';
  const posts = postsQuery.data?.items || [];
  const featured = page === 1 ? posts[0] : undefined;
  const latest = featured ? posts.slice(1) : posts;

  const changePage = (next: number) => {
    setSearchParams(next > 1 ? { page: String(next) } : {});
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <>
      <section className="home-hero page-container">
        <div className="hero-copy">
          <div className="eyebrow"><span /> INDEPENDENT TECH JOURNAL · 2026</div>
          <h1>在代码之间，<br />寻找<span>清晰的答案。</span></h1>
          <p className="hero-description">{siteDescription}</p>
          <div className="hero-actions">
            <Link to="/posts" className="button button-dark">浏览全部文章 <ArrowRight size={17} /></Link>
            <Link to="/about" className="text-link">认识作者 <MoveRight size={18} /></Link>
          </div>
          <div className="hero-note"><Asterisk size={17} /><span>保持好奇，拒绝浅薄。<br /><small>EST. 2024 · WRITTEN WITH CARE</small></span></div>
        </div>
        <div className="hero-feature">
          <div className="section-kicker"><span>本期精选</span><span>EDITOR'S PICK</span></div>
          {postsQuery.isLoading ? <LoadingSkeleton cards={1} /> : postsQuery.isError ? (
            <ErrorState message={getErrorMessage(postsQuery.error)} onRetry={() => postsQuery.refetch()} />
          ) : featured ? <PostCard post={featured} featured /> : (
            <EmptyState title="第一篇文章正在酝酿" description="好的内容值得等待，请先逛逛关于页面。" />
          )}
          <a href="#latest" className="scroll-cue"><ArrowDown size={15} /> SCROLL TO EXPLORE</a>
        </div>
      </section>

      <div className="ink-divider"><span>GOPRESS</span><p>CODE · SYSTEMS · CRAFT · CURIOSITY ·</p><span>GOPRESS</span></div>

      <section id="latest" className="latest-section page-container">
        <div className="section-heading">
          <div><p className="eyebrow">FRESH FROM THE KEYBOARD</p><h2>最新发布</h2></div>
          <Link to="/posts" className="text-link">查看全部 <ArrowRight size={16} /></Link>
        </div>
        {postsQuery.isLoading ? <LoadingSkeleton cards={3} /> : postsQuery.isError ? null : latest.length ? (
          <div className="post-grid">{latest.map((post) => <PostCard key={post.id} post={post} />)}</div>
        ) : featured ? <EmptyState title="只有一篇精选" description="更多文章正在路上。" /> : null}
        {postsQuery.data && <Pagination page={page} totalPages={postsQuery.data.total_pages} onChange={changePage} />}
      </section>

      {!!categoriesQuery.data?.length && (
        <section className="topic-section">
          <div className="page-container">
            <div className="section-heading light"><div><p className="eyebrow">EXPLORE BY TOPIC</p><h2>从一个主题，深入下去</h2></div></div>
            <div className="topic-grid">
              {categoriesQuery.data.slice(0, 6).map((category, index) => (
                <Link key={category.id} to={`/posts?category=${encodeURIComponent(category.slug)}`} className="topic-card">
                  <span>0{index + 1}</span><div><h3>{category.name}</h3><p>{category.description || '持续记录这个领域的实践与思考。'}</p><small>{category.post_count} 篇文章</small></div><ArrowRight size={20} />
                </Link>
              ))}
            </div>
          </div>
        </section>
      )}
    </>
  );
}
