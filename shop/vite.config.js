import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

const DEV_PROXY_TARGET = 'http://localhost:25443'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const viteEnv = loadEnv(mode, process.cwd())
  const { VITE_API_URL } = viteEnv
  const proxyTarget = VITE_API_URL || DEV_PROXY_TARGET

  return {
    base: '/shop/',
    server: {
      host: '0.0.0.0',
      proxy: {
        // 接口请求代理到后端。注意不能像 web 那样 rewrite 掉 /api：
        // 后端 shop 接口就注册在 /api/shop 下，前端路径与后端路由完全一致。
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
        // 上传文件（pkg/fileutil.SaveUploadedFile 落盘到后端 uploads/ 目录），开发环境代理到后端以便本地预览
        '/uploads': {
          target: proxyTarget,
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
