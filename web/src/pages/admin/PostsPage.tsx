import { useState, type Key } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Input, Select, Space, Table, Tabs, Tag } from 'antd';
import { CheckCircleOutlined, DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined, UndoOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { AdminPostQuery, BatchAction, PostStatus, PostSummary } from '../../types';

const statusText: Record<PostStatus, string> = { published: '已发布', draft: '草稿' };

export default function AdminPostsPage() {
  useDocumentTitle('文章管理');
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<AdminPostQuery>({ page: 1, page_size: 10, keyword: '', status: '', category: '' });
  const [keyword, setKeyword] = useState('');
  const [selectedKeys, setSelectedKeys] = useState<Key[]>([]);
  const isTrash = filters.status === 'trash';
  const selectedIds = selectedKeys.map(Number);

  const postsQuery = useQuery({ queryKey: ['admin', 'posts', filters], queryFn: () => adminApi.posts(filters), placeholderData: keepPreviousData });
  const categoriesQuery = useQuery({ queryKey: ['admin', 'categories'], queryFn: adminApi.categories });
  const dashboardQuery = useQuery({ queryKey: ['admin', 'dashboard'], queryFn: adminApi.dashboard });
  const dashboard = dashboardQuery.data;

  const invalidatePosts = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['admin', 'posts'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'post'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'related'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'categories'] }),
      queryClient.invalidateQueries({ queryKey: ['public', 'tags'] }),
    ]);
  };
  const deleteMutation = useMutation({
    mutationFn: adminApi.deletePost,
    onSuccess: async () => { message.success('文章已移入回收站'); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const restoreMutation = useMutation({
    mutationFn: adminApi.restorePost,
    onSuccess: async () => { message.success('文章已恢复'); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const purgeMutation = useMutation({
    mutationFn: adminApi.purgePost,
    onSuccess: async () => { message.success('文章已彻底删除'); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: PostStatus }) => adminApi.setPostStatus(id, status),
    onSuccess: async (_, variables) => { message.success(`状态已更新为「${statusText[variables.status]}」`); await invalidatePosts(); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const batchMutation = useMutation({
    mutationFn: ({ action, ids }: { action: BatchAction; ids: number[] }) => adminApi.batchPosts(action, ids),
    onSuccess: async (result) => {
      message.success(result.affected > 0 ? `已处理 ${result.affected} 篇文章` : '操作已完成');
      setSelectedKeys([]);
      await invalidatePosts();
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  const changeFilter = (changes: Partial<AdminPostQuery>) => {
    setSelectedKeys([]);
    setFilters((current) => ({ ...current, page: 1, ...changes }));
  };
  const confirmDelete = (post: PostSummary) => modal.confirm({
    title: '将这篇文章移入回收站？',
    content: `「${post.title}」会保留在回收站中，你可以随时恢复它。`,
    okText: '移入回收站', cancelText: '取消', okButtonProps: { danger: true },
    onOk: () => deleteMutation.mutateAsync(post.id).then(() => undefined),
  });
  const confirmPurge = (post: PostSummary) => modal.confirm({
    title: '彻底删除这篇文章？',
    content: `「${post.title}」及其标签关联将被永久删除，此操作无法撤销。`,
    okText: '彻底删除', cancelText: '取消', okButtonProps: { danger: true },
    onOk: () => purgeMutation.mutateAsync(post.id).then(() => undefined),
  });
  const runBatch = (action: BatchAction) => {
    const ids = selectedIds;
    if (!ids.length || batchMutation.isPending) return;
    if (action === 'delete') {
      modal.confirm({
        title: `将选中的 ${ids.length} 篇文章移入回收站？`,
        content: '移入后仍可在回收站中恢复。',
        okText: '移入回收站', cancelText: '取消', okButtonProps: { danger: true },
        onOk: () => batchMutation.mutateAsync({ action, ids }).then(() => undefined),
      });
      return;
    }
    if (action === 'purge') {
      modal.confirm({
        title: `彻底删除选中的 ${ids.length} 篇文章？`,
        content: '文章及标签关联会被永久删除，此操作无法撤销。',
        okText: '彻底删除', cancelText: '取消', okButtonProps: { danger: true },
        onOk: () => batchMutation.mutateAsync({ action, ids }).then(() => undefined),
      });
      return;
    }
    batchMutation.mutate({ action, ids });
  };

  const tabLabel = (label: string, count?: number) => (
    <span className="list-tab-label">{label}{dashboard && <small>{count ?? 0}</small>}</span>
  );
  const tabItems = [
    { key: '', label: tabLabel('全部', dashboard?.post_count) },
    { key: 'published', label: tabLabel('已发布', dashboard?.published_count) },
    { key: 'draft', label: tabLabel('草稿', dashboard?.draft_count) },
    { key: 'trash', label: tabLabel('回收站', dashboard?.trash_count) },
  ];

  return (
    <div className="admin-page">
      <div className="admin-page-heading"><div><p>CONTENT LIBRARY</p><h1>文章管理</h1><span>创作、编辑并管理你的全部文章。</span></div><Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/admin/posts/new')}>写新文章</Button></div>
      <Card className="admin-table-card">
        <Tabs className="admin-list-tabs" activeKey={filters.status || ''} onChange={(key) => changeFilter({ status: key as AdminPostQuery['status'] })} items={tabItems} />
        <div className="table-filters">
          <Input value={keyword} onChange={(event) => setKeyword(event.target.value)} onPressEnter={() => changeFilter({ keyword })} placeholder="搜索文章标题或摘要" prefix={<SearchOutlined />} allowClear />
          <Select value={filters.category} onChange={(category) => changeFilter({ category })} options={[{ value: '', label: '全部分类' }, ...(categoriesQuery.data?.map((item) => ({ value: item.slug, label: item.name })) || [])]} aria-label="筛选文章分类" />
          <Button type="primary" ghost icon={<SearchOutlined />} onClick={() => changeFilter({ keyword })}>查询</Button>
        </div>
        {selectedIds.length > 0 && (
          <div className="batch-bar">
            <span>已选择 <strong>{selectedIds.length}</strong> 篇</span>
            <Space wrap size={8}>
              {isTrash ? (
                <>
                  <Button size="small" loading={batchMutation.isPending} icon={<UndoOutlined />} onClick={() => runBatch('restore')}>恢复</Button>
                  <Button size="small" danger loading={batchMutation.isPending} icon={<DeleteOutlined />} onClick={() => runBatch('purge')}>彻底删除</Button>
                </>
              ) : (
                <>
                  <Button size="small" loading={batchMutation.isPending} icon={<CheckCircleOutlined />} onClick={() => runBatch('publish')}>发布</Button>
                  <Button size="small" loading={batchMutation.isPending} icon={<EditOutlined />} onClick={() => runBatch('draft')}>转为草稿</Button>
                  <Button size="small" danger loading={batchMutation.isPending} icon={<DeleteOutlined />} onClick={() => runBatch('delete')}>移入回收站</Button>
                </>
              )}
              <Button size="small" type="text" onClick={() => setSelectedKeys([])}>取消选择</Button>
            </Space>
          </div>
        )}
        <Table<PostSummary>
          rowKey="id" dataSource={postsQuery.data?.items} loading={postsQuery.isLoading} scroll={{ x: 980 }}
          rowSelection={{ selectedRowKeys: selectedKeys, onChange: (keys) => setSelectedKeys(keys) }}
          locale={{ emptyText: isTrash ? '回收站是空的' : '还没有文章，开始写一篇吧' }}
          pagination={{ current: postsQuery.data?.page || filters.page, pageSize: filters.page_size, total: postsQuery.data?.total, showSizeChanger: true, showTotal: (total) => `共 ${total} 篇`, onChange: (page, page_size) => { setSelectedKeys([]); setFilters((current) => ({ ...current, page, page_size })); } }}
          columns={[
            {
              title: '文章', dataIndex: 'title', width: 300,
              render: (title: string, post) => isTrash
                ? <div className="admin-post-cell is-static"><strong>{title}</strong><span>/{post.slug}</span></div>
                : <button className="admin-post-cell" onClick={() => navigate(`/admin/posts/${post.id}/edit`)}><strong>{title}</strong><span>/{post.slug}</span></button>,
            },
            { title: '分类', dataIndex: 'category', width: 110, render: (_, post) => post.category?.name || <span className="muted">未分类</span> },
            { title: '标签', dataIndex: 'tags', width: 160, render: (tags: PostSummary['tags']) => <div className="table-tags">{tags?.slice(0, 2).map((tag) => <Tag key={tag.id}>{tag.name}</Tag>)}{tags?.length > 2 && <span>+{tags.length - 2}</span>}</div> },
            {
              title: '状态', dataIndex: 'status', width: 160,
              render: (status: PostStatus, post) => {
                if (isTrash) return <Tag color="default">回收站</Tag>;
                const scheduled = status === 'published' && post.published_at && dayjs(post.published_at).isAfter(dayjs());
                return <div className="status-cell"><Select<PostStatus> size="small" value={status} loading={statusMutation.isPending && statusMutation.variables?.id === post.id} onChange={(next) => statusMutation.mutate({ id: post.id, status: next })} options={Object.entries(statusText).map(([value, label]) => ({ value, label }))} aria-label={`修改 ${post.title} 状态`} />{scheduled && <Tag className="status-scheduled">定时</Tag>}</div>;
              },
            },
            {
              title: isTrash ? '删除时间' : '更新时间', dataIndex: isTrash ? 'deleted_at' : 'updated_at', width: 130,
              render: (date: string | null) => date ? <span className="table-date">{dayjs(date).format('YYYY-MM-DD')}<small>{dayjs(date).format('HH:mm')}</small></span> : <span className="muted">—</span>,
            },
            {
              title: '操作', key: 'actions', fixed: 'right', width: 100,
              render: (_, post) => isTrash
                ? <Space><Button type="text" icon={<UndoOutlined />} onClick={() => restoreMutation.mutate(post.id)} aria-label={`恢复 ${post.title}`} /><Button type="text" danger icon={<DeleteOutlined />} onClick={() => confirmPurge(post)} aria-label={`彻底删除 ${post.title}`} /></Space>
                : <Space><Button type="text" icon={<EditOutlined />} onClick={() => navigate(`/admin/posts/${post.id}/edit`)} aria-label={`编辑 ${post.title}`} /><Button type="text" danger icon={<DeleteOutlined />} onClick={() => confirmDelete(post)} aria-label={`删除 ${post.title}`} /></Space>,
            },
          ]}
        />
      </Card>
    </div>
  );
}
