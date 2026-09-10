// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  getData: () => request.post('/appconfig/data'),
  applyData: (site_name, site_description, site_background_color, site_theme_color, site_icon, register, logo, login_bg, login_title, login_slogan) => request.post('/appconfig/save', { site_name, site_description, site_background_color, site_theme_color, site_icon, register, logo, login_bg, login_title, login_slogan }),
  applyOss: (oss_endpoint, oss_access_key_id, oss_access_key_secret, oss_bucket) => request.post('/appconfig/oss_save', { oss_endpoint, oss_access_key_id, oss_access_key_secret, oss_bucket }),
  uploadImage: (file, scene) => {
    const fd = new FormData()
    fd.append('scene', scene)
    fd.append('file', file)
    return request.post('/upload/image', fd, { timeout: 60000 })
  },
  syncOss: param => request.post('/upload/oss-sync', param, { timeout: 60000 }),
}
