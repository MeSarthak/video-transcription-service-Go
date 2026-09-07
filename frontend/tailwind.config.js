import { heroui } from "@heroui/theme";

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        yt: {
          red: "#FF0000",
          hover: "#e60000",
          dark: "#0f0f0f",
          surface: "#181818",
          card: "#212121",
          hoverCard: "#272727",
          border: "#303030",
          borderLight: "#e5e5e5",
          subtext: "#aaaaaa",
        },
      },
      aspectRatio: {
        "16/9": "16 / 9",
      },
    },
  },
  darkMode: "class",
  plugins: [
    heroui({
      themes: {
        dark: {
          colors: {
            background: "#0f0f0f",
            foreground: "#f1f1f1",
            primary: {
              50: "#ffe5e5",
              100: "#ffb8b8",
              200: "#ff8a8a",
              300: "#ff5c5c",
              400: "#ff2e2e",
              500: "#FF0000",
              600: "#e60000",
              700: "#cc0000",
              800: "#990000",
              900: "#660000",
              DEFAULT: "#FF0000",
              foreground: "#ffffff",
            },
            content1: "#181818",
            content2: "#212121",
            content3: "#272727",
            content4: "#303030",
          },
        },
        light: {
          colors: {
            background: "#f9f9f9",
            foreground: "#0f0f0f",
            primary: {
              50: "#ffe5e5",
              100: "#ffb8b8",
              200: "#ff8a8a",
              300: "#ff5c5c",
              400: "#ff2e2e",
              500: "#FF0000",
              600: "#e60000",
              700: "#cc0000",
              800: "#990000",
              900: "#660000",
              DEFAULT: "#FF0000",
              foreground: "#ffffff",
            },
            content1: "#ffffff",
            content2: "#f2f2f2",
            content3: "#e5e5e5",
            content4: "#d9d9d9",
          },
        },
      },
    }),
  ],
};
