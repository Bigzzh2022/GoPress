export type PostStatus = 'draft' | 'published';
export type AdminPostStatus = PostStatus | 'trash';
export type BatchAction = 'publish' | 'draft' | 'delete' | 'restore' | 'purge';

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

// Readers only ever receive the author fields below: the API omits the email.
export interface PublicUser {
  id: number;
  name: string;
  avatar: string;
  bio: string;
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
  author: PublicUser;
  published_at: string | null;
  deleted_at: string | null;
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
  trash_count: number;
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

export interface AdminPostQuery extends Omit<PostQuery, 'status'> {
  status?: AdminPostStatus | '';
}

export interface SiteSettings {
  site_name: string;
  site_tagline: string;
  site_description: string;
  footer_text: string;
  contact_email: string;
}

export interface PasswordPayload {
  current_password: string;
  new_password: string;
}

export interface MediaItem {
  id: number;
  original_name: string;
  url: string;
  mime_type: string;
  size: number;
  created_at: string;
}

export type MediaKind = 'image' | 'video' | 'audio' | 'other';

export interface MediaQuery {
  page?: number;
  page_size?: number;
  keyword?: string;
  kind?: MediaKind | '';
}

export interface CommentItem {
  id: number;
  post_id: number;
  parent_id: number | null;
  nickname: string;
  avatar: string;
  content: string;
  created_at: string;
  replies: CommentItem[];
}

export interface CommentList {
  items: CommentItem[];
  total: number;
  comment_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CommentPayload {
  nickname: string;
  email: string;
  qq?: string;
  content: string;
  parent_id?: number | null;
  website?: string;
}

export interface AdminComment {
  id: number;
  post_id: number;
  post_title: string;
  post_slug: string;
  parent_id: number | null;
  parent_nickname: string;
  nickname: string;
  email: string;
  qq: string;
  content: string;
  created_at: string;
}

export interface AdminCommentQuery {
  page?: number;
  page_size?: number;
  keyword?: string;
  post_id?: number;
}
