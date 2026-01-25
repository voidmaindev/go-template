import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Users, Trash2, Eye, Plus } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import DataTable, { Column } from '../components/DataTable'
import Modal, { ConfirmDialog } from '../components/Modal'
import FormField from '../components/FormField'
import { useUsers, useUser, useDeleteUser, useCreateUser } from '../hooks/useUsers'
import { useUserRoles, useRoles, useAssignRole, useRemoveRole } from '../hooks/useRBAC'
import { formatDateTime } from '../lib/utils'
import type { User, QueryParams, RegisterRequest } from '../types'

export default function UsersPage() {
  const [params, setParams] = useState<QueryParams>({
    page: 1,
    page_size: 10,
  })
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null)
  const [viewModalOpen, setViewModalOpen] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [selectedRole, setSelectedRole] = useState('')
  const [createModalOpen, setCreateModalOpen] = useState(false)

  const { data, isLoading } = useUsers(params)
  const { data: selectedUser } = useUser(selectedUserId || 0)
  const { data: userRoles, refetch: refetchRoles } = useUserRoles(selectedUserId || 0)
  const { data: allRoles } = useRoles({ page_size: 100 })
  const { mutate: deleteUser, isPending: isDeleting } = useDeleteUser()
  const { mutate: assignRole, isPending: isAssigning } = useAssignRole()
  const { mutate: removeRole, isPending: isRemoving } = useRemoveRole()
  const { mutate: createUser, isPending: isCreating } = useCreateUser()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<RegisterRequest>()

  const columns: Column<User>[] = [
    {
      key: 'id',
      header: 'ID',
      sortable: true,
      render: (user) => <span className="font-mono text-neon-cyan">#{user.id}</span>,
    },
    {
      key: 'name',
      header: 'Name',
      sortable: true,
    },
    {
      key: 'email',
      header: 'Email',
      sortable: true,
      render: (user) => <span className="text-gray-400">{user.email}</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      sortable: true,
      render: (user) => <span className="font-mono text-sm">{formatDateTime(user.created_at)}</span>,
    },
  ]

  const handleView = (user: User) => {
    setSelectedUserId(user.id)
    setViewModalOpen(true)
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteUser(deleteId, {
        onSuccess: () => setDeleteId(null),
      })
    }
  }

  const handleAssignRole = () => {
    if (selectedUserId && selectedRole) {
      assignRole(
        { userId: selectedUserId, data: { role_code: selectedRole } },
        {
          onSuccess: () => {
            setSelectedRole('')
            setRoleModalOpen(false)
            refetchRoles()
          },
        }
      )
    }
  }

  const handleRemoveRole = (roleCode: string) => {
    if (selectedUserId) {
      removeRole(
        { userId: selectedUserId, roleCode },
        {
          onSuccess: () => refetchRoles(),
        }
      )
    }
  }

  const openCreateModal = () => {
    reset({ email: '', password: '', name: '', role_codes: [] })
    setCreateModalOpen(true)
  }

  const closeCreateModal = () => {
    setCreateModalOpen(false)
    reset()
  }

  const onCreateSubmit = (formData: RegisterRequest) => {
    createUser(formData, { onSuccess: closeCreateModal })
  }

  return (
    <div>
      <PageHeader
        title="Users"
        subtitle={`${data?.total ?? 0} users registered`}
        icon={Users}
        action={
          <button onClick={openCreateModal} className="cyber-button-green flex items-center gap-2">
            <Plus className="w-5 h-5" />
            Create User
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
        searchPlaceholder="Search by name..."
        actions={(user) => (
          <div className="flex items-center gap-2">
            <button
              onClick={() => handleView(user)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-cyan transition-colors"
              title="View"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => setDeleteId(user.id)}
              className="p-2 hover:bg-cyber-light rounded-lg text-gray-400 hover:text-neon-pink transition-colors"
              title="Delete"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        )}
      />

      {/* View User Modal */}
      <Modal
        isOpen={viewModalOpen}
        onClose={() => {
          setViewModalOpen(false)
          setSelectedUserId(null)
        }}
        title="User Details"
        size="lg"
      >
        {selectedUser && (
          <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">ID</p>
                <p className="text-neon-cyan font-mono">#{selectedUser.id}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Name</p>
                <p className="text-white">{selectedUser.name}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border col-span-2">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Email</p>
                <p className="text-white">{selectedUser.email}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Created</p>
                <p className="text-white font-mono text-sm">{formatDateTime(selectedUser.created_at)}</p>
              </div>
              <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
                <p className="text-xs text-gray-500 font-mono uppercase mb-1">Updated</p>
                <p className="text-white font-mono text-sm">{formatDateTime(selectedUser.updated_at)}</p>
              </div>
            </div>

            {/* User Roles */}
            <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <div className="flex items-center justify-between mb-4">
                <p className="text-xs text-gray-500 font-mono uppercase">Assigned Roles</p>
                <button
                  onClick={() => setRoleModalOpen(true)}
                  className="cyber-button-ghost text-xs"
                >
                  + Assign Role
                </button>
              </div>
              <div className="flex flex-wrap gap-2">
                {userRoles?.roles?.length ? (
                  userRoles.roles.map((role) => (
                    <span
                      key={role.code}
                      className="inline-flex items-center gap-2 px-3 py-1 bg-neon-cyan/10 border border-neon-cyan/30 rounded-full text-sm text-neon-cyan"
                    >
                      {role.name}
                      <button
                        onClick={() => handleRemoveRole(role.code)}
                        disabled={isRemoving}
                        className="hover:text-neon-pink transition-colors"
                      >
                        ×
                      </button>
                    </span>
                  ))
                ) : (
                  <span className="text-gray-500 text-sm">No roles assigned</span>
                )}
              </div>
            </div>
          </div>
        )}
      </Modal>

      {/* Assign Role Modal */}
      <Modal
        isOpen={roleModalOpen}
        onClose={() => setRoleModalOpen(false)}
        title="Assign Role"
        size="sm"
      >
        <div className="space-y-4">
          <select
            value={selectedRole}
            onChange={(e) => setSelectedRole(e.target.value)}
            className="cyber-input"
          >
            <option value="">Select a role...</option>
            {allRoles?.data
              ?.filter((r) => !userRoles?.roles?.some((ur) => ur.code === r.code))
              .map((role) => (
                <option key={role.code} value={role.code}>
                  {role.name} ({role.code})
                </option>
              ))}
          </select>
          <div className="flex justify-end gap-3">
            <button onClick={() => setRoleModalOpen(false)} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              onClick={handleAssignRole}
              disabled={!selectedRole || isAssigning}
              className="cyber-button"
            >
              {isAssigning ? 'Assigning...' : 'Assign'}
            </button>
          </div>
        </div>
      </Modal>

      {/* Create User Modal */}
      <Modal
        isOpen={createModalOpen}
        onClose={closeCreateModal}
        title="Create User"
      >
        <form onSubmit={handleSubmit(onCreateSubmit)} className="space-y-6">
          <FormField
            label="Email"
            type="email"
            {...register('email', {
              required: 'Email is required',
              pattern: {
                value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                message: 'Invalid email address',
              },
            })}
            error={errors.email?.message}
            placeholder="user@example.com"
            required
          />

          <FormField
            label="Name"
            {...register('name', {
              required: 'Name is required',
              minLength: { value: 1, message: 'Name is required' },
              maxLength: { value: 200, message: 'Name is too long' },
            })}
            error={errors.name?.message}
            placeholder="Enter user name"
            required
          />

          <FormField
            label="Password"
            type="password"
            {...register('password', {
              required: 'Password is required',
              minLength: { value: 8, message: 'Password must be at least 8 characters' },
              pattern: {
                value: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]).{8,}$/,
                message: 'Password must include uppercase, lowercase, number, and special character',
              },
            })}
            error={errors.password?.message}
            placeholder="Enter password"
            hint="8+ chars with uppercase, lowercase, number, and special character"
            required
          />

          <div className="flex justify-end gap-3 pt-4">
            <button type="button" onClick={closeCreateModal} className="cyber-button-ghost">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isCreating}
              className="cyber-button"
            >
              {isCreating ? 'Creating...' : 'Create User'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        onConfirm={handleDelete}
        title="Delete User"
        message="Are you sure you want to delete this user? This action cannot be undone."
        confirmText="Delete"
        isLoading={isDeleting}
        variant="danger"
      />
    </div>
  )
}
