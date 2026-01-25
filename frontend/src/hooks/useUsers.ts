import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type { ApiResponse, PaginatedData, User, QueryParams, RegisterRequest } from '../types'

export function useUsers(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<User>>>(`/users${queryString}`)
      return response.data.data
    },
  })
}

export function useUser(id: number) {
  return useQuery({
    queryKey: ['user', id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<User>>(`/users/${id}`)
      return response.data.data
    },
    enabled: id > 0,
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/users/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success('USER DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<ApiResponse<User>>('/auth/register', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success('USER CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
