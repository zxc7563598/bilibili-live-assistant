import request from '@/static/request'

export default {
  getAddressList: type => request.post('/api/shop/address/list', { type }),
  getConfirm: () => request.post('/api/shop/order/confirm'),
  reOrder: id => request.post('/api/shop/order/again', { id }),
  confirmPayment: (draft_id, address_id) => request.post('/api/shop/order/payment', { draft_id, address_id }),
}
