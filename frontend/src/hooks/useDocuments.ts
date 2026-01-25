import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import api, { getErrorMessage } from '../lib/api'
import { buildQueryString } from '../lib/utils'
import type {
  ApiResponse,
  PaginatedData,
  Document,
  CreateDocumentRequest,
  UpdateDocumentRequest,
  CreateDocumentItemRequest,
  UpdateDocumentItemRequest,
  DocumentItem,
  QueryParams,
} from '../types'

export function useDocuments(params: QueryParams = {}) {
  return useQuery({
    queryKey: ['documents', params],
    queryFn: async () => {
      const queryString = buildQueryString(params)
      const response = await api.get<ApiResponse<PaginatedData<Document>>>(`/documents${queryString}`)
      return response.data.data
    },
  })
}

export function useDocument(id: number) {
  return useQuery({
    queryKey: ['document', id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Document>>(`/documents/${id}`)
      return response.data.data
    },
    enabled: id > 0,
  })
}

export function useCreateDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateDocumentRequest) => {
      const response = await api.post<ApiResponse<Document>>('/documents', data)
      return response.data.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      toast.success('DOCUMENT CREATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: UpdateDocumentRequest }) => {
      const response = await api.put<ApiResponse<Document>>(`/documents/${id}`, data)
      return response.data.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      queryClient.invalidateQueries({ queryKey: ['document', variables.id] })
      toast.success('DOCUMENT UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/documents/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      toast.success('DOCUMENT DELETED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useAddDocumentItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      documentId,
      data,
    }: {
      documentId: number
      data: CreateDocumentItemRequest
    }) => {
      const response = await api.post<ApiResponse<DocumentItem>>(
        `/documents/${documentId}/items`,
        data
      )
      return response.data.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['document', variables.documentId] })
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      toast.success('ITEM ADDED TO DOCUMENT')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useUpdateDocumentItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      documentId,
      itemId,
      data,
    }: {
      documentId: number
      itemId: number
      data: UpdateDocumentItemRequest
    }) => {
      const response = await api.put<ApiResponse<DocumentItem>>(
        `/documents/${documentId}/items/${itemId}`,
        data
      )
      return response.data.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['document', variables.documentId] })
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      toast.success('DOCUMENT ITEM UPDATED')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}

export function useDeleteDocumentItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ documentId, itemId }: { documentId: number; itemId: number }) => {
      await api.delete(`/documents/${documentId}/items/${itemId}`)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['document', variables.documentId] })
      queryClient.invalidateQueries({ queryKey: ['documents'] })
      toast.success('ITEM REMOVED FROM DOCUMENT')
    },
    onError: (error) => {
      toast.error(getErrorMessage(error))
    },
  })
}
