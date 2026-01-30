import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User, TokenResponse } from '../types'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null
  user: User | null
  isAuthenticated: boolean
  pendingVerificationEmail: string | null
  setAuth: (data: TokenResponse) => void
  updateUser: (user: User) => void
  setPendingVerificationEmail: (email: string | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      user: null,
      isAuthenticated: false,
      pendingVerificationEmail: null,
      setAuth: (data: TokenResponse) =>
        set({
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          expiresAt: data.expires_at,
          user: data.user,
          isAuthenticated: true,
          pendingVerificationEmail: null,
        }),
      updateUser: (user: User) =>
        set((state) => ({
          ...state,
          user,
        })),
      setPendingVerificationEmail: (email: string | null) =>
        set({ pendingVerificationEmail: email }),
      logout: () =>
        set({
          accessToken: null,
          refreshToken: null,
          expiresAt: null,
          user: null,
          isAuthenticated: false,
          pendingVerificationEmail: null,
        }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        expiresAt: state.expiresAt,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        pendingVerificationEmail: state.pendingVerificationEmail,
      }),
    }
  )
)
