/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        obsidian: {
          base: 'rgb(var(--color-base) / <alpha-value>)',
          surface: 'rgb(var(--color-surface) / <alpha-value>)',
          raised: 'rgb(var(--color-raised) / <alpha-value>)',
          highlight: 'rgb(var(--color-highlight) / <alpha-value>)',
          border: {
            dim: 'rgb(var(--color-border-dim) / <alpha-value>)',
            med: 'rgb(var(--color-border-med) / <alpha-value>)',
            bright: 'rgb(var(--color-border-bright) / <alpha-value>)',
          },
          text: {
            primary: 'rgb(var(--color-text-primary) / <alpha-value>)',
            secondary: 'rgb(var(--color-text-secondary) / <alpha-value>)',
            tertiary: 'rgb(var(--color-text-tertiary) / <alpha-value>)',
            dim: 'rgb(var(--color-text-dim) / <alpha-value>)',
          },
          accent: 'rgb(var(--color-accent) / <alpha-value>)',
          'accent-dim': 'rgb(var(--color-accent-dim) / <alpha-value>)',
          positive: 'rgb(var(--color-positive) / <alpha-value>)',
          'positive-dim': 'rgb(var(--color-positive-dim) / <alpha-value>)',
          negative: 'rgb(var(--color-negative) / <alpha-value>)',
          'negative-dim': 'rgb(var(--color-negative-dim) / <alpha-value>)',
          warning: 'rgb(var(--color-warning) / <alpha-value>)',
          info: 'rgb(var(--color-info) / <alpha-value>)',
          cyan: 'rgb(var(--color-cyan) / <alpha-value>)',
        },
      },
      fontFamily: {
        mono: [
          '"SF Mono"',
          'Menlo',
          'Monaco',
          '"Cascadia Mono"',
          'Consolas',
          '"DejaVu Sans Mono"',
          '"Liberation Mono"',
          '"Source Code Pro"',
          '"Courier New"',
          'monospace',
        ],
      },
      borderRadius: {
        sm: '2px',
        DEFAULT: '4px',
        md: '6px',
        lg: '8px',
      },
    },
  },
  plugins: [],
}
