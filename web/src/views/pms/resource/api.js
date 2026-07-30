// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  getMenuList: () => request.post('/menu/list'),
  getButtons: data => request.post('/menu/buttons', data),
  savePermission: data => request.post('/menu/save', data),
  togglePermission: id => request.post('/menu/toggle', { id }),
  deletePermission: id => request.post('/menu/delete', { id }),
}
