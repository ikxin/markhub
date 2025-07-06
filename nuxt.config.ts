// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2026-01-01',
  future: {
    compatibilityVersion: 4,
  },
  devtools: { enabled: true },
  modules: [
    '@nuxt/ui',
    '@nuxt/content',
    '@nuxt/eslint',
    'nuxt-security',
    '@nuxtjs/i18n',
  ],
  css: ['~/assets/css/main.css'],
  eslint: {
    config: {
      nuxt: {
        sortConfigKeys: true,
      },
    },
  },
  fonts: {
    provider: 'local',
    experimental: {
      processCSSVariables: true,
    },
  },
  security: {
    corsHandler: {
      origin: '*',
    },
  },
})
