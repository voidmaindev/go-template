import { useState } from 'react'
import { motion } from 'framer-motion'
import {
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Search,
  X,
  Loader2,
} from 'lucide-react'
import { cn } from '../lib/utils'

export interface Column<T> {
  key: string
  header: string
  sortable?: boolean
  render?: (item: T) => React.ReactNode
  className?: string
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  isLoading?: boolean
  // Pagination
  page?: number
  pageSize?: number
  totalPages?: number
  total?: number
  onPageChange?: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  // Sorting
  sortField?: string
  sortOrder?: 'asc' | 'desc'
  onSort?: (field: string, order: 'asc' | 'desc') => void
  // Search
  searchValue?: string
  onSearchChange?: (value: string) => void
  searchPlaceholder?: string
  // Actions
  onRowClick?: (item: T) => void
  actions?: (item: T) => React.ReactNode
}

export default function DataTable<T extends { id: number | string }>({
  columns,
  data,
  isLoading,
  page = 1,
  pageSize = 10,
  totalPages = 1,
  total = 0,
  onPageChange,
  onPageSizeChange,
  sortField,
  sortOrder,
  onSort,
  searchValue,
  onSearchChange,
  searchPlaceholder = 'Search...',
  onRowClick,
  actions,
}: DataTableProps<T>) {
  const [localSearch, setLocalSearch] = useState(searchValue || '')

  const handleSort = (field: string) => {
    if (!onSort) return
    const newOrder = sortField === field && sortOrder === 'asc' ? 'desc' : 'asc'
    onSort(field, newOrder)
  }

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSearchChange?.(localSearch)
  }

  const clearSearch = () => {
    setLocalSearch('')
    onSearchChange?.('')
  }

  return (
    <div className="space-y-4">
      {/* Search Bar */}
      {onSearchChange && (
        <form onSubmit={handleSearchSubmit} className="flex gap-2">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
            <input
              type="text"
              value={localSearch}
              onChange={(e) => setLocalSearch(e.target.value)}
              placeholder={searchPlaceholder}
              className="cyber-input pl-10 pr-10"
            />
            {localSearch && (
              <button
                type="button"
                onClick={clearSearch}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
              >
                <X className="w-5 h-5" />
              </button>
            )}
          </div>
          <button type="submit" className="cyber-button">
            Search
          </button>
        </form>
      )}

      {/* Table */}
      <div className="cyber-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="cyber-table">
            <thead>
              <tr>
                {columns.map((col) => (
                  <th
                    key={col.key}
                    className={cn(col.className, col.sortable && 'cursor-pointer select-none')}
                    onClick={() => col.sortable && handleSort(col.key)}
                  >
                    <div className="flex items-center gap-2">
                      {col.header}
                      {col.sortable && (
                        <span className="flex flex-col">
                          <ChevronUp
                            className={cn(
                              'w-3 h-3 -mb-1',
                              sortField === col.key && sortOrder === 'asc'
                                ? 'text-neon-cyan'
                                : 'text-gray-600'
                            )}
                          />
                          <ChevronDown
                            className={cn(
                              'w-3 h-3',
                              sortField === col.key && sortOrder === 'desc'
                                ? 'text-neon-cyan'
                                : 'text-gray-600'
                            )}
                          />
                        </span>
                      )}
                    </div>
                  </th>
                ))}
                {actions && <th className="w-20">Actions</th>}
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={columns.length + (actions ? 1 : 0)} className="text-center py-12">
                    <Loader2 className="w-8 h-8 text-neon-cyan animate-spin mx-auto" />
                    <p className="mt-2 text-gray-500 font-mono text-sm">LOADING DATA...</p>
                  </td>
                </tr>
              ) : data.length === 0 ? (
                <tr>
                  <td colSpan={columns.length + (actions ? 1 : 0)} className="text-center py-12">
                    <p className="text-gray-500 font-mono text-sm">NO RECORDS FOUND</p>
                  </td>
                </tr>
              ) : (
                data.map((item, index) => (
                  <motion.tr
                    key={item.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.03 }}
                    onClick={() => onRowClick?.(item)}
                    className={cn(onRowClick && 'cursor-pointer')}
                  >
                    {columns.map((col) => (
                      <td key={col.key} className={col.className}>
                        {col.render
                          ? col.render(item)
                          : (item as Record<string, unknown>)[col.key]?.toString() ?? '-'}
                      </td>
                    ))}
                    {actions && (
                      <td onClick={(e) => e.stopPropagation()}>{actions(item)}</td>
                    )}
                  </motion.tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 0 && (
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 px-4 py-3 border-t border-cyber-border">
            <div className="flex items-center gap-4 text-sm text-gray-400">
              <span className="font-mono">
                {total} records // Page {page} of {totalPages}
              </span>
              {onPageSizeChange && (
                <select
                  value={pageSize}
                  onChange={(e) => onPageSizeChange(Number(e.target.value))}
                  className="bg-cyber-darker border border-cyber-border rounded px-2 py-1 text-sm focus:outline-none focus:border-neon-cyan"
                >
                  {[10, 25, 50, 100].map((size) => (
                    <option key={size} value={size}>
                      {size} / page
                    </option>
                  ))}
                </select>
              )}
            </div>
            <div className="flex items-center gap-1">
              <button
                onClick={() => onPageChange?.(1)}
                disabled={page === 1}
                className="p-2 hover:bg-cyber-light rounded disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronsLeft className="w-4 h-4" />
              </button>
              <button
                onClick={() => onPageChange?.(page - 1)}
                disabled={page === 1}
                className="p-2 hover:bg-cyber-light rounded disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <div className="flex items-center gap-1 px-2">
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  let pageNum: number
                  if (totalPages <= 5) {
                    pageNum = i + 1
                  } else if (page <= 3) {
                    pageNum = i + 1
                  } else if (page >= totalPages - 2) {
                    pageNum = totalPages - 4 + i
                  } else {
                    pageNum = page - 2 + i
                  }
                  return (
                    <button
                      key={pageNum}
                      onClick={() => onPageChange?.(pageNum)}
                      className={cn(
                        'w-8 h-8 rounded font-mono text-sm',
                        page === pageNum
                          ? 'bg-neon-cyan/20 text-neon-cyan border border-neon-cyan/50'
                          : 'hover:bg-cyber-light'
                      )}
                    >
                      {pageNum}
                    </button>
                  )
                })}
              </div>
              <button
                onClick={() => onPageChange?.(page + 1)}
                disabled={page === totalPages}
                className="p-2 hover:bg-cyber-light rounded disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
              <button
                onClick={() => onPageChange?.(totalPages)}
                disabled={page === totalPages}
                className="p-2 hover:bg-cyber-light rounded disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronsRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
