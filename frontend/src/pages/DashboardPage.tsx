import { motion } from 'framer-motion'
import { LayoutDashboard, Users, Package, Globe, Building2, FileText, Shield, Activity } from 'lucide-react'
import { Link } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import StatCard from '../components/StatCard'
import { useUsers } from '../hooks/useUsers'
import { useItems } from '../hooks/useItems'
import { useCountries } from '../hooks/useCountries'
import { useCities } from '../hooks/useCities'
import { useDocuments } from '../hooks/useDocuments'
import { useRoles } from '../hooks/useRBAC'
import { useAuthStore } from '../store/auth'

const quickLinks = [
  { path: '/users', label: 'Users', icon: Users, color: 'cyan' },
  { path: '/items', label: 'Items', icon: Package, color: 'green' },
  { path: '/countries', label: 'Countries', icon: Globe, color: 'pink' },
  { path: '/cities', label: 'Cities', icon: Building2, color: 'yellow' },
  { path: '/documents', label: 'Documents', icon: FileText, color: 'cyan' },
  { path: '/rbac', label: 'RBAC', icon: Shield, color: 'pink' },
] as const

export default function DashboardPage() {
  const user = useAuthStore((state) => state.user)
  const { data: usersData } = useUsers({ page_size: 1 })
  const { data: itemsData } = useItems({ page_size: 1 })
  const { data: countriesData } = useCountries({ page_size: 1 })
  const { data: citiesData } = useCities({ page_size: 1 })
  const { data: documentsData } = useDocuments({ page_size: 1 })
  const { data: rolesData } = useRoles({ page_size: 1 })

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle={`Welcome back, ${user?.name || 'User'} // System Status: ONLINE`}
        icon={LayoutDashboard}
      />

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 mb-8">
        <StatCard
          title="Users"
          value={usersData?.total ?? '-'}
          icon={Users}
          color="cyan"
        />
        <StatCard
          title="Items"
          value={itemsData?.total ?? '-'}
          icon={Package}
          color="green"
        />
        <StatCard
          title="Countries"
          value={countriesData?.total ?? '-'}
          icon={Globe}
          color="pink"
        />
        <StatCard
          title="Cities"
          value={citiesData?.total ?? '-'}
          icon={Building2}
          color="yellow"
        />
        <StatCard
          title="Documents"
          value={documentsData?.total ?? '-'}
          icon={FileText}
          color="cyan"
        />
        <StatCard
          title="Roles"
          value={rolesData?.total ?? '-'}
          icon={Shield}
          color="pink"
        />
      </div>

      {/* Quick Links */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="cyber-card p-6"
      >
        <h2 className="font-display font-bold text-lg text-neon-cyan tracking-wider mb-6 flex items-center gap-2">
          <Activity className="w-5 h-5" />
          QUICK ACCESS
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
          {quickLinks.map((link, index) => {
            const Icon = link.icon
            return (
              <motion.div
                key={link.path}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 * index }}
              >
                <Link
                  to={link.path}
                  className="block p-4 rounded-xl bg-cyber-darker border border-cyber-border hover:border-neon-cyan/50 hover:shadow-neon-cyan transition-all duration-300 group text-center"
                >
                  <div className="w-12 h-12 mx-auto mb-3 rounded-xl bg-gradient-to-br from-neon-cyan/20 to-transparent border border-neon-cyan/30 flex items-center justify-center group-hover:scale-110 transition-transform">
                    <Icon className="w-6 h-6 text-neon-cyan" />
                  </div>
                  <p className="font-display font-medium text-sm text-gray-300 group-hover:text-white transition-colors">
                    {link.label}
                  </p>
                </Link>
              </motion.div>
            )
          })}
        </div>
      </motion.div>

      {/* System Info */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="mt-8 cyber-card p-6"
      >
        <h2 className="font-display font-bold text-lg text-neon-cyan tracking-wider mb-4">
          SYSTEM INFORMATION
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 font-mono text-sm">
          <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
            <p className="text-gray-500 mb-1">API VERSION</p>
            <p className="text-neon-green">v1.0.0</p>
          </div>
          <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
            <p className="text-gray-500 mb-1">BASE URL</p>
            <p className="text-neon-cyan truncate">localhost:3000/api/v1</p>
          </div>
          <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
            <p className="text-gray-500 mb-1">AUTH STATUS</p>
            <p className="text-neon-green">AUTHENTICATED</p>
          </div>
          <div className="p-4 bg-cyber-darker rounded-lg border border-cyber-border">
            <p className="text-gray-500 mb-1">USER ID</p>
            <p className="text-white">#{user?.id || '-'}</p>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
