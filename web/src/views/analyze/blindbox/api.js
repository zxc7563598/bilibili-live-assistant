import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/livegift/blindbox', params),
  fetchRoomGroups: () => request.post('/livegift/room'),
}
