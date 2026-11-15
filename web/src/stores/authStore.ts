import { create } from 'zustand';
import type { User } from '../types';

interface AuthState {
  accessToken: string | null;
  user: User | null;
  authReady: boolean;
  setAuth: (accessToken: string, user: User) => void;
  setUser: (user: User) => void;
  setAuthReady: (ready: boolean) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  user: null,
  authReady: false,
  setAuth: (accessToken, user) => set({ accessToken, user, authReady: true }),
  setUser: (user) => set({ user }),
  setAuthReady: (authReady) => set({ authReady }),
  clearAuth: () => set({ accessToken: null, user: null, authReady: true }),
}));
