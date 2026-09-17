import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/pk/list', params),
  fetchRoomGroups: () => request.post('/pk/room'),
}
