// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  // 获取用户信息
  getUser: () => request.post('/auth/detail'),
  // 刷新token
  refreshToken: token => request.post('/auth/refresh', { token }, { needToken: false, skipAuthRefresh: true }),
  // 登出
  logout: () => request.post('/auth/logout', {}, { needTip: false }),
  // 获取角色权限
  getRolePermissions: () => request.post('/roles/permissions'),
  // 切换当前角色
  switchCurrentRole: code => request.post('/auth/switch-role', { code }),
  // 验证菜单路径
  validateMenuPath: path => request.post('/menu/validate', { path }),
}
