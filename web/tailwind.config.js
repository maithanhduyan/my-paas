/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: '#0f0f0f',
          50: '#171717',
          100: '#1e1e1e',
          200: '#262626',
          300: '#2e2e2e',
        },
        accent: {
          DEFAULT: '#6366f1',
          hover: '#818cf8',
          muted: '#4f46e5',
        },
        success: '#22c55e',
        warning: '#eab308',
        danger: '#ef4444',
      },
    },
  },
  plugins: [],
}
