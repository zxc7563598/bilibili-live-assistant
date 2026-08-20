import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/livegift/list', params),
  fetchRoomGroups: () => request.post('/livegift/room'),
}
