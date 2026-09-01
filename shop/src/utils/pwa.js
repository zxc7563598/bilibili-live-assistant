// PWA / 站点配置：名称、图标、主题色等从后台动态读取。
// 后台接口：GET /api/shop/manifest，返回 { code, data, msg }，data 为站点配置
import { ref } from 'vue'
import request from '@/static/request'

// 站点配置，响应式，供 router 等模块使用
export const siteConfig = ref({
  name: '积分商城',
  short_name: '积分商城',
  description: '积分商城',
  theme_color: '#ffffff',
  background_color: '#ffffff',
  favicon: '',
  apple_touch_icon: '',
  start_url: '/shop/',
  scope: '/shop/',
  display: 'standalone',
  icons: [],
})

// 加载后台站点配置
export async function loadSiteConfig() {
  try {
    const res = await request.get('/api/shop/manifest')
    const config = res?.data || {}
    siteConfig.value = { ...siteConfig.value, ...config }
  }
  catch (e) {
    // 接口未就绪或请求失败时保留默认值，不影响页面使用
    console.warn('[pwa] 加载站点配置失败，使用默认值', e)
  }
  return siteConfig.value
}

// 根据配置动态设置 head 中的内容（title / favicon / apple-touch-icon / manifest）。
// 注意：meta theme-color 由 utils/theme.js 按亮暗主题接管，这里不再覆盖。
export function applySiteConfig(config) {
  const cfg = config || siteConfig.value
  if (cfg.name) {
    document.title = cfg.name
  }
  setLink('icon', cfg.favicon)
  if (cfg.apple_touch_icon) {
    setLink('apple-touch-icon', cfg.apple_touch_icon)
  }
  applyManifest(cfg)
}

// 用配置生成 manifest，以 data URL 注入到 <link rel="manifest">
function applyManifest(cfg) {
  const manifest = {
    name: cfg.name,
    short_name: cfg.short_name,
    description: cfg.description,
    theme_color: cfg.theme_color,
    background_color: cfg.background_color,
    start_url: cfg.start_url || '/shop/',
    scope: cfg.scope || '/shop/',
    display: cfg.display || 'standalone',
    icons: cfg.icons || [],
  }
  const url = `data:application/manifest+json,${encodeURIComponent(JSON.stringify(manifest))}`
  setLink('manifest', url)
}

function setLink(rel, href) {
  if (!href)
    return
  let el = document.querySelector(`link[rel="${rel}"]`)
  if (!el) {
    el = document.createElement('link')
    el.setAttribute('rel', rel)
    document.head.appendChild(el)
  }
  el.setAttribute('href', href)
}
