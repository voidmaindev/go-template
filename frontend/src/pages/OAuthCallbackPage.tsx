import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Loader2, CheckCircle, XCircle } from 'lucide-react'
import { useAuthStore } from '../store/auth'
import type { OAuthProvider } from '../types'

type CallbackState = 'loading' | 'success' | 'error'

export default function OAuthCallbackPage() {
  const [searchParams] = useSearchParams()
  const [state, setState] = useState<CallbackState>('loading')
  const [errorMessage, setErrorMessage] = useState<string>('')
  const setAuth = useAuthStore((state) => state.setAuth)

  useEffect(() => {
    // Check for auth data in URL hash (backend redirect flow)
    const hash = window.location.hash
    if (hash.startsWith('#auth=')) {
      try {
        const authJSON = decodeURIComponent(hash.slice(6)) // Remove '#auth='
        const authData = JSON.parse(authJSON)

        // Store auth data in Zustand store
        setAuth({
          accessToken: authData.access_token,
          refreshToken: authData.refresh_token,
          expiresAt: authData.expires_at,
          user: authData.user,
        })

        setState('success')

        // Notify opener and close popup
        if (window.opener) {
          try {
            window.opener.postMessage(
              { type: 'oauth-success', data: authData },
              window.location.origin
            )
          } catch {
            // COOP may block postMessage, but auth is already in localStorage
          }
        }

        // Close popup after brief delay
        setTimeout(() => window.close(), 1500)
        return
      } catch (err) {
        setState('error')
        setErrorMessage('Failed to parse authentication data')
        return
      }
    }

    const error = searchParams.get('error')
    const errorDescription = searchParams.get('error_description')
    const isLinking = searchParams.get('link') === 'true'
    const provider = searchParams.get('provider')

    // Handle OAuth error
    if (error) {
      setState('error')
      setErrorMessage(errorDescription || error || 'Authentication failed')

      // Notify opener if exists
      if (window.opener) {
        window.opener.postMessage(
          { type: 'oauth-error', error: errorDescription || error },
          window.location.origin
        )
      }
      return
    }

    // If linking, just send success message to opener
    if (isLinking && provider) {
      setState('success')
      if (window.opener) {
        window.opener.postMessage(
          { type: 'link-success', provider },
          window.location.origin
        )
        setTimeout(() => window.close(), 1500)
      }
      return
    }

    // If no hash auth data and no error, something went wrong
    setState('error')
    setErrorMessage('Missing authentication data')
  }, [searchParams, setAuth])

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-cyber-black">
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-neon-cyan/5 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-neon-green/5 rounded-full blur-3xl" />
      </div>

      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="w-full max-w-sm relative z-10"
      >
        <div className="cyber-card p-8 text-center">
          {state === 'loading' && (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-cyan/20 flex items-center justify-center">
                <Loader2 className="w-8 h-8 text-neon-cyan animate-spin" />
              </div>
              <h1 className="font-display font-bold text-xl text-white tracking-wider mb-2">
                AUTHENTICATING
              </h1>
              <p className="text-gray-400 text-sm">
                Please wait...
              </p>
            </>
          )}

          {state === 'success' && (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-green/20 flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-neon-green" />
              </div>
              <h1 className="font-display font-bold text-xl text-white tracking-wider mb-2">
                SUCCESS
              </h1>
              <p className="text-gray-400 text-sm">
                This window will close automatically...
              </p>
            </>
          )}

          {state === 'error' && (
            <>
              <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-neon-pink/20 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-neon-pink" />
              </div>
              <h1 className="font-display font-bold text-xl text-white tracking-wider mb-2">
                AUTHENTICATION FAILED
              </h1>
              <p className="text-gray-400 text-sm mb-4">
                {errorMessage}
              </p>
              <button
                type="button"
                onClick={() => window.close()}
                className="cyber-button-pink"
              >
                CLOSE WINDOW
              </button>
            </>
          )}
        </div>
      </motion.div>
    </div>
  )
}
