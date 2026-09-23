import { request } from '@/utils'

export default {
  /**
   * 获取用户三个档位的大航海到期时间。
   *
   * 返回值为 Unix 秒（当天的 0 点），未设置的档位是 null。
   * 已过期的档位也照原值返回，弹窗按原值回填，方便在原有时间上继续改。
   */
  getGuardExpire: user_id => request.post('/liveuser/guard-expire', { user_id }),
  /**
   * 变更用户的大航海身份。
   *
   * 三个到期时间都是 Unix 秒（当天的 0 点），传 null 表示清空该档位。
   * 三个时间一起提交，整体覆盖原有设置；后端按「到期时间大于当前时间」
   * 推导生效身份，多个档位同时有效时高等级覆盖低等级。
   *
   * 调用方是公共弹窗组件 GuardExpireModal，商城与数据分析两个列表页共用。
   */
  saveGuardExpire: params => request.post('/liveuser/update-guard-expire', params),
}
