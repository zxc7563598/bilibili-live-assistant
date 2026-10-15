import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/liveguard/list', params),
  backfill: (params = {}) => request.post('/liveguard/backfill', params),
}
