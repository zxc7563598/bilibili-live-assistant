import { reactive } from 'vue'

// 共享状态：AppDialog.vue 渲染它，dialog.* 写入它
// 单例设计：同一时刻只展示一个 Dialog（window.confirm 本来就是单例阻塞式，忠实替换）
// 重复 open() 会原地覆盖内容，不做队列/堆叠
export const state = reactive({
  visible: false,
  type: 'info', // success | fail | warning | info | confirm
  title: '',
  message: '',
  showConfirm: true,
  showCancel: false,
  confirmText: '知道了',
  cancelText: '取消',
  confirmVariant: 'primary', // primary | danger | ghost | outline | soft
  maskClosable: false, // 默认关闭：防止误触遮罩关掉危险操作
  loading: false, // onConfirm 返回 Promise 时置 true，驱动按钮 loading
})

let onConfirm = null
let onCancel = null

// 核心入口：合并默认值、重置 loading、挂载回调
export function open(options = {}) {
  Object.assign(state, {
    visible: true,
    type: options.type || 'info',
    title: options.title || '',
    message: options.message || '',
    showConfirm: options.showConfirm ?? true,
    showCancel: options.showCancel ?? false,
    confirmText: options.confirmText || '知道了',
    cancelText: options.cancelText || '取消',
    confirmVariant: options.confirmVariant || 'primary',
    maskClosable: options.maskClosable ?? false,
    loading: false,
  })
  onConfirm = typeof options.onConfirm === 'function' ? options.onConfirm : null
  onCancel = typeof options.onCancel === 'function' ? options.onCancel : null
  // 返回手动关闭句柄
  return close
}

export function close() {
  state.visible = false
  state.loading = false
  onConfirm = null
  onCancel = null
}

// AppDialog 的「确认」按钮绑定
// 契约：同步返回 → 立即关闭；返回 Promise → 按钮 loading 直到 settle 再关闭
// reject 时不关闭：保持弹窗，等调用方自行 toast.error 处理，失败不静默关闭
export function confirmClick() {
  if (state.loading)
    return
  if (typeof onConfirm !== 'function') {
    close()
    return
  }
  const r = onConfirm()
  if (r && typeof r.then === 'function') {
    state.loading = true
    Promise.resolve(r)
      .then(close)
      .catch(() => {
        state.loading = false
      })
  }
  else {
    close()
  }
}

// AppDialog 的「取消」按钮 / 遮罩(当 maskClosable) 绑定
export function cancelClick() {
  if (state.loading)
    return
  if (typeof onCancel === 'function')
    onCancel()
  close()
}

// 每种类型带默认文案/按钮，调用方显式 options 覆盖
const TYPE_DEFAULTS = {
  success: { type: 'success', confirmText: '知道了', showCancel: false },
  fail: { type: 'fail', confirmText: '知道了', showCancel: false },
  warning: { type: 'warning', confirmText: '知道了', showCancel: false },
  info: { type: 'info', confirmText: '知道了', showCancel: false },
  confirm: { type: 'confirm', confirmText: '确定', cancelText: '取消', showCancel: true },
}

function show(type, options = {}) {
  return open({ ...TYPE_DEFAULTS[type], ...options })
}

const dialog = {
  success: o => show('success', o),
  fail: o => show('fail', o),
  error: o => show('fail', o), // 与 toast.error 对齐的别名
  warning: o => show('warning', o),
  info: o => show('info', o),
  confirm: o => show('confirm', o),
  close,
}

export default dialog
