import tailwindcss from '@tailwindcss/vite'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  future: {
    compatibilityVersion: 4,
  },
  modules: [
    '@nuxt/icon',
    '@pinia/nuxt',
    '@vueuse/nuxt',
    '@vee-validate/nuxt',
  ],
  vite: {
    plugins: [tailwindcss()],
  },
  icon: {
    serverBundle: 'remote',
  },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    databasePath: './data/app.db',
    sessionSecret: '',
    public: {
      appUrl: '',
    },
  },
})
