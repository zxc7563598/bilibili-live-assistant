import { reactive } from 'vue'

// 共享状态：AppToast.vue 渲染它，toast.* 写入它
export const toasts = reactive([])

let seed = 0

// 移除一条 toast（splice 后触发 TransitionGroup 离开动画）
export function close(id) {
  const i = toasts.findIndex(t => t.id === id)
  if (i !== -1)
    toasts.splice(i, 1)
}

function show(message, type, options = {}) {
  const { position = 'top', duration = 2000, closable = false } = options
  const id = ++seed
  toasts.push({ id, message, type, position, duration, closable })

  // duration 为 0 表示常驻，需手动关闭
  if (duration > 0)
    setTimeout(close, duration, id)

  // 返回手动关闭句柄
  return () => close(id)
}

const toast = {
  success: (message, options) => show(message, 'success', options),
  fail: (message, options) => show(message, 'fail', options),
  error: (message, options) => show(message, 'fail', options),
  warning: (message, options) => show(message, 'warning', options),
  info: (message, options) => show(message, 'info', options),
}

export default toast
