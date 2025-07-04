import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt().override('nuxt/javascript', {
  rules: {
    'sort-imports': 'error',
  },
})
