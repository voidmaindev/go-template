import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { useAuthStore } from '../store/auth'
import type {
  ApiResponse,
  LoginRequest,
  RegisterRequest,
  TokenResponse,
  User,
  UpdateProfileRequest,
  ChangePasswordRequest,
} from '../types'

export function useLogin() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: LoginRequest) => {
      const response = await api.post<ApiResponse<TokenResponse>>('/auth/login', data)
      return response.data.data
    },
    onSuccess: (data) => {
      setAuth(data)
      toast.success('ACCESS GRANTED')
      navigate('/dashboard')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useRegister() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<ApiResponse<TokenResponse>>('/auth/register', data)
      return response.data.data
    },
    onSuccess: (data) => {
      setAuth(data)
      toast.success('USER CREATED // ACCESS GRANTED')
      navigate('/dashboard')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useLogout() {
  const navigate = useNavigate()
  const logout = useAuthStore((state) => state.logout)
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      await api.post('/auth/logout')
    },
    onSuccess: () => {
      logout()
      queryClient.clear()
      toast.success('SESSION TERMINATED')
      navigate('/login')
    },
    onError: () => {
      // Even on error, clear local state
      logout()
      queryClient.clear()
      navigate('/login')
    },
  })
}

export function useCurrentUser() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  return useQuery({
    queryKey: ['currentUser'],
    queryFn: async () => {
      const response = await api.get<ApiResponse<User>>('/users/me')
      return response.data.data
    },
    enabled: isAuthenticated,
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  const updateUser = useAuthStore((state) => state.updateUser)

  return useMutation({
    mutationFn: async (data: UpdateProfileRequest) => {
      const response = await api.put<ApiResponse<User>>('/users/me', data)
      return response.data.data
    },
    onSuccess: (data) => {
      updateUser(data)
      queryClient.invalidateQueries({ queryKey: ['currentUser'] })
      toast.success('PROFILE UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: ChangePasswordRequest) => {
      await api.put('/users/me/password', data)
    },
    onSuccess: () => {
      toast.success('PASSWORD CHANGED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
