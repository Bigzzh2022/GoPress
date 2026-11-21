import { useRef, useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { MessageSquare, Reply, Send, X } from 'lucide-react';
import dayjs from 'dayjs';
import { blogApi } from '../api/blog';
import { getErrorMessage } from '../api/client';
import type { CommentItem, CommentPayload } from '../types';

const IDENTITY_KEY = 'gopress:comment-identity';
const QQ_EMAIL = /^[^\s@]+@(qq\.com|vip\.qq\.com|foxmail\.com)$/i;
const QQ_NUMBER = /^[1-9][0-9]{4,11}$/;

interface Identity {
  nickname: string;
  email: string;
  qq: string;
}

function loadIdentity(): Identity {
  try {
    const raw = window.localStorage.getItem(IDENTITY_KEY);
    if (!raw) return { nickname: '', email: '', qq: '' };
    const parsed = JSON.parse(raw) as Partial<Identity>;
    return { nickname: parsed.nickname || '', email: parsed.email || '', qq: parsed.qq || '' };
  } catch {
    return { nickname: '', email: '', qq: '' };
  }
}

function CommentNode({ comment, depth, onReply, replyingId }: { comment: CommentItem; depth: number; onReply: (comment: CommentItem) => void; replyingId: number | null }) {
  return (
    <li className={`comment-item${replyingId === comment.id ? ' is-replying' : ''}`}>
      <div className="comment-avatar" aria-hidden="true">
        <span>{comment.nickname.slice(0, 1).toUpperCase()}</span>
        {comment.avatar && <img src={comment.avatar} alt="" loading="lazy" onError={(event) => { event.currentTarget.remove(); }} />}
      </div>
      <div className="comment-body">
        <div className="comment-meta">
          <strong>{comment.nickname}</strong>
          <time dateTime={comment.created_at}>{dayjs(comment.created_at).format('YYYY-MM-DD HH:mm')}</time>
        </div>
        <p className="comment-text">{comment.content}</p>
        <div className="comment-actions">
          <button type="button" onClick={() => onReply(comment)}><Reply size={13} /> 回复</button>
        </div>
        {!!comment.replies?.length && (
          <ol className="comment-children">
            {comment.replies.map((reply) => (
              <CommentNode key={reply.id} comment={reply} depth={depth + 1} onReply={onReply} replyingId={replyingId} />
            ))}
          </ol>
        )}
      </div>
    </li>
  );
}

export function CommentSection({ slug }: { slug: string }) {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [replyTo, setReplyTo] = useState<CommentItem | null>(null);
  const [form, setForm] = useState<Identity & { content: string; website: string }>(() => ({ ...loadIdentity(), content: '', website: '' }));
  const [formError, setFormError] = useState('');
  const [notice, setNotice] = useState('');
  const formRef = useRef<HTMLFormElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const commentsQuery = useQuery({
    queryKey: ['public', 'comments', slug, page],
    queryFn: () => blogApi.comments(slug, { page, page_size: 20 }),
  });
  const data = commentsQuery.data;

  const createComment = useMutation({
    mutationFn: (payload: CommentPayload) => blogApi.createComment(slug, payload),
    onSuccess: async (_, payload) => {
      try {
        window.localStorage.setItem(IDENTITY_KEY, JSON.stringify({ nickname: form.nickname.trim(), email: form.email.trim(), qq: form.qq.trim() }));
      } catch {
        // Storage can be unavailable in private mode; the comment itself is already saved.
      }
      setForm((current) => ({ ...current, content: '' }));
      setReplyTo(null);
      setNotice('评论已发布。');
      if (!payload.parent_id) setPage(1);
      await queryClient.invalidateQueries({ queryKey: ['public', 'comments', slug] });
    },
    onError: (error) => setFormError(getErrorMessage(error)),
  });

  const startReply = (comment: CommentItem) => {
    setReplyTo(comment);
    setNotice('');
    setFormError('');
    formRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    window.setTimeout(() => textareaRef.current?.focus({ preventScroll: true }), 220);
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setNotice('');
    if (form.website.trim() !== '') {
      setFormError('评论提交失败，请稍后再试');
      return;
    }
    if (!form.nickname.trim()) {
      setFormError('请填写昵称');
      return;
    }
    if (!QQ_EMAIL.test(form.email.trim())) {
      setFormError('请填写 QQ 邮箱（@qq.com、@vip.qq.com 或 @foxmail.com）');
      return;
    }
    if (form.qq.trim() && !QQ_NUMBER.test(form.qq.trim())) {
      setFormError('QQ 号需为 5-12 位数字');
      return;
    }
    if (!form.content.trim()) {
      setFormError('请填写评论内容');
      return;
    }
    setFormError('');
    createComment.mutate({
      nickname: form.nickname.trim(),
      email: form.email.trim(),
      qq: form.qq.trim() || undefined,
      content: form.content.trim(),
      parent_id: replyTo?.id ?? null,
      website: '',
    });
  };

  const totalPages = data?.total_pages ?? 0;

  return (
    <section className="comments-section" id="comments">
      <div className="page-container">
        <div className="comments-heading">
          <div>
            <p className="eyebrow">DISCUSSION</p>
            <h2><MessageSquare size={26} aria-hidden="true" /> 评论 <small>{data?.comment_count ?? 0}</small></h2>
          </div>
          <span>无需审核，发布后即时显示</span>
        </div>

        <form className="comment-form" ref={formRef} onSubmit={submit} noValidate>
          <div className="comment-form-head">
            <div>
              <p className="eyebrow">JOIN THE DISCUSSION</p>
              <h3>{replyTo ? `回复 @${replyTo.nickname}` : '写下你的评论'}</h3>
            </div>
            {replyTo && (
              <button type="button" className="comment-cancel-reply" onClick={() => { setReplyTo(null); setFormError(''); }}>
                <X size={14} /> 取消回复
              </button>
            )}
          </div>
          <div className="comment-form-grid">
            <label><span>昵称</span><input value={form.nickname} maxLength={40} onChange={(event) => setForm((current) => ({ ...current, nickname: event.target.value }))} placeholder="怎么称呼你" autoComplete="nickname" /></label>
            <label><span>QQ 邮箱</span><input type="email" value={form.email} maxLength={191} onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))} placeholder="123456@qq.com" autoComplete="email" /></label>
            <label><span>QQ 号（可选）</span><input inputMode="numeric" value={form.qq} maxLength={12} onChange={(event) => setForm((current) => ({ ...current, qq: event.target.value }))} placeholder="用于显示 QQ 头像" /></label>
          </div>
          <label className="comment-textarea-label">
            <span>评论内容</span>
            <textarea ref={textareaRef} value={form.content} maxLength={2000} rows={5} onChange={(event) => setForm((current) => ({ ...current, content: event.target.value }))} placeholder="保持友善，认真交流。" />
          </label>
          <input className="comment-honeypot" tabIndex={-1} autoComplete="off" aria-hidden="true" value={form.website} onChange={(event) => setForm((current) => ({ ...current, website: event.target.value }))} />
          {formError && <p className="comment-form-message is-error" role="alert">{formError}</p>}
          {notice && <p className="comment-form-message is-notice" role="status">{notice}</p>}
          <div className="comment-form-foot">
            <span>邮箱不会公开，QQ 头像来自 QQ 官方接口。</span>
            <button type="submit" className="button button-dark" disabled={createComment.isPending}>
              {createComment.isPending ? '发布中…' : <><Send size={15} /> 发表评论</>}
            </button>
          </div>
        </form>

        <div className="comment-list-area">
          {commentsQuery.isLoading ? (
            <p className="comment-state">正在加载评论…</p>
          ) : commentsQuery.isError ? (
            <div className="comment-state">
              <p>{getErrorMessage(commentsQuery.error, '评论加载失败。')}</p>
              <button type="button" className="text-link" onClick={() => commentsQuery.refetch()}>重新加载</button>
            </div>
          ) : !data?.items.length ? (
            <p className="comment-state">还没有评论，来说点什么吧。</p>
          ) : (
            <ol className="comment-list">
              {data.items.map((comment) => (
                <CommentNode key={comment.id} comment={comment} depth={0} onReply={startReply} replyingId={replyTo?.id ?? null} />
              ))}
            </ol>
          )}
        </div>

        {totalPages > 1 && (
          <nav className="comment-pagination" aria-label="评论分页">
            <button type="button" className="button button-outline" disabled={page <= 1 || commentsQuery.isFetching} onClick={() => setPage((current) => Math.max(1, current - 1))}>上一页</button>
            <span>第 {data?.page} / {totalPages} 页</span>
            <button type="button" className="button button-outline" disabled={page >= totalPages || commentsQuery.isFetching} onClick={() => setPage((current) => current + 1)}>下一页</button>
          </nav>
        )}
      </div>
    </section>
  );
}
