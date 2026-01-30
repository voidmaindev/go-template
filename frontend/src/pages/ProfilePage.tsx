import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { motion } from 'framer-motion'
import { User, Mail, Calendar, Key, Eye, EyeOff, Loader2, Shield, Link as LinkIcon, CheckCircle, XCircle, Plus } from 'lucide-react'
import toast from 'react-hot-toast'
import PageHeader from '../components/PageHeader'
import FormField from '../components/FormField'
import OAuthButton from '../components/OAuthButton'
import LinkedAccountCard from '../components/LinkedAccountCard'
import { useUpdateProfile, useChangePassword } from '../hooks/useAuth'
import { useIdentities, useUnlinkIdentity, useLinkIdentity, useSetPassword } from '../hooks/useSelfAuth'
import { useAuthStore } from '../store/auth'
import { formatDateTime } from '../lib/utils'
import type { UpdateProfileRequest, ChangePasswordRequest, SetPasswordRequest, OAuthProvider } from '../types'

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user)
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showSetPassword, setShowSetPassword] = useState(false)
  const [linkingProvider, setLinkingProvider] = useState<OAuthProvider | null>(null)

  const { mutate: updateProfile, isPending: isUpdating } = useUpdateProfile()
  const { mutate: changePassword, isPending: isChangingPassword } = useChangePassword()
  const { data: identities, isLoading: isLoadingIdentities } = useIdentities()
  const { mutate: unlinkIdentity, isPending: isUnlinking } = useUnlinkIdentity()
  const { openLinkPopup, onLinkSuccess } = useLinkIdentity()
  const { mutate: setPassword, isPending: isSettingPassword } = useSetPassword()

  const profileForm = useForm<UpdateProfileRequest>({
    defaultValues: {
      name: user?.name || '',
    },
  })

  const passwordForm = useForm<ChangePasswordRequest>()
  const setPasswordForm = useForm<SetPasswordRequest>()

  // Handle OAuth link message from popup
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return

      if (event.data?.type === 'link-success') {
        onLinkSuccess()
      } else if (event.data?.type === 'oauth-error') {
        toast.error(event.data.error || 'Failed to link account')
      }
      setLinkingProvider(null)
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [onLinkSuccess])

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

  const onSetPassword = (data: SetPasswordRequest) => {
    setPassword(data, {
      onSuccess: () => {
        setPasswordForm.reset()
      },
    })
  }

  const handleLinkAccount = (provider: OAuthProvider) => {
    setLinkingProvider(provider)
    const popup = openLinkPopup(provider)

    if (!popup) {
      toast.error('Popup was blocked. Please allow popups for this site.')
      setLinkingProvider(null)
      return
    }

    const checkClosed = setInterval(() => {
      if (popup.closed) {
        clearInterval(checkClosed)
        setLinkingProvider(null)
      }
    }, 500)
  }

  const linkedProviders = identities?.map((i) => i.provider) || []
  const availableProviders: OAuthProvider[] = (['google', 'facebook', 'apple'] as OAuthProvider[])
    .filter((p) => !linkedProviders.includes(p))

  // Can unlink if user has password OR more than one identity
  const canUnlink = user?.has_password || (identities?.length || 0) > 1

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
              <div className="flex-1">
                <p className="text-xs text-gray-500 font-mono uppercase">Email</p>
                <p className="text-white">{user?.email}</p>
              </div>
              {/* Email Verification Badge */}
              {user?.email_verified_at ? (
                <div className="flex items-center gap-1 px-2 py-1 bg-neon-green/10 rounded text-neon-green text-xs">
                  <CheckCircle className="w-3 h-3" />
                  Verified
                </div>
              ) : (
                <div className="flex items-center gap-1 px-2 py-1 bg-neon-pink/10 rounded text-neon-pink text-xs">
                  <XCircle className="w-3 h-3" />
                  Unverified
                </div>
              )}
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

        {/* Linked Accounts */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="cyber-card p-6 lg:col-span-2"
        >
          <h2 className="font-display font-bold text-lg text-neon-cyan tracking-wider mb-6 flex items-center gap-2">
            <LinkIcon className="w-5 h-5" />
            LINKED ACCOUNTS
          </h2>

          {isLoadingIdentities ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-neon-cyan" />
            </div>
          ) : (
            <div className="space-y-4">
              {/* Linked Accounts List */}
              {identities && identities.length > 0 ? (
                <div className="space-y-3">
                  {identities.map((identity) => (
                    <LinkedAccountCard
                      key={identity.id}
                      identity={identity}
                      onUnlink={unlinkIdentity}
                      isUnlinking={isUnlinking}
                      canUnlink={canUnlink}
                    />
                  ))}
                </div>
              ) : (
                <p className="text-gray-400 text-sm">No linked accounts</p>
              )}

              {/* Link New Account */}
              {availableProviders.length > 0 && (
                <div className="pt-4 border-t border-cyber-border">
                  <p className="text-sm text-gray-400 mb-3 flex items-center gap-2">
                    <Plus className="w-4 h-4" />
                    Link a new account
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    {availableProviders.map((provider) => (
                      <OAuthButton
                        key={provider}
                        provider={provider}
                        onClick={() => handleLinkAccount(provider)}
                        isLoading={linkingProvider === provider}
                        mode="link"
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </motion.div>

        {/* Set Password (for OAuth users without password) */}
        {user && !user.has_password && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="cyber-card p-6 lg:col-span-2"
          >
            <h2 className="font-display font-bold text-lg text-neon-green tracking-wider mb-6 flex items-center gap-2">
              <Shield className="w-5 h-5" />
              SET PASSWORD
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              You signed up with a social account. Set a password to also log in with your email.
            </p>
            <form onSubmit={setPasswordForm.handleSubmit(onSetPassword)} className="space-y-6">
              <div className="space-y-2 max-w-md">
                <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                  New Password <span className="text-neon-pink">*</span>
                </label>
                <div className="relative">
                  <input
                    type={showSetPassword ? 'text' : 'password'}
                    {...setPasswordForm.register('new_password', {
                      required: 'Password is required',
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
                    onClick={() => setShowSetPassword(!showSetPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-neon-cyan"
                  >
                    {showSetPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
                {setPasswordForm.formState.errors.new_password && (
                  <p className="text-sm text-neon-pink font-mono">
                    {setPasswordForm.formState.errors.new_password.message}
                  </p>
                )}
                <p className="text-xs text-gray-500">Minimum 8 characters required</p>
              </div>
              <button
                type="submit"
                disabled={isSettingPassword}
                className="cyber-button-green flex items-center justify-center gap-2"
              >
                {isSettingPassword ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    SETTING PASSWORD...
                  </>
                ) : (
                  'SET PASSWORD'
                )}
              </button>
            </form>
          </motion.div>
        )}

        {/* Change Password Form (only show if user has password) */}
        {user?.has_password && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
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
        )}
      </div>
    </div>
  )
}
