import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type {
  ApiResponse,
  PaginatedData,
  Country,
  CreateCountryRequest,
  UpdateCountryRequest,
  City,
  QueryParams,
} from '../types'

export function useCountries(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['countries', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<Country>>>(`/countries${queryString}`)
      return response.data.data
    },
  })
}

export function useCountry(id: number) {
  return useQuery({
    queryKey: ['country', id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Country>>(`/countries/${id}`)
      return response.data.data
    },
    enabled: id > 0,
  })
}

export function useCountryCities(countryId: number, params: QueryParams = {}) {
  return useQuery({
    queryKey: ['countryCities', countryId, params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<City>>>(
        `/countries/${countryId}/cities${queryString}`
      )
      return response.data.data
    },
    enabled: countryId > 0,
  })
}

export function useCreateCountry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateCountryRequest) => {
      const response = await api.post<ApiResponse<Country>>('/countries', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['countries'] })
      toast.success('COUNTRY CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateCountry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: UpdateCountryRequest }) => {
      const response = await api.put<ApiResponse<Country>>(`/countries/${id}`, data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['countries'] })
      toast.success('COUNTRY UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteCountry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/countries/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['countries'] })
      toast.success('COUNTRY DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
