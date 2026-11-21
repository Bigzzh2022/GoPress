import { api } from './client';
import type {
  ApiResponse,
  Category,
  CommentItem,
  CommentList,
  CommentPayload,
  Paginated,
  PostDetail,
  PostQuery,
  PostSummary,
  PublicUser,
  SiteSettings,
  Tag,
} from '../types';

const unwrap = <T,>(promise: Promise<{ data: ApiResponse<T> }>) =>
  promise.then((response) => response.data.data);

export const blogApi = {
  posts: (params: PostQuery) =>
    unwrap(api.get<ApiResponse<Paginated<PostSummary>>>('/posts', { params })),
  post: (slug: string) =>
    unwrap(api.get<ApiResponse<PostDetail>>(`/posts/${encodeURIComponent(slug)}`)),
  related: (slug: string) =>
    unwrap(api.get<ApiResponse<PostSummary[]>>(`/posts/${encodeURIComponent(slug)}/related`)),
  categories: () => unwrap(api.get<ApiResponse<Category[]>>('/categories')),
  tags: () => unwrap(api.get<ApiResponse<Tag[]>>('/tags')),
  profile: () => unwrap(api.get<ApiResponse<PublicUser>>('/profile')),
  settings: () => unwrap(api.get<ApiResponse<SiteSettings>>('/settings')),
  comments: (slug: string, params?: { page?: number; page_size?: number }) =>
    unwrap(api.get<ApiResponse<CommentList>>(`/posts/${encodeURIComponent(slug)}/comments`, { params })),
  createComment: (slug: string, payload: CommentPayload) =>
    unwrap(api.post<ApiResponse<CommentItem>>(`/posts/${encodeURIComponent(slug)}/comments`, payload)),
};
