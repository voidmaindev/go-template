import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Shield, Plus, Edit2, Trash2, Eye, Check } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField, { TextAreaField } from '../components/FormField'
import {
  useRoles,
  useRole,
  useCreateRole,
  useUpdateRolePermissions,
  useDeleteRole,
  useDomains,
  useActions,
} from '../hooks/useRBAC'
import { formatDateTime } from '../lib/utils'
import type { Role, CreateRoleRequest, Permission, QueryParams } from '../types'

export default function RBACPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [deleteCode, setDeleteCode] = useState<string | null>(null)
  const [viewRoleCode, setViewRoleCode] = useState<string | null>(null)
  const [permissionsModalOpen, setPermissionsModalOpen] = useState(false)
  const [selectedPermissions, setSelectedPermissions] = useState<Record<string, string[]>>({})

  const { data, isLoading } = useRoles(params)
  const { data: viewRole } = useRole(viewRoleCode || '')
  const { data: domainsData } = useDomains()
  const { data: actionsData } = useActions()
  const { mutate: createRole, isPending: isCreating } = useCreateRole()
  const { mutate: updatePermissions, isPending: isUpdatingPermissions } = useUpdateRolePermissions()
  const { mutate: deleteRole, isPending: isDeleting } = useDeleteRole()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<Omit<CreateRoleRequest, 'permissions'>>()

  const columns: Column<Role>[] = [
    {
      key: 'code',
      header: 'Code',
      sortable: true,
      render: (role) => <span className="font-mono text-neon-cyan">{role.code}</span>,
    },
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (role) => <span className="font-medium text-white">{role.name}</span>,
    },
    {
      key: 'description',
      header: 'Description',
      render: (role) => (
        <span className="text-gray-400 truncate max-w-xs block">{role.description || '-'}</span>
      ),
    },
    {
      key: 'is_system',
      header: 'System',
      render: (role) =>
        role.is_system ? (
          <span className="status-badge bg-neon-pink/20 text-neon-pink border border-neon-pink/30">
            System
          </span>
        ) : (
          <span className="status-badge bg-gray-500/20 text-gray-400 border border-gray-500/30">
            Custom
          </span>
        ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (role) => <span className="font-mono text-sm">{formatDateTime(role.created_at)}</span>,
    },
  ]

  const openCreateModal = () => {
    setEditingRole(null)
    reset({ code: '', name: '', description: '' })
    setSelectedPermissions({})
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingRole(null)
    reset()
  }

  const openPermissionsModal = (role: Role) => {
    setEditingRole(role)
    // Convert permissions array to record
    const permsMap: Record<string, string[]> = {}
    role.permissions?.forEach((p) => {
      permsMap[p.domain] = p.actions
    })
    setSelectedPermissions(permsMap)
    setPermissionsModalOpen(true)
  }

  const closePermissionsModal = () => {
    setPermissionsModalOpen(false)
    setEditingRole(null)
    setSelectedPermissions({})
  }

  const togglePermission = (domain: string, action: string) => {
    setSelectedPermissions((prev) => {
      const current = prev[domain] || []
      if (current.includes(action)) {
        const updated = current.filter((a) => a !== action)
        if (updated.length === 0) {
          const { [domain]: _, ...rest } = prev
          return rest
        }
        return { ...prev, [domain]: updated }
      } else {
        return { ...prev, [domain]: [...current, action] }
      }
    })
  }

  const onSubmit = (data: Omit<CreateRoleRequest, 'permissions'>) => {
    const permissions: Permission[] = Object.entries(selectedPermissions)
      .filter(([, actions]) => actions.length > 0)
      .map(([domain, actions]) => ({ domain, actions }))

    if (permissions.length === 0) {
      return
    }

    createRole(
      { ...data, permissions },
      { onSuccess: closeModal }
    )
  }

  const handleSavePermissions = () => {
    if (!editingRole) return

    const permissions: Permission[] = Object.entries(selectedPermissions)
      .filter(([, actions]) => actions.length > 0)
      .map(([domain, actions]) => ({ domain, actions }))

    updatePermissions(
      { code: editingRole.code, data: { permissions } },
      { onSuccess: closePermissionsModal }
    )
  }

  const handleDelete = () => {
    if (deleteCode) {
      deleteRole(deleteCode, {
        onSuccess: () => setDeleteCode(null),
      })
    }
  }

  const domains = domainsData?.domains || []
  const actions = actionsData?.actions || ['read', 'write', 'modify', 'delete']

  return (
    <div>
      <PageHeader
        title="RBAC Management"
        subtitle={`${data?.total ?? 0} roles defined`}
        icon={Shield}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New Role
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
        searchPlaceholder="Search roles..."
        actions={(role) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setViewRoleCode(role.code)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-green transition-colors"
              title="View"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => openPermissionsModal(role)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="Edit Permissions"
            >
              <Edit2 className="w-4 h-4" />
            </button>
            {!role.is_system && (
              <button
                onClick={() => setDeleteCode(role.code)}
                className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-pink transition-colors"
                title="Delete"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            )}
          </div>
        )}
      />

      {/* Create Role Modal */}
      <Modal isOpen={modalOpen} onClose={closeModal} title="Create Role" size="lg">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <FormField
              label="Role Code"
              {...register('code', {
                required: 'Code is required',
                pattern: {
                  value: /^[a-zA-Z0-9_]+$/,
                  message: 'Only alphanumeric and underscores',
                },
                minLength: { value: 2, message: 'Minimum 2 characters' },
                maxLength: { value: 50, message: 'Maximum 50 characters' },
              })}
              error={errors.code?.message}
              placeholder="custom_role"
              required
            />

            <FormField
              label="Role Name"
              {...register('name', {
                required: 'Name is required',
                minLength: { value: 2, message: 'Minimum 2 characters' },
                maxLength: { value: 100, message: 'Maximum 100 characters' },
              })}
              error={errors.name?.message}
              placeholder="Custom Role"
              required
            />
          </div>

          <TextAreaField
            label="Description"
            {...register('description')}
            placeholder="Role description (optional)"
          />

          {/* Permissions Selection */}
          <div className="space-y-4">
            <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
              Permissions <span className="text-neon-pink">*</span>
            </label>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-cyber-darker">
                    <th className="px-4 py-2 text-left font-display text-xs uppercase tracking-wider text-gray-400 border-b border-cyber-border">
                      Domain
                    </th>
                    {actions.map((action) => (
                      <th
                        key={action}
                        className="px-4 py-2 text-center font-display text-xs uppercase tracking-wider text-gray-400 border-b border-cyber-border"
                      >
                        {action}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {domains.map((domain) => (
                    <tr key={domain.name} className="border-b border-cyber-border/50">
                      <td className="px-4 py-2">
                        <span className="text-white">{domain.name}</span>
                        {domain.is_protected && (
                          <span className="ml-2 text-xs text-neon-pink">(protected)</span>
                        )}
                      </td>
                      {actions.map((action) => (
                        <td key={action} className="px-4 py-2 text-center">
                          <button
                            type="button"
                            onClick={() => togglePermission(domain.name, action)}
                            className={`w-6 h-6 rounded border ${
                              selectedPermissions[domain.name]?.includes(action)
                                ? 'bg-neon-cyan/20 border-neon-cyan text-neon-cyan'
                                : 'border-cyber-border text-gray-600 hover:border-gray-500'
                            }`}
                          >
                            {selectedPermissions[domain.name]?.includes(action) && (
                              <Check className="w-4 h-4 mx-auto" />
                            )}
                          </button>
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button type="button" onClick={closeModal} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isCreating || Object.keys(selectedPermissions).length === 0}
              className="cyber-button"
            >
              {isCreating ? 'Creating...' : 'Create Role'}
            </button>
          </div>
        </form>
      </Modal>

      {/* View Role Modal */}
      <Modal
        isOpen={!!viewRoleCode}
        onClose={() => setViewRoleCode(null)}
        title="Role Details"
        size="lg"
      >
        {viewRole && (
          <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Code</p>
                <p className="text-neon-cyan font-mono">{viewRole.code}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Name</p>
                <p className="text-white">{viewRole.name}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border col-span-2">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Description</p>
                <p className="text-white">{viewRole.description || '-'}</p>
              </div>
            </div>

            {/* Permissions */}
            <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <p className="text-xs text-gray-500 font-mono uppercase mb-4">Permissions</p>
              {viewRole.permissions?.length ? (
                <div className="space-y-2">
                  {viewRole.permissions.map((perm) => (
                    <div key={perm.domain} className="flex items-center gap-4">
                      <span className="text-white font-medium w-32">{perm.domain}</span>
                      <div className="flex flex-wrap gap-2">
                        {perm.actions.map((action) => (
                          <span
                            key={action}
                            className="px-2 py-1 bg-neon-cyan/10 border border-neon-cyan/30 rounded text-xs text-neon-cyan"
                          >
                            {action}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-sm">No permissions defined</p>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Edit Permissions Modal */}
      <Modal
        isOpen={permissionsModalOpen}
        onClose={closePermissionsModal}
        title={`Edit Permissions: ${editingRole?.name}`}
        size="lg"
      >
        <div className="space-y-6">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-cyber-darker">
                  <th className="px-4 py-2 text-left font-display text-xs uppercase tracking-wider text-gray-400 border-b border-cyber-border">
                    Domain
                  </th>
                  {actions.map((action) => (
                    <th
                      key={action}
                      className="px-4 py-2 text-center font-display text-xs uppercase tracking-wider text-gray-400 border-b border-cyber-border"
                    >
                      {action}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {domains.map((domain) => (
                  <tr key={domain.name} className="border-b border-cyber-border/50">
                    <td className="px-4 py-2">
                      <span className="text-white">{domain.name}</span>
                      {domain.is_protected && (
                        <span className="ml-2 text-xs text-neon-pink">(protected)</span>
                      )}
                    </td>
                    {actions.map((action) => (
                      <td key={action} className="px-4 py-2 text-center">
                        <button
                          type="button"
                          onClick={() => togglePermission(domain.name, action)}
                          className={`w-6 h-6 rounded border ${
                            selectedPermissions[domain.name]?.includes(action)
                              ? 'bg-neon-cyan/20 border-neon-cyan text-neon-cyan'
                              : 'border-cyber-border text-gray-600 hover:border-gray-500'
                          }`}
                        >
                          {selectedPermissions[domain.name]?.includes(action) && (
                            <Check className="w-4 h-4 mx-auto" />
                          )}
                        </button>
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button onClick={closePermissionsModal} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              onClick={handleSavePermissions}
              disabled={isUpdatingPermissions}
              className="cyber-button"
            >
              {isUpdatingPermissions ? 'Saving...' : 'Save Permissions'}
            </button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteCode}
        onClose={() => setDeleteCode(null)}
        onConfirm={handleDelete}
        title="Delete Role"
        message="Are you sure you want to delete this role? Users with this role will lose these permissions."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />
    </div>
  )
}
