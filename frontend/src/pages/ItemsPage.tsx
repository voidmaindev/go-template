import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Package, Plus, Edit2, Trash2 } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField, { TextAreaField } from '../components/FormField'
import { useItems, useCreateItem, useUpdateItem, useDeleteItem } from '../hooks/useItems'
import { formatPrice, priceToCents, formatDateTime } from '../lib/utils'
import type { Item, CreateItemRequest, QueryParams } from '../types'

export default function ItemsPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<Item | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useItems(params)
  const { mutate: createItem, isPending: isCreating } = useCreateItem()
  const { mutate: updateItem, isPending: isUpdating } = useUpdateItem()
  const { mutate: deleteItem, isPending: isDeleting } = useDeleteItem()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateItemRequest & { priceDisplay: string }>()

  const columns: Column<Item>[] = [
    {
      key: 'id',
      header: 'ID',
      sortable: true,
      render: (item) => <span className="font-mono text-neon-cyan">#{item.id}</span>,
    },
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (item) => <span className="font-medium text-white">{item.name}</span>,
    },
    {
      key: 'description',
      header: 'Description',
      render: (item) => (
        <span className="text-gray-400 truncate max-w-xs block">
          {item.description || '-'}
        </span>
      ),
    },
    {
      key: 'price',
      header: 'Price',
      sortable: true,
      render: (item) => (
        <span className="font-mono text-neon-green">{formatPrice(item.price)}</span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (item) => <span className="font-mono text-sm">{formatDateTime(item.created_at)}</span>,
    },
  ]

  const openCreateModal = () => {
    setEditingItem(null)
    reset({ name: '', description: '', priceDisplay: '' })
    setModalOpen(true)
  }

  const openEditModal = (item: Item) => {
    setEditingItem(item)
    reset({
      name: item.name,
      description: item.description,
      priceDisplay: (item.price / 100).toFixed(2),
    })
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingItem(null)
    reset()
  }

  const onSubmit = (data: CreateItemRequest & { priceDisplay: string }) => {
    const payload = {
      name: data.name,
      description: data.description,
      price: priceToCents(data.priceDisplay),
    }

    if (editingItem) {
      updateItem(
        { id: editingItem.id, data: payload },
        { onSuccess: closeModal }
      )
    } else {
      createItem(payload, { onSuccess: closeModal })
    }
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteItem(deleteId, {
        onSuccess: () => setDeleteId(null),
      })
    }
  }

  return (
    <div>
      <PageHeader
        title="Items"
        subtitle={`${data?.total ?? 0} items in catalog`}
        icon={Package}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New Item
          </button>
        }
      />

      <DataTable
        columns={columns}
        data={data?.data || []}
        isLoading={isLoading}
        page={params.page}
        pageSize={params.page_size}
        totalPages={data?.total_pages || 1}
        total={data?.total || 0}
        onPageChange={(page) => setParams({ ...params, page })}
        onPageSizeChange={(page_size) => setParams({ ...params, page_size, page: 1 })}
        sortField={params.sort}
        sortOrder={params.order}
        onSort={(field, order) => setParams({ ...params, sort: field, order })}
        searchValue={params['name__contains'] as string}
        onSearchChange={(value) =>
          setParams({ ...params, 'name__contains': value || undefined, page: 1 })
        }
        searchPlaceholder="Search items..."
        actions={(item) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => openEditModal(item)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="Edit"
            >
              <Edit2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setDeleteId(item.id)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-pink transition-colors"
              title="Delete"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        )}
      />

      {/* Create/Edit Modal */}
      <Modal
        isOpen={modalOpen}
        onClose={closeModal}
        title={editingItem ? 'Edit Item' : 'Create Item'}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            label="Name"
            {...register('name', {
              required: 'Name is required',
              minLength: { value: 1, message: 'Name is required' },
              maxLength: { value: 200, message: 'Name is too long' },
            })}
            error={errors.name?.message}
            placeholder="Enter item name"
            required
          />

          <TextAreaField
            label="Description"
            {...register('description', {
              maxLength: { value: 5000, message: 'Description is too long' },
            })}
            error={errors.description?.message}
            placeholder="Enter item description (optional)"
          />

          <FormField
            label="Price (USD)"
            type="number"
            step="0.01"
            min="0"
            {...register('priceDisplay', {
              required: 'Price is required',
              min: { value: 0, message: 'Price must be positive' },
            })}
            error={errors.priceDisplay?.message}
            placeholder="0.00"
            hint="Enter price in dollars (e.g., 19.99)"
            required
          />

          <div className="flex justify-end gap-3 pt-4">
            <button type="button" onClick={closeModal} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isCreating || isUpdating}
              className="cyber-button"
            >
              {isCreating || isUpdating
                ? 'Saving...'
                : editingItem
                ? 'Update Item'
                : 'Create Item'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        onConfirm={handleDelete}
        title="Delete Item"
        message="Are you sure you want to delete this item? This action cannot be undone."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />
    </div>
  )
}
