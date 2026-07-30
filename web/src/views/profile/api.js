// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  changePassword: data => request.post('/auth/change-password', data),
  updateProfile: data => request.post('/admin/update-profile', data),
}
