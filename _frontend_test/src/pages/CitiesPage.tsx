import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Building2, Plus, Edit2, Trash2 } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField, { SelectField } from '../components/FormField'
import { useCities, useCreateCity, useUpdateCity, useDeleteCity } from '../hooks/useCities'
import { useCountries } from '../hooks/useCountries'
import { formatDateTime } from '../lib/utils'
import type { City, CreateCityRequest, QueryParams } from '../types'

export default function CitiesPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingCity, setEditingCity] = useState<City | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const { data, isLoading } = useCities(params)
  const { data: countriesData } = useCountries({ page_size: 100 })
  const { mutate: createCity, isPending: isCreating } = useCreateCity()
  const { mutate: updateCity, isPending: isUpdating } = useUpdateCity()
  const { mutate: deleteCity, isPending: isDeleting } = useDeleteCity()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateCityRequest>()

  const columns: Column<City>[] = [
    {
      key: 'id',
      header: 'ID',
      sortable: true,
      render: (city) => <span className="font-mono text-neon-cyan">#{city.id}</span>,
    },
    {
      key: 'name',
      header: 'City Name',
      sortable: true,
      render: (city) => <span className="font-medium text-white">{city.name}</span>,
    },
    {
      key: 'country',
      header: 'Country',
      render: (city) => (
        <div className="flex items-center gap-2">
          <span className="font-mono text-neon-green text-xs uppercase">
            {city.country?.code}
          </span>
          <span className="text-gray-400">{city.country?.name || '-'}</span>
        </div>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (city) => (
        <span className="font-mono text-sm">{formatDateTime(city.created_at)}</span>
      ),
    },
  ]

  const countryOptions = (countriesData?.data || []).map((c) => ({
    value: c.id,
    label: `${c.name} (${c.code})`,
  }))

  const openCreateModal = () => {
    setEditingCity(null)
    reset({ name: '', country_id: 0 })
    setModalOpen(true)
  }

  const openEditModal = (city: City) => {
    setEditingCity(city)
    reset({ name: city.name, country_id: city.country_id })
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingCity(null)
    reset()
  }

  const onSubmit = (data: CreateCityRequest) => {
    const payload = {
      name: data.name,
      country_id: Number(data.country_id),
    }

    if (editingCity) {
      updateCity(
        { id: editingCity.id, data: payload },
        { onSuccess: closeModal }
      )
    } else {
      createCity(payload, { onSuccess: closeModal })
    }
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteCity(deleteId, {
        onSuccess: () => setDeleteId(null),
      })
    }
  }

  return (
    <div>
      <PageHeader
        title="Cities"
        subtitle={`${data?.total ?? 0} cities registered`}
        icon={Building2}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New City
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
        searchPlaceholder="Search cities..."
        actions={(city) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => openEditModal(city)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="Edit"
            >
              <Edit2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setDeleteId(city.id)}
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
        title={editingCity ? 'Edit City' : 'Create City'}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            label="City Name"
            {...register('name', {
              required: 'Name is required',
              minLength: { value: 1, message: 'Name is required' },
              maxLength: { value: 100, message: 'Name is too long' },
            })}
            error={errors.name?.message}
            placeholder="New York"
            required
          />

          <SelectField
            label="Country"
            {...register('country_id', {
              required: 'Country is required',
              validate: (value) => Number(value) > 0 || 'Please select a country',
            })}
            error={errors.country_id?.message}
            options={countryOptions}
            placeholder="Select a country"
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
                : editingCity
                ? 'Update City'
                : 'Create City'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        onConfirm={handleDelete}
        title="Delete City"
        message="Are you sure you want to delete this city? This action cannot be undone."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />
    </div>
  )
}
