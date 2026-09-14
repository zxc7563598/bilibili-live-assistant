import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/liveuser/list', params),
  getUserDetails: user_id => request.post('/liveuser/details', { user_id }),
  getAssets: (params = {}) => request.post('/liveuser/assets', params),
  saveAssets: (user_id, credit_type, change_type, change_amount, remark) => request.post('/liveuser/save-assets', { user_id, credit_type, change_type, change_amount, remark }),
  resetPassword: (user_id, password) => request.post('/liveuser/reset-password', { user_id, password }),
}
