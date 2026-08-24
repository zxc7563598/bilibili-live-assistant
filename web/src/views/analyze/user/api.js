import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/liveuser/list', params),
  getMonthlyByUID: (params = {}) => request.post('/liveuser/monthly', params),
  getDanmuByUID: (params = {}) => request.post('/liveuser/danmu', params),
}
