import { useState } from 'react'
import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  LayoutDashboard,
  Users,
  Package,
  Globe,
  Building2,
  FileText,
  Shield,
  User,
  LogOut,
  Menu,
  X,
  ChevronRight,
  Terminal,
} from 'lucide-react'
import { useLogout } from '../hooks/useAuth'
import { useAuthStore } from '../store/auth'
import { cn } from '../lib/utils'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/users', label: 'Users', icon: Users },
  { path: '/items', label: 'Items', icon: Package },
  { path: '/countries', label: 'Countries', icon: Globe },
  { path: '/cities', label: 'Cities', icon: Building2 },
  { path: '/documents', label: 'Documents', icon: FileText },
  { path: '/rbac', label: 'RBAC', icon: Shield },
]

export default function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const location = useLocation()
  const { mutate: logout, isPending: isLoggingOut } = useLogout()
  const user = useAuthStore((state) => state.user)

  return (
    <div className="min-h-screen flex">
      {/* Desktop Sidebar */}
      <motion.aside
        initial={false}
        animate={{ width: sidebarOpen ? 280 : 80 }}
        className={cn(
          'hidden lg:flex flex-col bg-cyber-darker border-r border-cyber-border',
          'fixed left-0 top-0 h-screen z-40'
        )}
      >
        {/* Logo */}
        <div className="p-4 border-b border-cyber-border">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-neon-cyan to-neon-green flex items-center justify-center">
              <Terminal className="w-6 h-6 text-cyber-black" />
            </div>
            <AnimatePresence>
              {sidebarOpen && (
                <motion.div
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -10 }}
                  className="overflow-hidden"
                >
                  <h1 className="font-display font-bold text-lg text-neon-cyan tracking-wider">
                    API TESTER
                  </h1>
                  <p className="text-xs text-gray-500 font-mono">v1.0.0 // ONLINE</p>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>

        {/* Toggle */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="absolute -right-3 top-20 w-6 h-6 bg-cyber-dark border border-cyber-border rounded-full flex items-center justify-center hover:border-neon-cyan transition-colors"
        >
          <ChevronRight
            className={cn(
              'w-4 h-4 text-gray-400 transition-transform',
              sidebarOpen && 'rotate-180'
            )}
          />
        </button>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={cn(
                  'nav-item',
                  isActive && 'active',
                  !sidebarOpen && 'justify-center px-3'
                )}
              >
                <Icon className="w-5 h-5 flex-shrink-0" />
                <AnimatePresence>
                  {sidebarOpen && (
                    <motion.span
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: -10 }}
                      className="font-medium"
                    >
                      {item.label}
                    </motion.span>
                  )}
                </AnimatePresence>
              </NavLink>
            )
          })}
        </nav>

        {/* User Section */}
        <div className="p-4 border-t border-cyber-border space-y-2">
          <NavLink
            to="/profile"
            className={cn(
              'nav-item',
              location.pathname === '/profile' && 'active',
              !sidebarOpen && 'justify-center px-3'
            )}
          >
            <User className="w-5 h-5 flex-shrink-0" />
            <AnimatePresence>
              {sidebarOpen && (
                <motion.div
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -10 }}
                  className="flex-1 min-w-0"
                >
                  <p className="font-medium truncate">{user?.name || 'User'}</p>
                  <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                </motion.div>
              )}
            </AnimatePresence>
          </NavLink>
          <button
            onClick={() => logout()}
            disabled={isLoggingOut}
            className={cn(
              'nav-item w-full text-neon-pink hover:bg-neon-pink/10',
              !sidebarOpen && 'justify-center px-3'
            )}
          >
            <LogOut className="w-5 h-5 flex-shrink-0" />
            <AnimatePresence>
              {sidebarOpen && (
                <motion.span
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -10 }}
                  className="font-medium"
                >
                  {isLoggingOut ? 'Terminating...' : 'Logout'}
                </motion.span>
              )}
            </AnimatePresence>
          </button>
        </div>
      </motion.aside>

      {/* Mobile Header */}
      <div className="lg:hidden fixed top-0 left-0 right-0 h-16 bg-cyber-darker border-b border-cyber-border z-40 flex items-center px-4">
        <button
          onClick={() => setMobileMenuOpen(true)}
          className="p-2 hover:bg-cyber-light rounded-lg"
        >
          <Menu className="w-6 h-6 text-gray-400" />
        </button>
        <div className="flex items-center gap-3 ml-4">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-neon-cyan to-neon-green flex items-center justify-center">
            <Terminal className="w-5 h-5 text-cyber-black" />
          </div>
          <h1 className="font-display font-bold text-neon-cyan tracking-wider">API TESTER</h1>
        </div>
      </div>

      {/* Mobile Menu */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="lg:hidden fixed inset-0 bg-black/60 z-40"
              onClick={() => setMobileMenuOpen(false)}
            />
            <motion.aside
              initial={{ x: -280 }}
              animate={{ x: 0 }}
              exit={{ x: -280 }}
              className="lg:hidden fixed left-0 top-0 h-screen w-72 bg-cyber-darker border-r border-cyber-border z-50"
            >
              <div className="p-4 border-b border-cyber-border flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-neon-cyan to-neon-green flex items-center justify-center">
                    <Terminal className="w-6 h-6 text-cyber-black" />
                  </div>
                  <div>
                    <h1 className="font-display font-bold text-lg text-neon-cyan tracking-wider">
                      API TESTER
                    </h1>
                    <p className="text-xs text-gray-500 font-mono">v1.0.0 // ONLINE</p>
                  </div>
                </div>
                <button
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 hover:bg-cyber-light rounded-lg"
                >
                  <X className="w-5 h-5 text-gray-400" />
                </button>
              </div>
              <nav className="p-4 space-y-1">
                {navItems.map((item) => {
                  const Icon = item.icon
                  const isActive = location.pathname === item.path
                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      onClick={() => setMobileMenuOpen(false)}
                      className={cn('nav-item', isActive && 'active')}
                    >
                      <Icon className="w-5 h-5" />
                      <span className="font-medium">{item.label}</span>
                    </NavLink>
                  )
                })}
              </nav>
              <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-cyber-border space-y-2">
                <NavLink
                  to="/profile"
                  onClick={() => setMobileMenuOpen(false)}
                  className={cn('nav-item', location.pathname === '/profile' && 'active')}
                >
                  <User className="w-5 h-5" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{user?.name || 'User'}</p>
                    <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                  </div>
                </NavLink>
                <button
                  onClick={() => logout()}
                  disabled={isLoggingOut}
                  className="nav-item w-full text-neon-pink hover:bg-neon-pink/10"
                >
                  <LogOut className="w-5 h-5" />
                  <span className="font-medium">{isLoggingOut ? 'Terminating...' : 'Logout'}</span>
                </button>
              </div>
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <main
        className={cn(
          'flex-1 transition-all duration-300',
          'lg:ml-[280px]',
          !sidebarOpen && 'lg:ml-20',
          'pt-16 lg:pt-0'
        )}
      >
        <div className="p-4 lg:p-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
