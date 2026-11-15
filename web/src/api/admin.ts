import { api } from './client';
import type {
  ApiResponse,
  Category,
  DashboardData,
  Paginated,
  PostDetail,
  PostPayload,
  PostQuery,
  PostStatus,
  PostSummary,
  Tag,
  User,
} from '../types';

const unwrap = <T,>(promise: Promise<{ data: ApiResponse<T> }>) =>
  promise.then((response) => response.data.data);

export const adminApi = {
  dashboard: () => unwrap(api.get<ApiResponse<DashboardData>>('/admin/dashboard')),
  posts: (params: PostQuery) =>
    unwrap(api.get<ApiResponse<Paginated<PostSummary>>>('/admin/posts', { params })),
  post: (id: number) => unwrap(api.get<ApiResponse<PostDetail>>(`/admin/posts/${id}`)),
  createPost: (payload: PostPayload) =>
    unwrap(api.post<ApiResponse<PostDetail>>('/admin/posts', payload)),
  updatePost: (id: number, payload: PostPayload) =>
    unwrap(api.put<ApiResponse<PostDetail>>(`/admin/posts/${id}`, payload)),
  deletePost: (id: number) => api.delete(`/admin/posts/${id}`),
  setPostStatus: (id: number, status: PostStatus) =>
    unwrap(api.patch<ApiResponse<PostDetail>>(`/admin/posts/${id}/status`, { status })),
  preview: (content: string) =>
    unwrap(api.post<ApiResponse<{ html: string }>>('/admin/markdown/preview', { content })),
  categories: () => unwrap(api.get<ApiResponse<Category[]>>('/admin/categories')),
  createCategory: (payload: Pick<Category, 'name' | 'slug' | 'description'>) =>
    unwrap(api.post<ApiResponse<Category>>('/admin/categories', payload)),
  updateCategory: (id: number, payload: Pick<Category, 'name' | 'slug' | 'description'>) =>
    unwrap(api.put<ApiResponse<Category>>(`/admin/categories/${id}`, payload)),
  deleteCategory: (id: number) => api.delete(`/admin/categories/${id}`),
  tags: () => unwrap(api.get<ApiResponse<Tag[]>>('/admin/tags')),
  createTag: (payload: Pick<Tag, 'name' | 'slug'>) =>
    unwrap(api.post<ApiResponse<Tag>>('/admin/tags', payload)),
  updateTag: (id: number, payload: Pick<Tag, 'name' | 'slug'>) =>
    unwrap(api.put<ApiResponse<Tag>>(`/admin/tags/${id}`, payload)),
  deleteTag: (id: number) => api.delete(`/admin/tags/${id}`),
  profile: () => unwrap(api.get<ApiResponse<User>>('/admin/profile')),
  updateProfile: (payload: Pick<User, 'name' | 'email' | 'avatar' | 'bio'>) =>
    unwrap(api.put<ApiResponse<User>>('/admin/profile', payload)),
};
