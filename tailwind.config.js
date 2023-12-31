/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./business/templates/**/*.templ"],
  theme: {
    extend: {
      fontFamily: {
        'mono': ['JetBrains Mono', 'monospace'],
        'rubik': ['Rubik', 'monospace'],
      },
      colors: {
        go: "#00ACD7",
      }
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
};
