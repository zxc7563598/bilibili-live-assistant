import request from '@/static/request'

export default {
  savedPassword: (old_password, new_password) => request.post('/api/shop/liveuser/change-password', { old_password, new_password }),
  logout: () => request.post('/api/shop/liveuser/logout'),
}
