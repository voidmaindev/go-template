import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useForm } from 'react-hook-form'
import { Terminal, Eye, EyeOff, Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'
import { useSelfRegister, useOAuthLogin } from '../hooks/useSelfAuth'
import { useAuthStore } from '../store/auth'
import OAuthButton from '../components/OAuthButton'
import type { SelfRegisterRequest, OAuthProvider } from '../types'

export default function RegisterPage() {
  const navigate = useNavigate()
  const [showPassword, setShowPassword] = useState(false)
  const [oauthLoading, setOauthLoading] = useState<OAuthProvider | null>(null)
  const { mutate: registerUser, isPending } = useSelfRegister()
  const setAuth = useAuthStore((state) => state.setAuth)
  const { openOAuthPopup: openGooglePopup } = useOAuthLogin('google')
  const { openOAuthPopup: openFacebookPopup } = useOAuthLogin('facebook')
  const { openOAuthPopup: openApplePopup } = useOAuthLogin('apple')

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SelfRegisterRequest>()

  // Handle OAuth message from popup
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return

      if (event.data?.type === 'oauth-success') {
        setAuth(event.data.data)
        toast.success('ACCOUNT CREATED // ACCESS GRANTED')
        navigate('/dashboard')
      } else if (event.data?.type === 'oauth-error') {
        toast.error(event.data.error || 'OAuth authentication failed')
      }
      setOauthLoading(null)
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [setAuth, navigate])

  const handleOAuthLogin = (provider: OAuthProvider) => {
    setOauthLoading(provider)
    const popup = provider === 'google'
      ? openGooglePopup()
      : provider === 'facebook'
      ? openFacebookPopup()
      : openApplePopup()

    if (!popup) {
      toast.error('Popup was blocked. Please allow popups for this site.')
      setOauthLoading(null)
      return
    }

    const checkClosed = setInterval(() => {
      if (popup.closed) {
        clearInterval(checkClosed)
        setOauthLoading(null)
      }
    }, 500)
  }

  const onSubmit = (data: SelfRegisterRequest) => {
    registerUser(data)
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 right-1/4 w-96 h-96 bg-neon-green/5 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 left-1/4 w-96 h-96 bg-neon-cyan/5 rounded-full blur-3xl" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
      >
        {/* Header */}
        <div className="text-center mb-8">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: 'spring', delay: 0.2 }}
            className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-neon-green to-neon-cyan flex items-center justify-center shadow-neon-green"
          >
            <Terminal className="w-10 h-10 text-cyber-black" />
          </motion.div>
          <h1 className="font-display font-bold text-3xl text-white tracking-wider">
            NEW USER
          </h1>
          <p className="mt-2 text-gray-400 font-mono text-sm">
            // CREATE YOUR SYSTEM ACCOUNT
          </p>
        </div>

        {/* Form */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="cyber-card p-8"
        >
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {/* Name */}
            <div className="space-y-2">
              <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                Display Name
              </label>
              <input
                type="text"
                {...register('name', {
                  required: 'Name is required',
                  minLength: {
                    value: 2,
                    message: 'Name must be at least 2 characters',
                  },
                })}
                className="cyber-input"
                placeholder="Agent Smith"
                autoComplete="name"
              />
              {errors.name && (
                <p className="text-sm text-neon-pink font-mono">{errors.name.message}</p>
              )}
            </div>

            {/* Email */}
            <div className="space-y-2">
              <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                Email Address
              </label>
              <input
                type="email"
                {...register('email', {
                  required: 'Email is required',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: 'Invalid email address',
                  },
                })}
                className="cyber-input"
                placeholder="user@system.net"
                autoComplete="email"
              />
              {errors.email && (
                <p className="text-sm text-neon-pink font-mono">{errors.email.message}</p>
              )}
            </div>

            {/* Password */}
            <div className="space-y-2">
              <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  {...register('password', {
                    required: 'Password is required',
                    minLength: {
                      value: 8,
                      message: 'Password must be at least 8 characters',
                    },
                  })}
                  className="cyber-input pr-12"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-neon-cyan transition-colors"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {errors.password && (
                <p className="text-sm text-neon-pink font-mono">{errors.password.message}</p>
              )}
              <p className="text-xs text-gray-500">Minimum 8 characters required</p>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isPending || !!oauthLoading}
              className="cyber-button-green w-full flex items-center justify-center gap-2"
            >
              {isPending ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  CREATING ACCOUNT...
                </>
              ) : (
                'CREATE ACCOUNT'
              )}
            </button>
          </form>

          {/* Divider */}
          <div className="my-6 flex items-center gap-4">
            <div className="flex-1 border-t border-cyber-border" />
            <span className="text-gray-500 text-sm font-mono">OR</span>
            <div className="flex-1 border-t border-cyber-border" />
          </div>

          {/* OAuth Buttons */}
          <div className="space-y-3">
            <OAuthButton
              provider="google"
              onClick={() => handleOAuthLogin('google')}
              isLoading={oauthLoading === 'google'}
            />
            <OAuthButton
              provider="facebook"
              onClick={() => handleOAuthLogin('facebook')}
              isLoading={oauthLoading === 'facebook'}
            />
            <OAuthButton
              provider="apple"
              onClick={() => handleOAuthLogin('apple')}
              isLoading={oauthLoading === 'apple'}
            />
          </div>

          {/* Login Link */}
          <div className="mt-6 text-center">
            <p className="text-gray-400 text-sm">
              Already have access?{' '}
              <Link
                to="/login"
                className="text-neon-cyan hover:underline font-medium"
              >
                Sign In
              </Link>
            </p>
          </div>
        </motion.div>

        {/* Footer */}
        <p className="mt-8 text-center text-gray-600 font-mono text-xs">
          API VISUAL TESTER v1.0.0 // SECURE CONNECTION
        </p>
      </motion.div>
    </div>
  )
}
