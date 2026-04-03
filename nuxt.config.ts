// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxt/ui',
    // '@nuxt/content',
    '@nuxt/eslint',
    'nuxt-security',
    '@nuxtjs/i18n',
  ],
  devtools: { enabled: true },
  css: ['~/assets/css/main.css'],
  future: {
    compatibilityVersion: 4,
  },
  compatibilityDate: '2026-01-01',
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
