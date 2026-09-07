import request from '@/static/request'

export default {
  getOrderList: (pageNo, pageSize, order_status) => request.post('/api/shop/order/list', { pageNo, pageSize, order_status }),
}
