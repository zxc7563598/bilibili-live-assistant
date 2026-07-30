// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  save: data => request.post('/roles/save', data),
  read: (params = {}) => request.post('/roles/list', params),
  delete: id => request.post('/roles/delete', { id }),
  getAllPermissionTree: () => request.post('/menu/list'),
  getAllUsers: (params = {}) => request.post('/admin/list', params),
  addRoleUsers: (roleId, adminIds) => request.post('/roles/add-role-users', { roleId, adminIds }),
  removeRoleUsers: (roleId, adminIds) => request.post('/roles/remove-role-users', { roleId, adminIds }),
}
