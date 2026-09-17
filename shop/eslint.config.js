import antfu from '@antfu/eslint-config'

export default antfu({
  formatters: true,
  stylistic: true,
  ignores: [
    'src/data/regions.json', // 省市区静态数据，体积大（~480KB）无需 lint
  ],
  rules: {
    'n/prefer-global/process': 'off',
    'no-undef': 'error',
    'no-fallthrough': 'off',
    'vue/block-order': 'off',
    'prefer-promise-reject-errors': 'off',
  },
})
