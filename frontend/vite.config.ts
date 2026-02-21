import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    UnoCSS(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json'],
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Vue 核心库
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
          // CodeMirror 相关
          'codemirror': [
            '@codemirror/state',
            '@codemirror/view',
            '@codemirror/commands',
            '@codemirror/search',
            '@codemirror/language',
            '@codemirror/lang-json',
            '@codemirror/lang-xml',
            '@codemirror/lang-html',
            '@codemirror/theme-one-dark',
            '@codemirror/highlight',
            '@lezer/common',
            '@lezer/highlight',
            '@lezer/lr',
          ],
          // 工具库
          'utils': ['@vueuse/core', 'crypto-js', 'js-yaml', 'mitt'],
          // 国际化
          'i18n': ['vue-i18n'],
          // 表格组件
          'table': ['@tanstack/table-core', '@tanstack/vue-table', '@tanstack/vue-virtual'],
        },
      },
    },
    // 调整 chunk 大小警告阈值 (已通过 manualChunks 优化，从 869KB 降至 638KB)
    chunkSizeWarningLimit: 700,
  },
})
