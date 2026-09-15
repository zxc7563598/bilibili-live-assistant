import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/order/list', params),
  getDetails: id => request.post('/order/details', { id }),
  updateShipStatus: data => request.post('/order/ship-status', data),
  updateOrderStatus: (id, order_status) => request.post('/order/status', { id, order_status }),
  updateReceiverInfo: data => request.post('/order/receiver', data),
}
