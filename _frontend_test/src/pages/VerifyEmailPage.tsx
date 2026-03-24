import { useEffect } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { useVerifyEmail } from '../hooks/useSelfAuth'

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const { mutate: verify, isPending, isSuccess, isError, error } = useVerifyEmail()

  useEffect(() => {
    if (token) {
      verify({ token })
    }
  }, [token, verify])

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-neon-cyan/5 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-neon-green/5 rounded-full blur-3xl" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
      >
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="cyber-card p-8 text-center"
        >
          {!token ? (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-pink/20 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-neon-pink" />
              </div>
              <h1 className="font-display font-bold text-2xl text-white tracking-wider mb-4">
                INVALID LINK
              </h1>
              <p className="text-gray-400 mb-6">
                The verification link is missing or invalid. Please check your email for the correct link.
              </p>
              <Link to="/login" className="cyber-button inline-block">
                GO TO LOGIN
              </Link>
            </>
          ) : isPending ? (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-cyan/20 flex items-center justify-center">
                <Loader2 className="w-8 h-8 text-neon-cyan animate-spin" />
              </div>
              <h1 className="font-display font-bold text-2xl text-white tracking-wider mb-4">
                VERIFYING...
              </h1>
              <p className="text-gray-400">
                Please wait while we verify your email address.
              </p>
            </>
          ) : isSuccess ? (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-green/20 flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-neon-green" />
              </div>
              <h1 className="font-display font-bold text-2xl text-white tracking-wider mb-4">
                EMAIL VERIFIED
              </h1>
              <p className="text-gray-400 mb-6">
                Your email has been verified. Redirecting to login...
              </p>
            </>
          ) : isError ? (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-pink/20 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-neon-pink" />
              </div>
              <h1 className="font-display font-bold text-2xl text-white tracking-wider mb-4">
                VERIFICATION FAILED
              </h1>
              <p className="text-gray-400 mb-6">
                {(error as Error)?.message || 'The verification link may have expired or is invalid.'}
              </p>
              <Link to="/login" className="cyber-button inline-block">
                GO TO LOGIN
              </Link>
            </>
          ) : null}
        </motion.div>

        {/* Footer */}
        <p className="mt-8 text-center text-gray-600 font-mono text-xs">
          API VISUAL TESTER v1.0.0 // SECURE CONNECTION
        </p>
      </motion.div>
    </div>
  )
}
