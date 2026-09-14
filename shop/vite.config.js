import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const viteEnv = loadEnv(mode, process.cwd())
  const { VITE_API_URL } = viteEnv

  return {
    base: '/shop/',
    server: {
      host: '0.0.0.0',
      // 上传文件（pkg/fileutil.SaveUploadedFile 落盘到后端 uploads/ 目录），开发环境代理到后端以便本地预览
      proxy: {
        '/uploads': {
          target: VITE_API_URL,
          changeOrigin: true,
          secure: false,
        },
      },
    },
    plugins: [
      vue(),
      tailwindcss(),
      VitePWA({
        registerType: 'autoUpdate',
        injectRegister: false,
        pwaAssets: false,
        manifest: false,
        workbox: {
          globPatterns: ['**/*.{js,css,html}'],
          cleanupOutdatedCaches: true,
          clientsClaim: true,
        },
        devOptions: {
          enabled: false,
          navigateFallback: 'index.html',
          suppressWarnings: true,
          type: 'module',
        },
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  }
})
