import antfu from '@antfu/eslint-config'

export default antfu({
  formatters: true,
  stylistic: true,
  rules: {
    'n/prefer-global/process': 'off',
    'no-undef': 'error',
    'no-fallthrough': 'off',
    'vue/block-order': 'off',
    'prefer-promise-reject-errors': 'off',
  },
})
