import request from '@/static/request'

export default {
  getUserInfo: () => request.post('/api/shop/liveuser/info'),
  getRoomID: () => request.post('/api/shop/liveuser/room-id'),
  logout: () => request.post('/api/shop/liveuser/logout'),
}
