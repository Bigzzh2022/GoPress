import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, ArrowRight, CalendarDays, Clock3, List, UserRound } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import dayjs from 'dayjs';
import { blogApi } from '../../api/blog';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { CoverArt } from '../../components/CoverArt';
import { ErrorState, LoadingSkeleton } from '../../components/Feedback';
import { MarkdownContent } from '../../components/MarkdownContent';
import { PostCard } from '../../components/PostCard';
import { extractHeadings, readingTime } from '../../utils/markdown';

export default function PostDetailPage() {
  const { slug = '' } = useParams();
  const [progress, setProgress] = useState(0);
  const postQuery = useQuery({ queryKey: ['public', 'post', slug], queryFn: () => blogApi.post(slug), enabled: !!slug });
  const relatedQuery = useQuery({ queryKey: ['public', 'related', slug], queryFn: () => blogApi.related(slug), enabled: !!slug });
  const post = postQuery.data;
  useDocumentTitle(post?.title || '文章');
  const toc = extractHeadings(post?.content_md || '');

  useEffect(() => {
    const update = () => {
      const distance = document.documentElement.scrollHeight - window.innerHeight;
      setProgress(distance > 0 ? Math.min(100, (window.scrollY / distance) * 100) : 0);
    };
    update();
    window.addEventListener('scroll', update, { passive: true });
    window.addEventListener('resize', update);
    return () => { window.removeEventListener('scroll', update); window.removeEventListener('resize', update); };
  }, [post]);

  if (postQuery.isLoading) return <div className="detail-loading page-container"><LoadingSkeleton cards={2} /></div>;
  if (postQuery.isError || !post) return <div className="page-container detail-loading"><ErrorState message={getErrorMessage(postQuery.error, '文章不存在或已下线。')} onRetry={() => postQuery.refetch()} /><Link to="/posts" className="text-link"><ArrowLeft size={16} /> 返回文章列表</Link></div>;

  const date = post.published_at || post.created_at;
  return (
    <article className="post-detail">
      <div className="reading-progress" style={{ width: `${progress}%` }} />
      <header className="article-header page-container">
        <Link to="/posts" className="back-link"><ArrowLeft size={16} /> 返回文章</Link>
        <div className="article-header-grid">
          <div>
            <div className="article-category">{post.category?.name || '技术随笔'}<span />GOPRESS JOURNAL</div>
            <h1>{post.title}</h1>
            <p className="article-deck">{post.excerpt}</p>
            <div className="article-meta">
              <span><UserRound size={16} />{post.author?.name || 'GoPress 作者'}</span>
              <span><CalendarDays size={16} /><time dateTime={date}>{dayjs(date).format('YYYY 年 M 月 D 日')}</time></span>
              <span><Clock3 size={16} />约 {readingTime(post.content_md)} 分钟阅读</span>
            </div>
          </div>
          <CoverArt seed={post.id} label={post.category?.name || 'ENGINEERING'} large />
        </div>
      </header>

      <div className="article-layout page-container">
        <aside className="article-aside">
          <div className="toc">
            <div className="toc-title"><List size={16} /> 本文目录</div>
            {toc.length ? <nav aria-label="文章目录">{toc.map((item, index) => <a key={`${item.id}-${index}`} className={item.level === 3 ? 'toc-sub' : ''} href={`#${encodeURIComponent(item.id)}`}>{item.text}</a>)}</nav> : <p>本文没有章节标题</p>}
          </div>
          <div className="article-tags">{post.tags?.map((tag) => <Link key={tag.id} to={`/posts?tag=${encodeURIComponent(tag.slug)}`}>#{tag.name}</Link>)}</div>
        </aside>
        {!!toc.length && <details className="mobile-toc"><summary><List size={16} /> 查看本文目录</summary><nav>{toc.map((item, index) => <a key={`${item.id}-${index}`} href={`#${encodeURIComponent(item.id)}`}>{item.text}</a>)}</nav></details>}
        <MarkdownContent content={post.content_md} />
      </div>

      <footer className="article-end page-container">
        <div className="article-end-mark">G<span>·</span>P</div>
        <div><p>感谢读到这里。</p><span>如果这篇文章带来了启发，欢迎将它分享给同样热爱技术的人。</span></div>
        <button className="button button-outline" onClick={() => navigator.clipboard?.writeText(window.location.href)}>复制文章链接</button>
      </footer>

      {!!relatedQuery.data?.length && (
        <section className="related-section">
          <div className="page-container">
            <div className="section-heading"><div><p className="eyebrow">KEEP READING</p><h2>继续阅读</h2></div><Link to="/posts" className="text-link">全部文章 <ArrowRight size={16} /></Link></div>
            <div className="post-grid">{relatedQuery.data.slice(0, 3).map((item) => <PostCard key={item.id} post={item} />)}</div>
          </div>
        </section>
      )}
    </article>
  );
}
