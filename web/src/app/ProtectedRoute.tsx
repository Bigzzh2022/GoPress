import { useEffect } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { refreshSession } from '../api/client';
import { useAuthStore } from '../stores/authStore';

export function ProtectedRoute() {
  const { accessToken, authReady, setAuthReady } = useAuthStore();
  const location = useLocation();

  useEffect(() => {
    if (!accessToken && !authReady) {
      refreshSession().catch(() => undefined).finally(() => setAuthReady(true));
    }
  }, [accessToken, authReady, setAuthReady]);

  if (!authReady) return <div className="admin-boot"><Spin size="large" tip="正在恢复登录状态" /></div>;
  if (!accessToken) return <Navigate to="/admin/login" replace state={{ from: location }} />;
  return <Outlet />;
}
