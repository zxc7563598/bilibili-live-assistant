import { request } from '@/utils'

export default {
  // 登录相关
  getLoginStatus: () => request.post('/live/login/status'),
  getQRCode: () => request.post('/live/login/qrcode'),
  pollQRCode: qrcodeKey => request.post('/live/login/poll', { qrcodeKey }),
  logout: () => request.post('/live/login/logout'),

  // 直播间相关
  updateRoom: roomId => request.post('/live/room/update', { roomId }),
  sendDanmu: message => request.post('/live/room/send-danmu', { message }),

  // 监听相关
  startListener: () => request.post('/live/listener/start'),
  stopListener: () => request.post('/live/listener/stop'),
  getListenerStatus: () => request.post('/live/listener/status'),
}

/** 构建 WebSocket 连接 URL，token 通过 query string 传递（浏览器 WebSocket API 限制） */
export function buildWsUrl(token) {
  try {
    const baseUrl = import.meta.env.VITE_AXIOS_BASE_URL
    if (!baseUrl) {
      throw new Error('VITE_AXIOS_BASE_URL is not defined')
    }
    const url = new URL(baseUrl)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${url.host}/api/admin/live/messages/stream?token=${token}`
  }
  catch (error) {
    console.error('Failed to build WebSocket URL:', error)
    throw error
  }
}
