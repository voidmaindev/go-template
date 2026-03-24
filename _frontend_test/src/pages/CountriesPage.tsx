import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Globe, Plus, Edit2, Trash2, Eye } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField from '../components/FormField'
import {
  useCountries,
  useCountry,
  useCountryCities,
  useCreateCountry,
  useUpdateCountry,
  useDeleteCountry,
} from '../hooks/useCountries'
import { formatDateTime } from '../lib/utils'
import type { Country, CreateCountryRequest, City, QueryParams } from '../types'

export default function CountriesPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingCountry, setEditingCountry] = useState<Country | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [viewCountryId, setViewCountryId] = useState<number | null>(null)

  const { data, isLoading } = useCountries(params)
  const { data: viewCountry } = useCountry(viewCountryId || 0)
  const { data: countryCities } = useCountryCities(viewCountryId || 0, { page_size: 100 })
  const { mutate: createCountry, isPending: isCreating } = useCreateCountry()
  const { mutate: updateCountry, isPending: isUpdating } = useUpdateCountry()
  const { mutate: deleteCountry, isPending: isDeleting } = useDeleteCountry()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateCountryRequest>()

  const columns: Column<Country>[] = [
    {
      key: 'id',
      header: 'ID',
      sortable: true,
      render: (country) => <span className="font-mono text-neon-cyan">#{country.id}</span>,
    },
    {
      key: 'code',
      header: 'Code',
      sortable: true,
      render: (country) => (
        <span className="font-mono text-neon-green uppercase">{country.code}</span>
      ),
    },
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (country) => <span className="font-medium text-white">{country.name}</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (country) => (
        <span className="font-mono text-sm">{formatDateTime(country.created_at)}</span>
      ),
    },
  ]

  const openCreateModal = () => {
    setEditingCountry(null)
    reset({ name: '', code: '' })
    setModalOpen(true)
  }

  const openEditModal = (country: Country) => {
    setEditingCountry(country)
    reset({ name: country.name, code: country.code })
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingCountry(null)
    reset()
  }

  const onSubmit = (data: CreateCountryRequest) => {
    const payload = {
      name: data.name,
      code: data.code.toUpperCase(),
    }

    if (editingCountry) {
      updateCountry(
        { id: editingCountry.id, data: payload },
        { onSuccess: closeModal }
      )
    } else {
      createCountry(payload, { onSuccess: closeModal })
    }
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteCountry(deleteId, {
        onSuccess: () => setDeleteId(null),
      })
    }
  }

  return (
    <div>
      <PageHeader
        title="Countries"
        subtitle={`${data?.total ?? 0} countries registered`}
        icon={Globe}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New Country
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
        searchPlaceholder="Search countries..."
        actions={(country) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setViewCountryId(country.id)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-green transition-colors"
              title="View Cities"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => openEditModal(country)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="Edit"
            >
              <Edit2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setDeleteId(country.id)}
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
        title={editingCountry ? 'Edit Country' : 'Create Country'}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            label="Country Name"
            {...register('name', {
              required: 'Name is required',
              minLength: { value: 1, message: 'Name is required' },
              maxLength: { value: 100, message: 'Name is too long' },
            })}
            error={errors.name?.message}
            placeholder="United States"
            required
          />

          <FormField
            label="ISO Code (3 letters)"
            {...register('code', {
              required: 'Code is required',
              pattern: {
                value: /^[A-Za-z]{3}$/,
                message: 'Must be exactly 3 letters',
              },
            })}
            error={errors.code?.message}
            placeholder="USA"
            maxLength={3}
            className="uppercase"
            hint="ISO 3166-1 alpha-3 code (e.g., USA, DEU, GBR)"
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
                : editingCountry
                ? 'Update Country'
                : 'Create Country'}
            </button>
          </div>
        </form>
      </Modal>

      {/* View Country Modal */}
      <Modal
        isOpen={!!viewCountryId}
        onClose={() => setViewCountryId(null)}
        title="Country Details"
        size="lg"
      >
        {viewCountry && (
          <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">ID</p>
                <p className="text-neon-cyan font-mono">#{viewCountry.id}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Code</p>
                <p className="text-neon-green font-mono uppercase">{viewCountry.code}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border col-span-2">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Name</p>
                <p className="text-white">{viewCountry.name}</p>
              </div>
            </div>

            {/* Cities in this country */}
            <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <p className="text-xs text-gray-500 font-mono uppercase mb-4">
                Cities ({countryCities?.total || 0})
              </p>
              {countryCities?.data?.length ? (
                <div className="flex flex-wrap gap-2">
                  {countryCities.data.map((city: City) => (
                    <span
                      key={city.id}
                      className="px-3 py-1 bg-neon-cyan/10 border border-neon-cyan/30 rounded-full text-sm text-neon-cyan"
                    >
                      {city.name}
                    </span>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-sm">No cities found</p>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        onConfirm={handleDelete}
        title="Delete Country"
        message="Are you sure you want to delete this country? All associated cities will also be affected."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />
    </div>
  )
}
