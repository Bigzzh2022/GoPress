import { api } from './client';
import type { ApiResponse, AuthPayload, User } from '../types';

export const authApi = {
  login: async (email: string, password: string) => {
    const response = await api.post<ApiResponse<AuthPayload>>('/auth/login', { email, password });
    return response.data.data;
  },
  logout: () => api.post('/auth/logout'),
  me: async () => {
    const response = await api.get<ApiResponse<User>>('/auth/me');
    return response.data.data;
  },
};
