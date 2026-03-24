import { motion } from 'framer-motion'
import { LucideIcon } from 'lucide-react'
import { cn } from '../lib/utils'

interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  trend?: {
    value: number
    isPositive: boolean
  }
  color?: 'cyan' | 'pink' | 'green' | 'yellow'
}

const colorClasses = {
  cyan: {
    icon: 'from-neon-cyan/20 to-neon-cyan/5 border-neon-cyan/30',
    iconColor: 'text-neon-cyan',
    shadow: 'hover:shadow-neon-cyan',
  },
  pink: {
    icon: 'from-neon-pink/20 to-neon-pink/5 border-neon-pink/30',
    iconColor: 'text-neon-pink',
    shadow: 'hover:shadow-neon-pink',
  },
  green: {
    icon: 'from-neon-green/20 to-neon-green/5 border-neon-green/30',
    iconColor: 'text-neon-green',
    shadow: 'hover:shadow-neon-green',
  },
  yellow: {
    icon: 'from-neon-yellow/20 to-neon-yellow/5 border-neon-yellow/30',
    iconColor: 'text-neon-yellow',
    shadow: 'hover:shadow-[0_0_20px_rgba(245,255,0,0.3)]',
  },
}

export default function StatCard({ title, value, icon: Icon, trend, color = 'cyan' }: StatCardProps) {
  const colors = colorClasses[color]

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ y: -4 }}
      className={cn(
        'cyber-card p-6 transition-shadow duration-300',
        colors.shadow
      )}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-display uppercase tracking-wider text-gray-400">{title}</p>
          <p className="mt-2 text-3xl font-display font-bold text-white">{value}</p>
          {trend && (
            <p
              className={cn(
                'mt-2 text-sm font-mono',
                trend.isPositive ? 'text-neon-green' : 'text-neon-pink'
              )}
            >
              {trend.isPositive ? '+' : ''}{trend.value}%
            </p>
          )}
        </div>
        <div
          className={cn(
            'w-12 h-12 rounded-xl bg-gradient-to-br border flex items-center justify-center',
            colors.icon
          )}
        >
          <Icon className={cn('w-6 h-6', colors.iconColor)} />
        </div>
      </div>
    </motion.div>
  )
}
