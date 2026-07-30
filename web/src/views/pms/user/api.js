// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {

  read: (params = {}) => request.post('/admin/list', params),
  delete: id => request.post('/admin/delete', { id }),
  save: data => request.post('/admin/save', data),
  resetPwd: (id, password) => request.post('/admin/update-password', { id, password }),
  getAllRoles: () => request.post('/roles/all'),
}
