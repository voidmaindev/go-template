import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { motion } from 'framer-motion'
import { User, Mail, Calendar, Key, Eye, EyeOff, Loader2 } from 'lucide-react'
import PageHeader from '../components/PageHeader'
import FormField from '../components/FormField'
import { useUpdateProfile, useChangePassword } from '../hooks/useAuth'
import { useAuthStore } from '../store/auth'
import { formatDateTime } from '../lib/utils'
import type { UpdateProfileRequest, ChangePasswordRequest } from '../types'

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user)
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)

  const { mutate: updateProfile, isPending: isUpdating } = useUpdateProfile()
  const { mutate: changePassword, isPending: isChangingPassword } = useChangePassword()

  const profileForm = useForm<UpdateProfileRequest>({
    defaultValues: {
      name: user?.name || '',
    },
  })

  const passwordForm = useForm<ChangePasswordRequest>()

  const onUpdateProfile = (data: UpdateProfileRequest) => {
    updateProfile(data)
  }

  const onChangePassword = (data: ChangePasswordRequest) => {
    changePassword(data, {
      onSuccess: () => {
        passwordForm.reset()
      },
    })
  }

  return (
    <div>
      <PageHeader
        title="Profile"
        subtitle="Manage your account settings"
        icon={User}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* User Info Card */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="cyber-card p-6"
        >
          <h2 className="font-display font-bold text-lg text-neon-cyan tracking-wider mb-6">
            USER INFORMATION
          </h2>
          <div className="space-y-4">
            <div className="flex items-center gap-4 p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <div className="w-10 h-10 rounded-lg bg-neon-cyan/20 flex items-center justify-center">
                <User className="w-5 h-5 text-neon-cyan" />
              </div>
              <div>
                <p className="text-xs text-gray-500 font-mono uppercase">User ID</p>
                <p className="text-white font-mono">#{user?.id}</p>
              </div>
            </div>
            <div className="flex items-center gap-4 p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <div className="w-10 h-10 rounded-lg bg-neon-green/20 flex items-center justify-center">
                <Mail className="w-5 h-5 text-neon-green" />
              </div>
              <div>
                <p className="text-xs text-gray-500 font-mono uppercase">Email</p>
                <p className="text-white">{user?.email}</p>
              </div>
            </div>
            <div className="flex items-center gap-4 p-4 bg-cyber-darker rounded-lg border border-cyber-border">
              <div className="w-10 h-10 rounded-lg bg-neon-pink/20 flex items-center justify-center">
                <Calendar className="w-5 h-5 text-neon-pink" />
              </div>
              <div>
                <p className="text-xs text-gray-500 font-mono uppercase">Created At</p>
                <p className="text-white">{user?.created_at ? formatDateTime(user.created_at) : '-'}</p>
              </div>
            </div>
          </div>
        </motion.div>

        {/* Update Profile Form */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="cyber-card p-6"
        >
          <h2 className="font-display font-bold text-lg text-neon-cyan tracking-wider mb-6">
            UPDATE PROFILE
          </h2>
          <form onSubmit={profileForm.handleSubmit(onUpdateProfile)} className="space-y-6">
            <FormField
              label="Display Name"
              {...profileForm.register('name', {
                required: 'Name is required',
                minLength: {
                  value: 2,
                  message: 'Name must be at least 2 characters',
                },
              })}
              error={profileForm.formState.errors.name?.message}
              placeholder="Your name"
            />
            <button
              type="submit"
              disabled={isUpdating}
              className="cyber-button w-full flex items-center justify-center gap-2"
            >
              {isUpdating ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  UPDATING...
                </>
              ) : (
                'UPDATE PROFILE'
              )}
            </button>
          </form>
        </motion.div>

        {/* Change Password Form */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="cyber-card p-6 lg:col-span-2"
        >
          <h2 className="font-display font-bold text-lg text-neon-pink tracking-wider mb-6 flex items-center gap-2">
            <Key className="w-5 h-5" />
            CHANGE PASSWORD
          </h2>
          <form onSubmit={passwordForm.handleSubmit(onChangePassword)} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                  Current Password <span className="text-neon-pink">*</span>
                </label>
                <div className="relative">
                  <input
                    type={showCurrentPassword ? 'text' : 'password'}
                    {...passwordForm.register('current_password', {
                      required: 'Current password is required',
                    })}
                    className="cyber-input pr-12"
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-neon-cyan"
                  >
                    {showCurrentPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
                {passwordForm.formState.errors.current_password && (
                  <p className="text-sm text-neon-pink font-mono">
                    {passwordForm.formState.errors.current_password.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                  New Password <span className="text-neon-pink">*</span>
                </label>
                <div className="relative">
                  <input
                    type={showNewPassword ? 'text' : 'password'}
                    {...passwordForm.register('new_password', {
                      required: 'New password is required',
                      minLength: {
                        value: 8,
                        message: 'Password must be at least 8 characters',
                      },
                    })}
                    className="cyber-input pr-12"
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowNewPassword(!showNewPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-neon-cyan"
                  >
                    {showNewPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
                {passwordForm.formState.errors.new_password && (
                  <p className="text-sm text-neon-pink font-mono">
                    {passwordForm.formState.errors.new_password.message}
                  </p>
                )}
              </div>
            </div>
            <button
              type="submit"
              disabled={isChangingPassword}
              className="cyber-button-pink flex items-center justify-center gap-2"
            >
              {isChangingPassword ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  CHANGING...
                </>
              ) : (
                'CHANGE PASSWORD'
              )}
            </button>
          </form>
        </motion.div>
      </div>
    </div>
  )
}
