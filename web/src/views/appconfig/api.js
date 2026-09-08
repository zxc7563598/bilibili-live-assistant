// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  getData: () => request.post('/appconfig/data'),
  applyData: (site_name, site_description, site_background_color, site_theme_color, site_icon, register, logo, login_bg, login_title, login_slogan) => request.post('/appconfig/save', { site_name, site_description, site_background_color, site_theme_color, site_icon, register, logo, login_bg, login_title, login_slogan }),
  uploadImage: (file, scene) => {
    const fd = new FormData()
    fd.append('scene', scene)
    fd.append('file', file)
    return request.post('/appconfig/upload', fd, { timeout: 60000 })
  },
}
