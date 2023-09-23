/* eslint-env node */
// require("@rushstack/eslint-patch/modern-module-resolution");

module.exports = {
  env: {
    node: true,
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 2020,
    sourceType: 'module',
  },
  extends: [
    'plugin:vue/vue3-recommended',
    '@vue/typescript/recommended',
    'eslint:recommended',
    'prettier',
  ],
  plugins: ['prettier'],
  rules: {
    // Required rules
    //////////////////////////////////////////////
    'prettier/prettier': 'error',
    'arrow-body-style': 0,
    'prefer-arrow-callback': 0,
    // "vue/order-in-components": "error",
    //////////////////////////////////////////////

    // Non-required rules
    // Vue rules
    'vue/no-reserved-component-names': 0,
    'vue/multi-word-component-names': 0,
    // @typescript-eslint rules
    '@typescript-eslint/ban-ts-comment': 0,
    '@typescript-eslint/no-empty-function': 0,
    // other rules
    'no-async-promise-executor': 0,
    'vue/no-v-html': 0,
  },
};
