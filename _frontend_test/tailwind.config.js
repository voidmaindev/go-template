/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cyber: {
          black: '#0a0a0f',
          darker: '#0d0d14',
          dark: '#12121a',
          medium: '#1a1a24',
          light: '#24242f',
          border: '#2a2a3a',
        },
        neon: {
          cyan: '#00f5ff',
          pink: '#ff00aa',
          green: '#00ff88',
          yellow: '#f5ff00',
          orange: '#ff8800',
          purple: '#aa00ff',
        },
        terminal: {
          green: '#00ff41',
          amber: '#ffb000',
        }
      },
      fontFamily: {
        display: ['"Orbitron"', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
        body: ['"Rajdhani"', 'sans-serif'],
      },
      boxShadow: {
        'neon-cyan': '0 0 20px rgba(0, 245, 255, 0.3), 0 0 40px rgba(0, 245, 255, 0.1)',
        'neon-pink': '0 0 20px rgba(255, 0, 170, 0.3), 0 0 40px rgba(255, 0, 170, 0.1)',
        'neon-green': '0 0 20px rgba(0, 255, 136, 0.3), 0 0 40px rgba(0, 255, 136, 0.1)',
        'inner-glow': 'inset 0 0 20px rgba(0, 245, 255, 0.05)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'scan': 'scan 8s linear infinite',
        'flicker': 'flicker 0.15s infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
      },
      keyframes: {
        scan: {
          '0%': { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(100vh)' },
        },
        flicker: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.8' },
        },
        glow: {
          '0%': { boxShadow: '0 0 5px rgba(0, 245, 255, 0.2), 0 0 10px rgba(0, 245, 255, 0.1)' },
          '100%': { boxShadow: '0 0 20px rgba(0, 245, 255, 0.4), 0 0 40px rgba(0, 245, 255, 0.2)' },
        },
      },
      backgroundImage: {
        'grid-pattern': 'linear-gradient(rgba(0, 245, 255, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 245, 255, 0.03) 1px, transparent 1px)',
        'gradient-radial': 'radial-gradient(ellipse at center, var(--tw-gradient-stops))',
      },
      backgroundSize: {
        'grid': '50px 50px',
      },
    },
  },
  plugins: [],
}
