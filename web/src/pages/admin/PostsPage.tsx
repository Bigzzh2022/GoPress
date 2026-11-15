import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Input, Select, Space, Table, Tag } from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { PostQuery, PostStatus, PostSummary } from '../../types';

const statusText: Record<PostStatus, string> = { published: '已发布', draft: '草稿' };

export default function AdminPostsPage() {
  useDocumentTitle('文章管理');
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<PostQuery>({ page: 1, page_size: 10, keyword: '', status: '', category: '' });
  const [keyword, setKeyword] = useState('');
  const postsQuery = useQuery({ queryKey: ['admin', 'posts', filters], queryFn: () => adminApi.posts(filters), placeholderData: keepPreviousData });
  const categoriesQuery = useQuery({ queryKey: ['admin', 'categories'], queryFn: adminApi.categories });

  const invalidatePosts = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['admin', 'posts'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'post'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'related'] }),
    ]);
  };
  const deleteMutation = useMutation({
    mutationFn: adminApi.deletePost,
    onSuccess: async () => { message.success('文章已删除'); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: PostStatus }) => adminApi.setPostStatus(id, status),
    onSuccess: async (_, variables) => { message.success(`状态已更新为「${statusText[variables.status]}」`); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const changeFilter = (changes: PostQuery) => setFilters((current) => ({ ...current, page: 1, ...changes }));
  const confirmDelete = (post: PostSummary) => modal.confirm({
    title: '确定删除这篇文章？',
    content: `「${post.title}」将被永久删除，此操作无法撤销。`,
    okText: '确认删除', cancelText: '取消', okButtonProps: { danger: true },
    onOk: () => deleteMutation.mutateAsync(post.id).then(() => undefined),
  });

  return (
    <div className="admin-page">
      <div className="admin-page-heading"><div><p>CONTENT LIBRARY</p><h1>文章管理</h1><span>创作、编辑并管理你的全部文章。</span></div><Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/admin/posts/new')}>写新文章</Button></div>
      <Card className="admin-table-card">
        <div className="table-filters">
          <Input value={keyword} onChange={(event) => setKeyword(event.target.value)} onPressEnter={() => changeFilter({ keyword })} placeholder="搜索文章标题或摘要" prefix={<SearchOutlined />} allowClear />
          <Select value={filters.status} onChange={(status) => changeFilter({ status })} options={[{ value: '', label: '全部状态' }, { value: 'published', label: '已发布' }, { value: 'draft', label: '草稿' }]} aria-label="筛选文章状态" />
          <Select value={filters.category} onChange={(category) => changeFilter({ category })} options={[{ value: '', label: '全部分类' }, ...(categoriesQuery.data?.map((item) => ({ value: item.slug, label: item.name })) || [])]} aria-label="筛选文章分类" />
          <Button type="primary" ghost icon={<SearchOutlined />} onClick={() => changeFilter({ keyword })}>查询</Button>
        </div>
        <Table<PostSummary>
          rowKey="id" dataSource={postsQuery.data?.items} loading={postsQuery.isLoading} scroll={{ x: 920 }}
          pagination={{ current: postsQuery.data?.page || filters.page, pageSize: filters.page_size, total: postsQuery.data?.total, showSizeChanger: true, showTotal: (total) => `共 ${total} 篇`, onChange: (page, page_size) => setFilters((current) => ({ ...current, page, page_size })) }}
          columns={[
            { title: '文章', dataIndex: 'title', width: 310, render: (title: string, post) => <button className="admin-post-cell" onClick={() => navigate(`/admin/posts/${post.id}/edit`)}><strong>{title}</strong><span>/{post.slug}</span></button> },
            { title: '分类', dataIndex: 'category', width: 110, render: (_, post) => post.category?.name || <span className="muted">未分类</span> },
            { title: '标签', dataIndex: 'tags', width: 170, render: (tags: PostSummary['tags']) => <div className="table-tags">{tags?.slice(0, 2).map((tag) => <Tag key={tag.id}>{tag.name}</Tag>)}{tags?.length > 2 && <span>+{tags.length - 2}</span>}</div> },
            { title: '状态', dataIndex: 'status', width: 120, render: (status: PostStatus, post) => <Select<PostStatus> size="small" value={status} loading={statusMutation.isPending && statusMutation.variables?.id === post.id} onChange={(next) => statusMutation.mutate({ id: post.id, status: next })} options={Object.entries(statusText).map(([value, label]) => ({ value, label }))} aria-label={`修改 ${post.title} 状态`} /> },
            { title: '更新时间', dataIndex: 'updated_at', width: 125, render: (date: string) => <span className="table-date">{dayjs(date).format('YYYY-MM-DD')}<small>{dayjs(date).format('HH:mm')}</small></span> },
            { title: '操作', key: 'actions', fixed: 'right', width: 100, render: (_, post) => <Space><Button type="text" icon={<EditOutlined />} onClick={() => navigate(`/admin/posts/${post.id}/edit`)} aria-label={`编辑 ${post.title}`} /><Button type="text" danger icon={<DeleteOutlined />} onClick={() => confirmDelete(post)} aria-label={`删除 ${post.title}`} /></Space> },
          ]}
        />
      </Card>
    </div>
  );
}
