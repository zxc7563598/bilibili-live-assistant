// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  login: data => request.post('/auth/login', data, { needToken: false }),
  isCaptcha: () => request.post('/auth/captcha', { needToken: false }),
}
