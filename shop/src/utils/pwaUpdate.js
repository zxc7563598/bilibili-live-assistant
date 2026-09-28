// PWA 更新：注册 Service Worker 并主动检查新版本
import { registerSW } from 'virtual:pwa-register'

const PERIOD = 60 * 60 * 1000

// 页面加载时已被 SW 接管，才有"被新版本接管"这回事
const wasControlled = 'serviceWorker' in navigator && !!navigator.serviceWorker.controller

let reloading = false

// 整页刷新闸门：activated 与 controllerchange 会前后脚到达，避免刷新两次
function reloadOnce() {
  if (reloading)
    return
  reloading = true
  window.location.reload()
}

// clientsClaim 会让新 SW 接管所有标签页，非触发更新的标签页只能靠这个事件兜底
if (wasControlled) {
  navigator.serviceWorker.addEventListener('controllerchange', reloadOnce)
}

// 定时巡检 + 回到前台时补查一次
function armUpdateCheck(swUrl, r) {
  setInterval(async () => {
    if ('onLine' in navigator && !navigator.onLine)
      return
    const resp = await fetch(swUrl, {
      cache: 'no-store',
      headers: { 'cache': 'no-store', 'cache-control': 'no-cache' },
    })
    if (resp?.status === 200)
      await r.update()
  }, PERIOD)

  // SPA 内部跳转与从后台恢复都不触发浏览器自带的更新检查
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible')
      r.update().catch(() => {})
  })
}

// 新版本 activate 后整页刷新，这里只负责注册与主动检查
export function registerPwaUpdate() {
  registerSW({
    immediate: true,
    onNeedReload: reloadOnce,
    onRegisteredSW(swUrl, r) {
      if (!r)
        return
      if (r.active?.state === 'activated') {
        armUpdateCheck(swUrl, r)
      }
      else if (r.installing) {
        r.installing.addEventListener('statechange', (e) => {
          if (e.target.state === 'activated')
            armUpdateCheck(swUrl, r)
        })
      }
    },
  })
}
