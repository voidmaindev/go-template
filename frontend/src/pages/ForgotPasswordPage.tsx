import { useState } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useForm } from 'react-hook-form'
import { KeyRound, Mail, Loader2, CheckCircle } from 'lucide-react'
import { useForgotPassword } from '../hooks/useSelfAuth'
import type { ForgotPasswordRequest } from '../types'

export default function ForgotPasswordPage() {
  const [emailSent, setEmailSent] = useState(false)
  const { mutate: forgotPassword, isPending } = useForgotPassword()

  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm<ForgotPasswordRequest>()

  const onSubmit = (data: ForgotPasswordRequest) => {
    forgotPassword(data, {
      onSuccess: () => setEmailSent(true),
    })
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
            RESET PASSWORD
          </h1>
          <p className="mt-2 text-gray-400 font-mono text-sm">
            // RECOVER YOUR ACCOUNT
          </p>
        </div>

        {/* Content */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="cyber-card p-8"
        >
          {emailSent ? (
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-green/20 flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-neon-green" />
              </div>
              <h2 className="font-display font-bold text-xl text-white mb-4">
                CHECK YOUR EMAIL
              </h2>
              <p className="text-gray-400 mb-2">
                We sent a password reset link to:
              </p>
              <p className="text-neon-cyan font-mono mb-6">{getValues('email')}</p>
              <p className="text-gray-500 text-sm mb-6">
                Click the link in the email to reset your password.
              </p>
              <Link
                to="/login"
                className="block w-full py-3 px-4 text-center text-gray-400 hover:text-white transition-colors"
              >
                Back to Login
              </Link>
            </div>
          ) : (
            <>
              <p className="text-gray-400 text-center mb-6">
                Enter your email address and we'll send you a link to reset your password.
              </p>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                {/* Email */}
                <div className="space-y-2">
                  <label className="block text-sm font-display uppercase tracking-wider text-gray-400">
                    Email Address
                  </label>
                  <div className="relative">
                    <input
                      type="email"
                      {...register('email', {
                        required: 'Email is required',
                        pattern: {
                          value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                          message: 'Invalid email address',
                        },
                      })}
                      className="cyber-input pl-12"
                      placeholder="user@system.net"
                      autoComplete="email"
                    />
                    <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
                  </div>
                  {errors.email && (
                    <p className="text-sm text-neon-pink font-mono">{errors.email.message}</p>
                  )}
                </div>

                {/* Submit */}
                <button
                  type="submit"
                  disabled={isPending}
                  className="cyber-button-pink w-full flex items-center justify-center gap-2"
                >
                  {isPending ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      SENDING...
                    </>
                  ) : (
                    'SEND RESET LINK'
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
            </>
          )}
        </motion.div>

        {/* Footer */}
        <p className="mt-8 text-center text-gray-600 font-mono text-xs">
          API VISUAL TESTER v1.0.0 // SECURE CONNECTION
        </p>
      </motion.div>
    </div>
  )
}
