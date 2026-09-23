import { request } from '@/utils'

export default {
  getList: (params = {}) => request.post('/product/list', params),
  productEnable: (id, enable) => request.post('/product/enable', { id, enable }),
  productMinMemberLevel: (id, min_member_level) => request.post('/product/min-member-level', { id, min_member_level }),
  getDetails: id => request.post('/product/details', { id }),
  productSave: data => request.post('/product/save', data),
  uploadImage: (file, scene) => {
    const fd = new FormData()
    fd.append('scene', scene)
    fd.append('file', file)
    return request.post('/upload/image', fd, { timeout: 60000 })
  },
  syncOss: param => request.post('/upload/oss-sync', param, { timeout: 60000 }),
}
