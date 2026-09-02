import request from '@/static/request'

export default {
  getShopDetails: id => request.post('/api/shop/product/detail', { id }),
}
