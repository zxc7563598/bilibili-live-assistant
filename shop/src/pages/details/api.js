import request from '@/static/request'

export default {
  getShopDetails: id => request.post('/api/shop/product/detail', { id }),
  orderPlace: (sku_id, count) => request.post('/api/shop/order/place', { sku_id, count }),
}
