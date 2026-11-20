import { useMemo, useState } from 'react';
import { Avatar, Dropdown, Drawer, Layout, Menu, Button } from 'antd';
import {
  DashboardOutlined, FileTextOutlined, FolderOutlined, TagsOutlined,
  UserOutlined, MenuFoldOutlined, MenuUnfoldOutlined, LogoutOutlined,
  GlobalOutlined, MenuOutlined, SettingOutlined,
} from '@ant-design/icons';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { authApi } from '../api/auth';
import { queryClient } from '../app/queryClient';
import { useAuthStore } from '../stores/authStore';
import { useUIStore } from '../stores/uiStore';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/admin', icon: <DashboardOutlined />, label: '工作台' },
  { key: '/admin/posts', icon: <FileTextOutlined />, label: '文章管理' },
  { key: '/admin/categories', icon: <FolderOutlined />, label: '分类管理' },
  { key: '/admin/tags', icon: <TagsOutlined />, label: '标签管理' },
  { key: '/admin/settings', icon: <SettingOutlined />, label: '站点设置' },
  { key: '/admin/profile', icon: <UserOutlined />, label: '个人资料' },
];

export function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);
  const { sidebarCollapsed, toggleSidebar } = useUIStore();
  const { user, clearAuth } = useAuthStore();
  const selectedKey = useMemo(() => location.pathname.startsWith('/admin/posts') ? '/admin/posts' : location.pathname, [location.pathname]);
  const logout = useMutation({
    mutationFn: authApi.logout,
    onSettled: () => {
      clearAuth();
      queryClient.clear();
      navigate('/admin/login', { replace: true });
    },
  });

  const menu = <Menu mode="inline" items={menuItems} selectedKeys={[selectedKey]} onClick={({ key }) => { navigate(key); setMobileOpen(false); }} />;
  const sidebar = <><Link to="/admin" className="admin-brand"><span>G</span>{!sidebarCollapsed && <div>GoPress<small>CONTENT STUDIO</small></div>}</Link>{menu}<div className="admin-sidebar-note">{!sidebarCollapsed && <><span>写作提示</span><p>清晰的表达，始于清晰的思考。</p></>}</div></>;

  return (
    <Layout className="admin-shell">
      <Sider className="admin-sider" width={236} collapsedWidth={76} collapsed={sidebarCollapsed} trigger={null} breakpoint="lg">{sidebar}</Sider>
      <Drawer className="admin-mobile-drawer" placement="left" open={mobileOpen} onClose={() => setMobileOpen(false)} width={250} closable={false} styles={{ body: { padding: 0 } }}>{sidebar}</Drawer>
      <Layout>
        <Header className="admin-header">
          <div className="admin-header-left">
            <Button type="text" className="desktop-collapse" icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={toggleSidebar} aria-label="折叠侧边栏" />
            <Button type="text" className="mobile-collapse" icon={<MenuOutlined />} onClick={() => setMobileOpen(true)} aria-label="打开后台导航" />
            <div><strong>{menuItems.find((item) => item.key === selectedKey)?.label || '文章编辑器'}</strong><span>管理你的内容与个人设置</span></div>
          </div>
          <div className="admin-header-actions">
            <Button icon={<GlobalOutlined />} onClick={() => window.open('/', '_blank', 'noopener,noreferrer')}>查看站点</Button>
            <Dropdown menu={{ items: [{ key: 'profile', icon: <UserOutlined />, label: '个人资料', onClick: () => navigate('/admin/profile') }, { type: 'divider' }, { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true, onClick: () => logout.mutate() }] }} placement="bottomRight">
              <button className="admin-user" aria-label="打开用户菜单"><Avatar size={34}>{user?.name?.slice(0, 1) || 'A'}</Avatar><span><strong>{user?.name || '管理员'}</strong><small>{user?.role || 'admin'}</small></span></button>
            </Dropdown>
          </div>
        </Header>
        <Content className="admin-content"><Outlet /></Content>
      </Layout>
    </Layout>
  );
}
