import { useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Mail, RefreshCw, Loader2 } from 'lucide-react'
import { useAuthStore } from '../store/auth'
import { useResendVerification } from '../hooks/useSelfAuth'

export default function VerificationPendingPage() {
  const pendingEmail = useAuthStore((state) => state.pendingVerificationEmail)
  const [cooldown, setCooldown] = useState(0)
  const { mutate: resend, isPending } = useResendVerification()

  // Redirect if no pending email
  if (!pendingEmail) {
    return <Navigate to="/login" replace />
  }

  const handleResend = () => {
    if (cooldown > 0) return

    resend(
      { email: pendingEmail },
      {
        onSuccess: () => {
          setCooldown(60)
          const interval = setInterval(() => {
            setCooldown((prev) => {
              if (prev <= 1) {
                clearInterval(interval)
                return 0
              }
              return prev - 1
            })
          }, 1000)
        },
      }
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-neon-green/5 rounded-full blur-3xl" />
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
            className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-neon-green to-neon-cyan flex items-center justify-center shadow-neon-green"
          >
            <Mail className="w-10 h-10 text-cyber-black" />
          </motion.div>
          <h1 className="font-display font-bold text-3xl text-white tracking-wider">
            CHECK YOUR EMAIL
          </h1>
          <p className="mt-2 text-gray-400 font-mono text-sm">
            // VERIFICATION REQUIRED
          </p>
        </div>

        {/* Content */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="cyber-card p-8 text-center"
        >
          <p className="text-gray-300 mb-2">
            We sent a verification link to:
          </p>
          <p className="text-neon-cyan font-mono mb-6">{pendingEmail}</p>
          <p className="text-gray-400 text-sm mb-8">
            Click the link in the email to verify your account and complete registration.
          </p>

          <div className="space-y-4">
            <button
              type="button"
              onClick={handleResend}
              disabled={isPending || cooldown > 0}
              className="cyber-button w-full flex items-center justify-center gap-2"
            >
              {isPending ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  SENDING...
                </>
              ) : cooldown > 0 ? (
                <>
                  <RefreshCw className="w-5 h-5" />
                  RESEND IN {cooldown}s
                </>
              ) : (
                <>
                  <RefreshCw className="w-5 h-5" />
                  RESEND VERIFICATION EMAIL
                </>
              )}
            </button>

            <Link
              to="/login"
              className="block w-full py-3 px-4 text-center text-gray-400 hover:text-white transition-colors"
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
