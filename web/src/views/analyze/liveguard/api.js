import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/liveguard/list', params),
  backfill: (params = {}) => request.post('/liveguard/backfill', params),
  // 补录弹窗取房间号默认值用，room_id 由直播间绑定写入，读它不打 B站 接口
  getRoomConfig: () => request.post('/robot/room/get'),
}
