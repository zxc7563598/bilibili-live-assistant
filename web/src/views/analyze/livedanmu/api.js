import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/livedanmu/list', params),
  fetchRoomGroups: () => request.post('/livedanmu/room'),
}
