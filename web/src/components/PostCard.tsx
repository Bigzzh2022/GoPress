import { ArrowUpRight, Clock3 } from 'lucide-react';
import { Link } from 'react-router-dom';
import dayjs from 'dayjs';
import type { PostSummary } from '../types';
import { CoverArt } from './CoverArt';
import { readingTime } from '../utils/markdown';

export function PostCard({ post, featured = false }: { post: PostSummary; featured?: boolean }) {
  const date = post.published_at || post.created_at;
  return (
    <article className={`post-card ${featured ? 'post-card-featured' : ''}`}>
      <Link to={`/posts/${post.slug}`} className="post-card-cover" tabIndex={-1} aria-hidden="true">
        <CoverArt seed={post.id} label={post.category?.name || 'ENGINEERING'} large={featured} />
      </Link>
      <div className="post-card-content">
        <div className="post-card-meta">
          <span>{post.category?.name || '技术随笔'}</span>
          <time dateTime={date}>{dayjs(date).format('YYYY.MM.DD')}</time>
        </div>
        <h3>
          <Link to={`/posts/${post.slug}`}>
            {post.title}<ArrowUpRight size={featured ? 25 : 19} aria-hidden="true" />
          </Link>
        </h3>
        <p>{post.excerpt || '一篇关于代码、系统与工程实践的记录。'}</p>
        <div className="post-card-footer">
          <div className="tag-list" aria-label="文章标签">
            {post.tags?.slice(0, featured ? 3 : 2).map((tag) => (
              <Link key={tag.id} to={`/posts?tag=${encodeURIComponent(tag.slug)}`}>#{tag.name}</Link>
            ))}
          </div>
          <span className="reading-time"><Clock3 size={14} aria-hidden="true" />{readingTime('content_md' in post ? String(post.content_md) : post.excerpt)} 分钟</span>
        </div>
      </div>
    </article>
  );
}
