import request from '@/static/request'

export default {
  getUserInfo: () => request.post('/api/shop/liveuser/info'),
  getShopList: (page, search) => request.post('/api/shop/product/list', { pageNo: page, pageSize: 10, name: search }),
}
