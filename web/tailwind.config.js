/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  corePlugins: {
    preflight: false,
  },
  theme: {
    extend: {
      colors: {
        paper: '#f7f3eb',
        ink: '#20231f',
        ember: '#e65b2f',
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'sans-serif'],
        serif: ['"Noto Serif SC"', '"Songti SC"', 'SimSun', 'Georgia', 'serif'],
      },
    },
  },
  plugins: [],
};
