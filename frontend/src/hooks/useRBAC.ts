import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type {
  ApiResponse,
  PaginatedData,
  Role,
  CreateRoleRequest,
  UpdateRolePermissionsRequest,
  AssignRoleRequest,
  UserRolesResponse,
  DomainsResponse,
  ActionsResponse,
  QueryParams,
} from '../types'

export function useRoles(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['roles', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<Role>>>(`/rbac/roles${queryString}`)
      return response.data.data
    },
  })
}

export function useRole(code: string) {
  return useQuery({
    queryKey: ['role', code],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Role>>(`/rbac/roles/${code}`)
      return response.data.data
    },
    enabled: !!code,
  })
}

export function useCreateRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateRoleRequest) => {
      const response = await api.post<ApiResponse<Role>>('/rbac/roles', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      toast.success('ROLE CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateRolePermissions() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ code, data }: { code: string; data: UpdateRolePermissionsRequest }) => {
      const response = await api.put<ApiResponse<Role>>(`/rbac/roles/${code}/permissions`, data)
      return response.data.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      queryClient.invalidateQueries({ queryKey: ['role', variables.code] })
      toast.success('PERMISSIONS UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (code: string) => {
      await api.delete(`/rbac/roles/${code}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      toast.success('ROLE DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUserRoles(userId: number) {
  return useQuery({
    queryKey: ['userRoles', userId],
    queryFn: async () => {
      const response = await api.get<ApiResponse<UserRolesResponse>>(`/rbac/users/${userId}/roles`)
      return response.data.data
    },
    enabled: userId > 0,
  })
}

export function useAssignRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ userId, data }: { userId: number; data: AssignRoleRequest }) => {
      await api.post(`/rbac/users/${userId}/roles`, data)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['userRoles', variables.userId] })
      toast.success('ROLE ASSIGNED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useRemoveRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ userId, roleCode }: { userId: number; roleCode: string }) => {
      await api.delete(`/rbac/users/${userId}/roles/${roleCode}`)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['userRoles', variables.userId] })
      toast.success('ROLE REMOVED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDomains() {
  return useQuery({
    queryKey: ['domains'],
    queryFn: async () => {
      const response = await api.get<ApiResponse<DomainsResponse>>('/rbac/domains')
      return response.data.data
    },
  })
}

export function useActions() {
  return useQuery({
    queryKey: ['actions'],
    queryFn: async () => {
      const response = await api.get<ApiResponse<ActionsResponse>>('/rbac/actions')
      return response.data.data
    },
  })
}
