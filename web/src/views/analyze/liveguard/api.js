import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/liveguard/list', params),
  backfill: (params = {}) => request.post('/liveguard/backfill', params),
  // 补录弹窗取房间号默认值用，room_id 由直播间绑定写入，读它不打 B站 接口
  getRoomConfig: () => request.post('/robot/room/get'),

  // 帮用户下单：商品、规格与地址都借订单/地址/商品模块的后台接口
  productList: (params = {}) => request.post('/product/list', params),
  productDetails: id => request.post('/product/details', { id }),
  userAddressList: params => request.post('/address/list', params),
  userAddressSave: data => request.post('/address/save', data),
  placeForUser: data => request.post('/order/place-for-user', data),
}
