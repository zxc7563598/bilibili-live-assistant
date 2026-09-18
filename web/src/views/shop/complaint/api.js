import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/feedback/list', params),
  getDetails: id => request.post('/feedback/details', { id }),
}
