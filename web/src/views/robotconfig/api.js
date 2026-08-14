// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import { request } from '@/utils'

export default {
  getRoom: () => request.post('/robot/room/get'), // 获取房间模块配置
  getSign: () => request.post('/robot/sign/get'), // 获取签到模块配置
  getAd: () => request.post('/robot/ad/get'), // 获取定时广告模块配置
  getGift: () => request.post('/robot/gift/get'), // 获取礼物答谢模块配置
  getPk: () => request.post('/robot/pk/get'), // 获取PK播报模块配置
  getWelcome: () => request.post('/robot/welcome/get'), // 获取进房欢迎模块配置
  getFollow: () => request.post('/robot/follow/get'), // 获取感谢关注模块配置
  getShare: () => request.post('/robot/share/get'), // 获取感谢分享模块配置
  getReply: () => request.post('/robot/reply/get'), // 获取自动回复模块配置
  applyRoom: (is_listening, max_name_length, name_trim_mode) => request.post('/robot/room/apply', { is_listening, max_name_length, name_trim_mode }), // 变更房间模块配置
  applySign: (enabled, scene, requirement, reward_type, reward_amount, keyword, query_keyword, success_reply, fail_reply, repeat_reply, query_reply) => request.post('/robot/sign/apply', { enabled, scene, requirement, reward_type, reward_amount, keyword, query_keyword, success_reply, fail_reply, repeat_reply, query_reply }), // 变更签到模块配置
  applyAd: (enabled, scene, interval, send_mode, content) => request.post('/robot/ad/apply', { enabled, scene, interval, send_mode, content }), // 变更定时广告模块配置
  applyGift: (enabled, scene, requirement, show_count, merge_gift, include_blindbox, min_battery, content) => request.post('/robot/gift/apply', { enabled, scene, requirement, show_count, merge_gift, include_blindbox, min_battery, content }), // 变更礼物答谢模块配置
  applyPk: (enabled, content) => request.post('/robot/pk/apply', { enabled, content }), // 变更PK播报模块配置
  applyWelcome: (enabled, scene, requirement, content) => request.post('/robot/welcome/apply', { enabled, scene, requirement, content }), // 变更进房欢迎模块配置
  applyFollow: (enabled, scene, requirement, content) => request.post('/robot/follow/apply', { enabled, scene, requirement, content }), // 变更感谢关注模块配置
  applyShare: (enabled, scene, requirement, content) => request.post('/robot/share/apply', { enabled, scene, requirement, content }), // 变更感谢分享模块配置
  applyReply: (enabled, scene, requirement, content) => request.post('/robot/reply/apply', { enabled, scene, requirement, content }), // 变更自动回复模块配置

}
