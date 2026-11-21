import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Input, Space, Table, Tag } from 'antd';
import { DeleteOutlined, SearchOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { AdminComment, AdminCommentQuery } from '../../types';

export default function CommentsPage() {
  useDocumentTitle('评论管理');
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<AdminCommentQuery>({ page: 1, page_size: 20, keyword: '' });
  const [keyword, setKeyword] = useState('');

  const commentsQuery = useQuery({
    queryKey: ['admin', 'comments', filters],
    queryFn: () => adminApi.comments(filters),
    placeholderData: keepPreviousData,
  });

  const deleteMutation = useMutation({
    mutationFn: adminApi.deleteComment,
    onSuccess: async () => {
      message.success('评论已删除');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'comments'] }),
      ]);
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  const changeFilter = (changes: Partial<AdminCommentQuery>) => setFilters((current) => ({ ...current, page: 1, ...changes }));
  const confirmDelete = (comment: AdminComment) => modal.confirm({
    title: '删除这条评论？',
    content: `「${comment.nickname}」的这条评论及其全部回复都会被永久删除，此操作无法撤销。`,
    okText: '删除', cancelText: '取消', okButtonProps: { danger: true },
    onOk: () => deleteMutation.mutateAsync(comment.id).then(() => undefined),
  });

  return (
    <div className="admin-page">
      <div className="admin-page-heading">
        <div><p>CONVERSATIONS</p><h1>评论管理</h1><span>访客评论即时显示，你可以在这里删除不合适的内容。</span></div>
        <span className="admin-count-chip">共 {commentsQuery.data?.total ?? 0} 条</span>
      </div>
      <Card className="admin-table-card">
        <div className="table-filters is-compact">
          <Input value={keyword} onChange={(event) => setKeyword(event.target.value)} onPressEnter={() => changeFilter({ keyword })} placeholder="搜索昵称、内容、邮箱或 QQ" prefix={<SearchOutlined />} allowClear />
          <Button type="primary" ghost icon={<SearchOutlined />} onClick={() => changeFilter({ keyword })}>查询</Button>
        </div>
        <Table<AdminComment>
          rowKey="id" dataSource={commentsQuery.data?.items} loading={commentsQuery.isLoading} scroll={{ x: 940 }}
          locale={{ emptyText: '还没有评论' }}
          pagination={{
            current: commentsQuery.data?.page || filters.page,
            pageSize: filters.page_size,
            total: commentsQuery.data?.total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, page_size) => setFilters((current) => ({ ...current, page, page_size })),
          }}
          columns={[
            {
              title: '评论', dataIndex: 'content', width: 420,
              render: (content: string, comment) => (
                <div className="comment-cell">
                  <div className="comment-cell-head">
                    <strong>{comment.nickname}</strong>
                    {comment.parent_nickname && <Tag color="default">回复 @{comment.parent_nickname}</Tag>}
                  </div>
                  <p>{content}</p>
                </div>
              ),
            },
            {
              title: '文章', dataIndex: 'post_title', width: 220,
              render: (title: string, comment) => (
                <Link className="comment-post-link" to={`/posts/${comment.post_slug}`} target="_blank" rel="noreferrer">{title}</Link>
              ),
            },
            {
              title: '联系方式', key: 'contact', width: 190,
              render: (_, comment) => (
                <div className="comment-contact">
                  <span>{comment.email}</span>
                  {comment.qq && <small>QQ {comment.qq}</small>}
                </div>
              ),
            },
            {
              title: '时间', dataIndex: 'created_at', width: 130,
              render: (date: string) => <span className="table-date">{dayjs(date).format('YYYY-MM-DD')}<small>{dayjs(date).format('HH:mm')}</small></span>,
            },
            {
              title: '操作', key: 'actions', fixed: 'right', width: 80,
              render: (_, comment) => (
                <Space>
                  <Button type="text" danger icon={<DeleteOutlined />} aria-label={`删除 ${comment.nickname} 的评论`} onClick={() => confirmDelete(comment)} />
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
}
