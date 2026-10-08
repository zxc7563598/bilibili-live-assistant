// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { useAuthStore } from '@/store'

let isConfirming = false

export function handleAuthExpired(content, needTip) {
  if (isConfirming || !needTip)
    return
  isConfirming = true
  $dialog.confirm({
    title: '提示',
    type: 'info',
    content,
    confirm() {
      useAuthStore().logout()
      window.$message?.success('已退出登录')
      isConfirming = false
    },
    cancel() {
      isConfirming = false
    },
  })
  return false
}

export function resolveResError(code, message, needTip = true) {
  switch (code) {
    case 10002:
    case 10003:
    case 10004:
    case 10005:
    case 10006:
    case 10007:
    case 10008:
    case 20001:
    case 20101:
    case 20102:
    case 20103:
      return handleAuthExpired('登录已过期，是否重新登录？', needTip)
    // 请求体解密失败：后端只会回一句「请求参数不合法」，这里补上有指向性的排查方向
    case 10009:
      message = '请求安全校验失败，请刷新页面重试；若持续失败，请检查前端构建的 VITE_SIGN_SECRET 是否与后端 crypto.sign_secret 一致'
      console.error('[http] 请求被后端拒绝（10009）:', message)
      break
    default:
      message = message ?? `【${code}】: 未知异常!`
      break
  }
  needTip && window.$message?.error(message)
  return message
}
