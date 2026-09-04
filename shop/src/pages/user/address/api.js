import request from '@/static/request'

export default {
  getAddressList: () => request.post('/api/shop/address/list'),
  getAddressDetails: id => request.post('/api/shop/address/detail', { id }),
  savedAddress: (id, name, phone, region_code, region, detail, email, type, is_default) => request.post('/api/shop/address/save', { id, name, phone, region_code, region, detail, email, type, is_default }),
  deleteAddress: id => request.post('/api/shop/address/delete', { id }),
}
