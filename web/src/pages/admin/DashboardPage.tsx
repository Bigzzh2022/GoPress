import { useQuery } from '@tanstack/react-query';
import { Button, Card, Table, Tag, Skeleton, Alert } from 'antd';
import { ArrowRightOutlined, CheckCircleOutlined, EditOutlined, FileTextOutlined, FolderOutlined, PlusOutlined, TagsOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { PostSummary } from '../../types';
import { useAuthStore } from '../../stores/authStore';

export default function DashboardPage() {
  useDocumentTitle('工作台');
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const dashboardQuery = useQuery({ queryKey: ['admin', 'dashboard'], queryFn: adminApi.dashboard });
  const data = dashboardQuery.data;
  const stats = [
    { label: '全部文章', value: data?.post_count, icon: <FileTextOutlined />, tone: 'orange' },
    { label: '已发布', value: data?.published_count, icon: <CheckCircleOutlined />, tone: 'green' },
    { label: '草稿箱', value: data?.draft_count, icon: <EditOutlined />, tone: 'blue' },
    { label: '内容分类', value: data?.category_count, icon: <FolderOutlined />, tone: 'purple' },
    { label: '内容标签', value: data?.tag_count, icon: <TagsOutlined />, tone: 'cyan' },
  ];
  return (
    <div className="admin-page dashboard-page">
      <section className="admin-welcome"><div><p>CONTENT OVERVIEW</p><h1>{new Date().getHours() < 12 ? '早上好' : new Date().getHours() < 18 ? '下午好' : '晚上好'}，{user?.name || '写作者'}</h1><span>今天也写一点值得留下的内容。</span></div><Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/admin/posts/new')}>写新文章</Button></section>
      {dashboardQuery.isError && <Alert className="dashboard-error" type="error" showIcon message="工作台数据加载失败" description={getErrorMessage(dashboardQuery.error)} action={<Button size="small" onClick={() => dashboardQuery.refetch()}>重试</Button>} />}
      <div className="stat-grid">{stats.map((stat) => <Card key={stat.label} className="stat-card"><div className={`stat-icon ${stat.tone}`}>{stat.icon}</div><div><span>{stat.label}</span>{dashboardQuery.isLoading ? <Skeleton.Input active size="small" /> : <strong>{stat.value ?? 0}</strong>}<small>实时内容统计</small></div></Card>)}</div>
      <section className="dashboard-grid">
        <Card className="recent-card" title={<div className="admin-card-title"><span>最近文章</span><small>最近更新的内容</small></div>} extra={<Link to="/admin/posts">全部文章 <ArrowRightOutlined /></Link>}>
          <Table<PostSummary> rowKey="id" dataSource={data?.recent_posts} loading={dashboardQuery.isLoading} pagination={false} scroll={{ x: 600 }} columns={[
            { title: '文章', dataIndex: 'title', render: (title: string, record) => <Link className="table-post-title" to={`/admin/posts/${record.id}/edit`}>{title}<small>{record.category?.name || '未分类'}</small></Link> },
            { title: '状态', dataIndex: 'status', width: 95, render: (status: PostSummary['status']) => <Tag color={status === 'published' ? 'green' : 'gold'}>{status === 'published' ? '已发布' : '草稿'}</Tag> },
            { title: '更新时间', dataIndex: 'updated_at', width: 120, render: (date: string) => dayjs(date).format('MM-DD HH:mm') },
            { title: '', key: 'action', width: 45, render: (_, record) => <Button type="text" icon={<ArrowRightOutlined />} onClick={() => navigate(`/admin/posts/${record.id}/edit`)} aria-label={`编辑 ${record.title}`} /> },
          ]} />
        </Card>
        <Card className="quick-card" title="快速开始"><p>让创作流程保持简单。</p><Button block type="primary" icon={<PlusOutlined />} onClick={() => navigate('/admin/posts/new')}>创建一篇文章</Button><Button block icon={<FolderOutlined />} onClick={() => navigate('/admin/categories')}>整理内容分类</Button><Button block icon={<EditOutlined />} onClick={() => navigate('/admin/profile')}>完善个人资料</Button><div className="writing-tip"><span>今日写作提示</span><p>先完成，再完善。草稿不需要一开始就完美。</p></div></Card>
      </section>
    </div>
  );
}
