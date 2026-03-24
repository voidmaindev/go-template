import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type {
  ApiResponse,
  PaginatedData,
  Item,
  CreateItemRequest,
  UpdateItemRequest,
  QueryParams,
} from '../types'

export function useItems(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['items', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<Item>>>(`/items${queryString}`)
      return response.data.data
    },
  })
}

export function useItem(id: number) {
  return useQuery({
    queryKey: ['item', id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Item>>(`/items/${id}`)
      return response.data.data
    },
    enabled: id > 0,
  })
}

export function useCreateItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateItemRequest) => {
      const response = await api.post<ApiResponse<Item>>('/items', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['items'] })
      toast.success('ITEM CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: UpdateItemRequest }) => {
      const response = await api.put<ApiResponse<Item>>(`/items/${id}`, data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['items'] })
      toast.success('ITEM UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/items/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['items'] })
      toast.success('ITEM DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
