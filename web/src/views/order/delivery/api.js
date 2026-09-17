import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/order/list', params),
}
