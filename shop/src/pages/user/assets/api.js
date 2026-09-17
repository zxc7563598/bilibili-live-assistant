import request from '@/static/request'

export default {
  getAssetList: (pageNo, pageSize, credit_type) => request.post('/api/shop/liveuser/assets', { pageNo, pageSize, credit_type }),
}
