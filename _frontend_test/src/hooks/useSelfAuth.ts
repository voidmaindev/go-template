import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { useAuthStore } from '../store/auth'
import type {
  ApiResponse,
  SelfRegisterRequest,
  SelfRegisterResponse,
  VerifyEmailRequest,
  ResendVerificationRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  SetPasswordRequest,
  OAuthProvider,
  OAuthTokenRequest,
  TokenResponse,
  IdentitiesResponse,
  User,
} from '../types'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000/api/v1'

export function useSelfRegister() {
  const navigate = useNavigate()
  const setPendingVerificationEmail = useAuthStore((state) => state.setPendingVerificationEmail)

  return useMutation({
    mutationFn: async (data: SelfRegisterRequest) => {
      const response = await api.post<ApiResponse<SelfRegisterResponse>>('/auth/self/register', data)
      return { ...response.data.data, email: data.email }
    },
    onSuccess: (data) => {
      setPendingVerificationEmail(data.email)
      toast.success('ACCOUNT CREATED // CHECK YOUR EMAIL')
      navigate('/verification-pending')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useVerifyEmail() {
  const navigate = useNavigate()

  return useMutation({
    mutationFn: async (data: VerifyEmailRequest) => {
      const response = await api.post<ApiResponse<{ message: string }>>('/auth/self/verify-email', data)
      return response.data
    },
    onSuccess: () => {
      toast.success('EMAIL VERIFIED // PLEASE LOGIN')
      setTimeout(() => navigate('/login'), 2000)
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useResendVerification() {
  return useMutation({
    mutationFn: async (data: ResendVerificationRequest) => {
      const response = await api.post<ApiResponse<{ message: string }>>('/auth/self/resend-verification', data)
      return response.data.data
    },
    onSuccess: () => {
      toast.success('VERIFICATION EMAIL SENT')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useForgotPassword() {
  return useMutation({
    mutationFn: async (data: ForgotPasswordRequest) => {
      const response = await api.post<ApiResponse<{ message: string }>>('/auth/self/forgot-password', data)
      return response.data.data
    },
    onSuccess: () => {
      toast.success('PASSWORD RESET EMAIL SENT')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useResetPassword() {
  const navigate = useNavigate()

  return useMutation({
    mutationFn: async (data: ResetPasswordRequest) => {
      const response = await api.post<ApiResponse<{ message: string }>>('/auth/self/reset-password', data)
      return response.data.data
    },
    onSuccess: () => {
      toast.success('PASSWORD RESET // PLEASE LOGIN')
      navigate('/login')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useOAuthLogin(provider: OAuthProvider) {
  const setAuth = useAuthStore((state) => state.setAuth)

  const openOAuthPopup = () => {
    const width = 500
    const height = 600
    const left = window.screenX + (window.outerWidth - width) / 2
    const top = window.screenY + (window.outerHeight - height) / 2

    const redirectUrl = `${window.location.origin}/auth/oauth/callback`
    const authUrl = `${API_BASE_URL}/auth/oauth/${provider}?redirect_url=${encodeURIComponent(redirectUrl)}`

    const popup = window.open(
      authUrl,
      `oauth-${provider}`,
      `width=${width},height=${height},left=${left},top=${top}`
    )

    return popup
  }

  return { openOAuthPopup, setAuth }
}

export function useOAuthCallback() {
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: OAuthTokenRequest) => {
      const response = await api.post<ApiResponse<TokenResponse>>(
        `/auth/oauth/${data.provider}/token`,
        { code: data.code, state: data.state }
      )
      return response.data.data
    },
    onSuccess: (data) => {
      setAuth(data)
      // Send message to opener window
      if (window.opener) {
        window.opener.postMessage({ type: 'oauth-success', data }, window.location.origin)
        window.close()
      }
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
      if (window.opener) {
        window.opener.postMessage({ type: 'oauth-error', error: getErrorMessage(error) }, window.location.origin)
        window.close()
      }
    },
  })
}

export function useIdentities() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  return useQuery({
    queryKey: ['identities'],
    queryFn: async () => {
      const response = await api.get<ApiResponse<IdentitiesResponse>>('/users/me/identities')
      return response.data.data.identities
    },
    enabled: isAuthenticated,
  })
}

export function useLinkIdentity() {
  const queryClient = useQueryClient()

  const openLinkPopup = (provider: OAuthProvider) => {
    const width = 500
    const height = 600
    const left = window.screenX + (window.outerWidth - width) / 2
    const top = window.screenY + (window.outerHeight - height) / 2

    const redirectUrl = `${window.location.origin}/auth/oauth/callback?link=true`
    const authUrl = `${API_BASE_URL}/users/me/identities/${provider}?redirect_url=${encodeURIComponent(redirectUrl)}`

    const popup = window.open(
      authUrl,
      `link-${provider}`,
      `width=${width},height=${height},left=${left},top=${top}`
    )

    return popup
  }

  const onLinkSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['identities'] })
    toast.success('ACCOUNT LINKED')
  }

  return { openLinkPopup, onLinkSuccess }
}

export function useUnlinkIdentity() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (provider: OAuthProvider) => {
      await api.delete(`/users/me/identities/${provider}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['identities'] })
      toast.success('ACCOUNT UNLINKED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useSetPassword() {
  const queryClient = useQueryClient()
  const updateUser = useAuthStore((state) => state.updateUser)

  return useMutation({
    mutationFn: async (data: SetPasswordRequest) => {
      const response = await api.post<ApiResponse<User>>('/users/me/set-password', data)
      return response.data.data
    },
    onSuccess: (data) => {
      updateUser(data)
      queryClient.invalidateQueries({ queryKey: ['currentUser'] })
      toast.success('PASSWORD SET')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
