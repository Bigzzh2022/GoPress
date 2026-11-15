import { lazy, Suspense, type ReactNode } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { PublicLayout } from './layouts/PublicLayout';
import { AdminLayout } from './layouts/AdminLayout';
import { ProtectedRoute } from './app/ProtectedRoute';

const HomePage = lazy(() => import('./pages/public/HomePage'));
const PostsPage = lazy(() => import('./pages/public/PostsPage'));
const PostDetailPage = lazy(() => import('./pages/public/PostDetailPage'));
const AboutPage = lazy(() => import('./pages/public/AboutPage'));
const NotFoundPage = lazy(() => import('./pages/public/NotFoundPage'));
const LoginPage = lazy(() => import('./pages/admin/LoginPage'));
const DashboardPage = lazy(() => import('./pages/admin/DashboardPage'));
const AdminPostsPage = lazy(() => import('./pages/admin/PostsPage'));
const PostEditorPage = lazy(() => import('./pages/admin/PostEditorPage'));
const CategoriesPage = lazy(() => import('./pages/admin/CategoriesPage'));
const TagsPage = lazy(() => import('./pages/admin/TagsPage'));
const ProfilePage = lazy(() => import('./pages/admin/ProfilePage'));

function RouteLoading() {
  return <div className="route-loading" role="status"><span className="route-loader-spinner" aria-hidden="true" /><span>正在载入页面…</span></div>;
}

function load(page: ReactNode) {
  return <Suspense fallback={<RouteLoading />}>{page}</Suspense>;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin/login" element={load(<LoginPage />)} />
        <Route element={<ProtectedRoute />}>
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={load(<DashboardPage />)} />
            <Route path="posts" element={load(<AdminPostsPage />)} />
            <Route path="posts/new" element={load(<PostEditorPage />)} />
            <Route path="posts/:id/edit" element={load(<PostEditorPage />)} />
            <Route path="categories" element={load(<CategoriesPage />)} />
            <Route path="tags" element={load(<TagsPage />)} />
            <Route path="profile" element={load(<ProfilePage />)} />
            <Route path="*" element={load(<NotFoundPage />)} />
          </Route>
        </Route>
        <Route element={<PublicLayout />}>
          <Route index element={load(<HomePage />)} />
          <Route path="posts" element={load(<PostsPage />)} />
          <Route path="posts/:slug" element={load(<PostDetailPage />)} />
          <Route path="about" element={load(<AboutPage />)} />
          <Route path="*" element={load(<NotFoundPage />)} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
