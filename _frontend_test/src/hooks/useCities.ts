import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type {
  ApiResponse,
  PaginatedData,
  City,
  CreateCityRequest,
  UpdateCityRequest,
  QueryParams,
} from '../types'

export function useCities(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['cities', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<City>>>(`/cities${queryString}`)
      return response.data.data
    },
  })
}

export function useCity(id: number) {
  return useQuery({
    queryKey: ['city', id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<City>>(`/cities/${id}`)
      return response.data.data
    },
    enabled: id > 0,
  })
}

export function useCreateCity() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateCityRequest) => {
      const response = await api.post<ApiResponse<City>>('/cities', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cities'] })
      queryClient.invalidateQueries({ queryKey: ['countryCities'] })
      toast.success('CITY CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateCity() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: UpdateCityRequest }) => {
      const response = await api.put<ApiResponse<City>>(`/cities/${id}`, data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cities'] })
      queryClient.invalidateQueries({ queryKey: ['countryCities'] })
      toast.success('CITY UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteCity() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/cities/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cities'] })
      queryClient.invalidateQueries({ queryKey: ['countryCities'] })
      toast.success('CITY DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
