export type PostStatus = 'draft' | 'published';

export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  post_count: number;
  created_at: string;
}

export interface Tag {
  id: number;
  name: string;
  slug: string;
  post_count: number;
  created_at: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
  bio: string;
  role: string;
}

export interface PostSummary {
  id: number;
  title: string;
  slug: string;
  excerpt: string;
  cover_image: string;
  status: PostStatus;
  category_id: number | null;
  category: Category | null;
  tags: Tag[];
  author: User;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface PostDetail extends PostSummary {
  content_md: string;
  content_html: string;
}

export interface PostPayload {
  title: string;
  slug: string;
  excerpt: string;
  content_md: string;
  cover_image: string;
  category_id: number | null;
  tag_ids: number[];
  status: PostStatus;
  published_at: string | null;
}

export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ApiResponse<T> {
  data: T;
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
  };
}

export interface AuthPayload {
  access_token: string;
  user: User;
}

export interface DashboardData {
  post_count: number;
  published_count: number;
  draft_count: number;
  category_count: number;
  tag_count: number;
  recent_posts: PostSummary[];
}

export interface PostQuery {
  page?: number;
  page_size?: number;
  keyword?: string;
  category?: string;
  tag?: string;
  status?: PostStatus | '';
}
