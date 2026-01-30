import { useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useForm } from 'react-hook-form'
import { KeyRound, Eye, EyeOff, Loader2, XCircle } from 'lucide-react'
import { useResetPassword } from '../hooks/useSelfAuth'

interface ResetPasswordForm {
  new_password: string
  confirm_password: string
}

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const { mutate: resetPassword, isPending } = useResetPassword()

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ResetPasswordForm>()

  const password = watch('new_password')

  const onSubmit = (data: ResetPasswordForm) => {
    if (!token) return
    resetPassword({ token, new_password: data.new_password })
  }

  // Show error if no token
  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <div className="fixed inset-0 pointer-events-none">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-neon-pink/5 rounded-full blur-3xl" />
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="w-full max-w-md relative z-10"
        >
          <div className="cyber-card p-8 text-center">
            <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-pink/20 flex items-center justify-center">
              <XCircle className="w-8 h-8 text-neon-pink" />
            </div>
            <h1 className="font-display font-bold text-2xl text-white tracking-wider mb-4">
              INVALID LINK
            </h1>
            <p className="text-gray-400 mb-6">
              The reset link is missing or invalid. Please request a new password reset.
            </p>
            <Link to="/forgot-password" className="cyber-button inline-block">
              REQUEST NEW LINK
            </Link>
          </div>
        </motion.div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-neon-pink/5 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-neon-cyan/5 rounded-full blur-3xl" />
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
            className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-neon-pink to-neon-cyan flex items-center justify-center shadow-neon-pink"
          >
            <KeyRound className="w-10 h-10 text-cyber-black" />
          </motion.div>
          <h1 className="font-display font-bold text-3xl text-white tracking-wider">
            NEW PASSWORD
          </h1>
          <p className="mt-2 text-gray-400 font-mono text-sm">
            // CREATE A NEW SECURE PASSWORD
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
            {/* New Password */}
            <div className="space-y-2">
              <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                New Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  {...register('new_password', {
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
              {errors.new_password && (
                <p className="text-sm text-neon-pink font-mono">{errors.new_password.message}</p>
              )}
              <p className="text-xs text-gray-500">Minimum 8 characters required</p>
            </div>

            {/* Confirm Password */}
            <div className="space-y-2">
              <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  type={showConfirmPassword ? 'text' : 'password'}
                  {...register('confirm_password', {
                    required: 'Please confirm your password',
                    validate: (value) => value === password || 'Passwords do not match',
                  })}
                  className="cyber-input pr-12"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-neon-cyan transition-colors"
                >
                  {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {errors.confirm_password && (
                <p className="text-sm text-neon-pink font-mono">{errors.confirm_password.message}</p>
              )}
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isPending}
              className="cyber-button w-full flex items-center justify-center gap-2"
            >
              {isPending ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  RESETTING...
                </>
              ) : (
                'RESET PASSWORD'
              )}
            </button>
          </form>

          {/* Login Link */}
          <div className="mt-6 text-center">
            <Link
              to="/login"
              className="text-gray-400 hover:text-neon-cyan transition-colors text-sm"
            >
              Back to Login
            </Link>
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
