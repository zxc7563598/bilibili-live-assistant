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
      return handleAuthExpired('登录已过期，是否重新登录？', needTip)
    default:
      message = message ?? `【${code}】: 未知异常!`
      break
  }
  needTip && window.$message?.error(message)
  return message
}
