import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
    './providers/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-sans)',  'Space Grotesk', 'sans-serif'],
        mono: ['var(--font-mono)',  'IBM Plex Mono', 'monospace'],
      },
      boxShadow: {
        card:           'var(--shadow-card)',
        floating:       'var(--shadow-floating)',
        pressed:        'var(--shadow-pressed)',
        recessed:       'var(--shadow-recessed)',
        glow:           '0 0 12px 3px rgba(255,71,87,0.55)',
        'glow-green':   '0 0 12px 3px rgba(34,197,94,0.55)',
        'glow-yellow':  '0 0 12px 3px rgba(250,204,21,0.55)',
        accent:         '4px 4px 10px rgba(160,28,42,0.6), -2px -2px 8px rgba(255,90,100,0.22)',
        'accent-pressed':'inset 4px 4px 8px rgba(160,28,42,0.55), inset -2px -2px 7px rgba(255,90,100,0.18)',
      },
      keyframes: {
        'blink-red': {
          '0%,100%': { boxShadow: '0 0 8px 2px rgba(255,71,87,0.85)' },
          '50%':     { boxShadow: '0 0 2px 1px rgba(255,71,87,0.25)' },
        },
        'blink-green': {
          '0%,100%': { boxShadow: '0 0 8px 2px rgba(34,197,94,0.85)' },
          '50%':     { boxShadow: '0 0 2px 1px rgba(34,197,94,0.25)' },
        },
        'blink-yellow': {
          '0%,100%': { boxShadow: '0 0 8px 2px rgba(250,204,21,0.85)' },
          '50%':     { boxShadow: '0 0 2px 1px rgba(250,204,21,0.25)' },
        },
      },
      animation: {
        'led-red':    'blink-red    2s ease-in-out infinite',
        'led-green':  'blink-green  2s ease-in-out infinite',
        'led-yellow': 'blink-yellow 2s ease-in-out infinite',
      },
    },
  },
  plugins: [],
}

export default config
