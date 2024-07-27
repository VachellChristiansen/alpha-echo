/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/*.{html,js}",
    "./views/components/*.{html,js}",
    "./views/components/opus/*.{html,js}",
    "./views/components/chrysus/*.{html,js}",
    "./views/components/vacuus/*.{html,js}",
    "./views/components/nuntius/*.{html,js}"
  ],
  theme: {
    extend: {
      fontFamily: {
        jsans: ["Josefin Sans", "sans-serif"],
        montserrat: ["Montserrat", "sans-serif"],
      },
      keyframes: {
        'fade-in-left': {
          '0%': { opacity: '0', transform: 'translateX(-5%)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        'fade-in-right': {
          '0%': { opacity: '0', transform: 'translateX(5%)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        'fade-in-top': {
          '0%': { opacity: '0', transform: 'translateY(-5%)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-in-bottom': {
          '0%': { opacity: '0', transform: 'translateY(5%)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-out-top': {
          '0%': { opacity: '1', transform: 'translateY(0)' },
          '100%': { opacity: '0', transform: 'translateY(-100px)' },
        },
      },
      animation: {
        'fade-in-left': 'fade-in-left 0.5s ease-out',
        'fade-in-right': 'fade-in-right 0.5s ease-out',
        'fade-in-top': 'fade-in-top 0.5s ease-out',
        'fade-in-bottom': 'fade-in-bottom 0.5s ease-out',
        'fade-out-top': 'fade-out-top 0.5s ease-out',
      },
    },
  },
  plugins: [],
}

