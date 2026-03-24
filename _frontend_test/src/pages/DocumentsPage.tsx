import { useState } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import { FileText, Plus, Edit2, Trash2, Eye, Package } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField, { SelectField } from '../components/FormField'
import {
  useDocuments,
  useDocument,
  useCreateDocument,
  useUpdateDocument,
  useDeleteDocument,
  useAddDocumentItem,
  useUpdateDocumentItem,
  useDeleteDocumentItem,
} from '../hooks/useDocuments'
import { useCities } from '../hooks/useCities'
import { useItems } from '../hooks/useItems'
import { formatPrice, priceToCents, formatDate, formatDateForInput } from '../lib/utils'
import type { Document, CreateDocumentRequest, QueryParams, DocumentItem } from '../types'

interface FormData {
  code: string
  city_id: number
  document_date: string
  items: Array<{
    item_id: number
    quantity: number
    priceDisplay: string
  }>
}

export default function DocumentsPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingDocument, setEditingDocument] = useState<Document | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [viewDocumentId, setViewDocumentId] = useState<number | null>(null)
  const [addItemModalOpen, setAddItemModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<DocumentItem | null>(null)
  const [deleteItemId, setDeleteItemId] = useState<number | null>(null)

  const { data, isLoading } = useDocuments(params)
  const { data: viewDocument, refetch: refetchDocument } = useDocument(viewDocumentId || 0)
  const { data: citiesData } = useCities({ page_size: 100 })
  const { data: itemsData } = useItems({ page_size: 100 })
  const { mutate: createDocument, isPending: isCreating } = useCreateDocument()
  const { mutate: updateDocument, isPending: isUpdating } = useUpdateDocument()
  const { mutate: deleteDocument, isPending: isDeleting } = useDeleteDocument()
  const { mutate: addDocumentItem, isPending: isAddingItem } = useAddDocumentItem()
  const { mutate: updateDocumentItem, isPending: isUpdatingItem } = useUpdateDocumentItem()
  const { mutate: deleteDocumentItem, isPending: isDeletingItem } = useDeleteDocumentItem()

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<FormData>({
    defaultValues: {
      items: [{ item_id: 0, quantity: 1, priceDisplay: '' }],
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items',
  })

  const addItemForm = useForm<{
    item_id: number
    quantity: number
    priceDisplay: string
  }>()

  const columns: Column<Document>[] = [
    {
      key: 'id',
      header: 'ID',
      sortable: true,
      render: (doc) => <span className="font-mono text-neon-cyan">#{doc.id}</span>,
    },
    {
      key: 'code',
      header: 'Code',
      sortable: true,
      render: (doc) => <span className="font-mono text-neon-green">{doc.code}</span>,
    },
    {
      key: 'city',
      header: 'City',
      render: (doc) => (
        <span className="text-gray-400">
          {doc.city?.name || '-'}
          {doc.city?.country && (
            <span className="text-gray-600 ml-1">({doc.city.country.code})</span>
          )}
        </span>
      ),
    },
    {
      key: 'document_date',
      header: 'Date',
      sortable: true,
      render: (doc) => <span className="font-mono text-sm">{formatDate(doc.document_date)}</span>,
    },
    {
      key: 'total_amount',
      header: 'Total',
      sortable: true,
      render: (doc) => (
        <span className="font-mono text-neon-yellow">{formatPrice(doc.total_amount)}</span>
      ),
    },
  ]

  const cityOptions = (citiesData?.data || []).map((c) => ({
    value: c.id,
    label: `${c.name} (${c.country?.code || ''})`,
  }))

  const itemOptions = (itemsData?.data || []).map((i) => ({
    value: i.id,
    label: `${i.name} - ${formatPrice(i.price)}`,
  }))

  const openCreateModal = () => {
    setEditingDocument(null)
    reset({
      code: '',
      city_id: 0,
      document_date: new Date().toISOString().split('T')[0],
      items: [{ item_id: 0, quantity: 1, priceDisplay: '' }],
    })
    setModalOpen(true)
  }

  const openEditModal = (doc: Document) => {
    setEditingDocument(doc)
    reset({
      code: doc.code,
      city_id: doc.city_id,
      document_date: formatDateForInput(doc.document_date),
      items: [],
    })
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingDocument(null)
    reset()
  }

  const onSubmit = (data: FormData) => {
    if (editingDocument) {
      updateDocument(
        {
          id: editingDocument.id,
          data: {
            code: data.code,
            city_id: Number(data.city_id),
            document_date: data.document_date,
          },
        },
        { onSuccess: closeModal }
      )
    } else {
      const payload: CreateDocumentRequest = {
        code: data.code,
        city_id: Number(data.city_id),
        document_date: data.document_date,
        items: data.items
          .filter((i) => i.item_id > 0)
          .map((i) => ({
            item_id: Number(i.item_id),
            quantity: Number(i.quantity),
            price: priceToCents(i.priceDisplay),
          })),
      }
      createDocument(payload, { onSuccess: closeModal })
    }
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteDocument(deleteId, {
        onSuccess: () => setDeleteId(null),
      })
    }
  }

  const openAddItemModal = () => {
    setEditingItem(null)
    addItemForm.reset({ item_id: 0, quantity: 1, priceDisplay: '' })
    setAddItemModalOpen(true)
  }

  const openEditItemModal = (item: DocumentItem) => {
    setEditingItem(item)
    addItemForm.reset({
      item_id: item.item_id,
      quantity: item.quantity,
      priceDisplay: (item.price / 100).toFixed(2),
    })
    setAddItemModalOpen(true)
  }

  const closeAddItemModal = () => {
    setAddItemModalOpen(false)
    setEditingItem(null)
    addItemForm.reset()
  }

  const onAddItem = (data: { item_id: number; quantity: number; priceDisplay: string }) => {
    if (!viewDocumentId) return

    if (editingItem) {
      updateDocumentItem(
        {
          documentId: viewDocumentId,
          itemId: editingItem.id,
          data: {
            quantity: Number(data.quantity),
            price: priceToCents(data.priceDisplay),
          },
        },
        {
          onSuccess: () => {
            closeAddItemModal()
            refetchDocument()
          },
        }
      )
    } else {
      addDocumentItem(
        {
          documentId: viewDocumentId,
          data: {
            item_id: Number(data.item_id),
            quantity: Number(data.quantity),
            price: priceToCents(data.priceDisplay),
          },
        },
        {
          onSuccess: () => {
            closeAddItemModal()
            refetchDocument()
          },
        }
      )
    }
  }

  const handleDeleteItem = () => {
    if (deleteItemId && viewDocumentId) {
      deleteDocumentItem(
        { documentId: viewDocumentId, itemId: deleteItemId },
        {
          onSuccess: () => {
            setDeleteItemId(null)
            refetchDocument()
          },
        }
      )
    }
  }

  return (
    <div>
      <PageHeader
        title="Documents"
        subtitle={`${data?.total ?? 0} documents created`}
        icon={FileText}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New Document
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
        searchValue={params['code__contains'] as string}
        onSearchChange={(value) =>
          setParams({ ...params, 'code__contains': value || undefined, page: 1 })
        }
        searchPlaceholder="Search by code..."
        actions={(doc) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setViewDocumentId(doc.id)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-green transition-colors"
              title="View"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => openEditModal(doc)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="Edit"
            >
              <Edit2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setDeleteId(doc.id)}
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
        title={editingDocument ? 'Edit Document' : 'Create Document'}
        size="lg"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FormField
              label="Document Code"
              {...register('code', {
                required: 'Code is required',
                maxLength: { value: 50, message: 'Code is too long' },
              })}
              error={errors.code?.message}
              placeholder="DOC-001"
              required
            />

            <FormField
              label="Document Date"
              type="date"
              {...register('document_date', {
                required: 'Date is required',
              })}
              error={errors.document_date?.message}
              required
            />

            <div className="md:col-span-2">
              <SelectField
                label="City"
                {...register('city_id', {
                  required: 'City is required',
                  validate: (value) => Number(value) > 0 || 'Please select a city',
                })}
                error={errors.city_id?.message}
                options={cityOptions}
                placeholder="Select a city"
                required
              />
            </div>
          </div>

          {/* Line Items (only for create) */}
          {!editingDocument && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                  Line Items <span className="text-neon-pink">*</span>
                </label>
                <button
                  type="button"
                  onClick={() => append({ item_id: 0, quantity: 1, priceDisplay: '' })}
                  className="cyber-button-ghost text-xs"
                >
                  + Add Item
                </button>
              </div>

              {fields.map((field, index) => (
                <div
                  key={field.id}
                  className="grid grid-cols-12 gap-2 p-4 bg-cyber-darker rounded-lg border border-cyber-border"
                >
                  <div className="col-span-5">
                    <SelectField
                      label="Item"
                      {...register(`items.${index}.item_id` as const, {
                        validate: (value) => Number(value) > 0 || 'Select an item',
                      })}
                      options={itemOptions}
                      placeholder="Select item"
                      required
                    />
                  </div>
                  <div className="col-span-2">
                    <FormField
                      label="Qty"
                      type="number"
                      min="1"
                      {...register(`items.${index}.quantity` as const, {
                        required: true,
                        min: 1,
                      })}
                      required
                    />
                  </div>
                  <div className="col-span-4">
                    <FormField
                      label="Price ($)"
                      type="number"
                      step="0.01"
                      min="0"
                      {...register(`items.${index}.priceDisplay` as const, {
                        required: true,
                        min: 0,
                      })}
                      required
                    />
                  </div>
                  <div className="col-span-1 flex items-end pb-2">
                    {fields.length > 1 && (
                      <button
                        type="button"
                        onClick={() => remove(index)}
                        className="p-2 text-gray-500 hover:text-neon-pink"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

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
                : editingDocument
                ? 'Update Document'
                : 'Create Document'}
            </button>
          </div>
        </form>
      </Modal>

      {/* View Document Modal */}
      <Modal
        isOpen={!!viewDocumentId}
        onClose={() => setViewDocumentId(null)}
        title="Document Details"
        size="xl"
      >
        {viewDocument && (
          <div className="space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">ID</p>
                <p className="text-neon-cyan font-mono">#{viewDocument.id}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Code</p>
                <p className="text-neon-green font-mono">{viewDocument.code}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Date</p>
                <p className="text-white font-mono">{formatDate(viewDocument.document_date)}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Total</p>
                <p className="text-neon-yellow font-mono text-lg">
                  {formatPrice(viewDocument.total_amount)}
                </p>
              </div>
            </div>

            {/* Document Items */}
            <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <div className="flex items-center justify-between mb-4">
                <p className="text-sm text-gray-400 font-display uppercase tracking-wider flex items-center gap-2">
                  <Package className="w-4 h-4" />
                  Line Items ({viewDocument.items?.length || 0})
                </p>
                <button onClick={openAddItemModal} className="cyber-button-ghost text-xs">
                  + Add Item
                </button>
              </div>

              {viewDocument.items?.length ? (
                <div className="overflow-x-auto">
                  <table className="cyber-table">
                    <thead>
                      <tr>
                        <th>Item</th>
                        <th>Quantity</th>
                        <th>Unit Price</th>
                        <th>Line Total</th>
                        <th className="w-20">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {viewDocument.items.map((item) => (
                        <tr key={item.id}>
                          <td>{item.item?.name || `Item #${item.item_id}`}</td>
                          <td className="font-mono">{item.quantity}</td>
                          <td className="font-mono text-neon-cyan">{formatPrice(item.price)}</td>
                          <td className="font-mono text-neon-green">{formatPrice(item.line_total)}</td>
                          <td>
                            <div className="flex items-center gap-1">
                              <button
                                onClick={() => openEditItemModal(item)}
                                className="p-1 hover:bg-cyber-light rounded text-gray-400 hover:text-neon-cyan"
                              >
                                <Edit2 className="w-3 h-3" />
                              </button>
                              <button
                                onClick={() => setDeleteItemId(item.id)}
                                className="p-1 hover:bg-cyber-light rounded text-gray-400 hover:text-neon-pink"
                              >
                                <Trash2 className="w-3 h-3" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-gray-500 text-sm text-center py-4">No items in this document</p>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Add/Edit Item Modal */}
      <Modal
        isOpen={addItemModalOpen}
        onClose={closeAddItemModal}
        title={editingItem ? 'Edit Line Item' : 'Add Line Item'}
        size="sm"
      >
        <form onSubmit={addItemForm.handleSubmit(onAddItem)} className="space-y-4">
          {!editingItem && (
            <SelectField
              label="Item"
              {...addItemForm.register('item_id', {
                validate: (value) => Number(value) > 0 || 'Select an item',
              })}
              options={itemOptions}
              placeholder="Select item"
              required
            />
          )}

          <FormField
            label="Quantity"
            type="number"
            min="1"
            {...addItemForm.register('quantity', {
              required: 'Quantity is required',
              min: { value: 1, message: 'Must be at least 1' },
            })}
            error={addItemForm.formState.errors.quantity?.message}
            required
          />

          <FormField
            label="Unit Price ($)"
            type="number"
            step="0.01"
            min="0"
            {...addItemForm.register('priceDisplay', {
              required: 'Price is required',
              min: { value: 0, message: 'Must be positive' },
            })}
            error={addItemForm.formState.errors.priceDisplay?.message}
            required
          />

          <div className="flex justify-end gap-3 pt-4">
            <button type="button" onClick={closeAddItemModal} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isAddingItem || isUpdatingItem}
              className="cyber-button"
            >
              {isAddingItem || isUpdatingItem ? 'Saving...' : editingItem ? 'Update' : 'Add'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Delete Document Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        onConfirm={handleDelete}
        title="Delete Document"
        message="Are you sure you want to delete this document? All line items will also be deleted."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />

      {/* Delete Item Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteItemId}
        onClose={() => setDeleteItemId(null)}
        onConfirm={handleDeleteItem}
        title="Remove Item"
        message="Are you sure you want to remove this item from the document?"
        confirmText="Remove"
        isLoading={isDeletingItem}
        variant="danger"
      />
    </div>
  )
}
