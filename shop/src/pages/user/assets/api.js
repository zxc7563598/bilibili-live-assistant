export default {
  getAssetList: (page, type) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        const assetLogs = [
          { id: 1, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 2, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 3, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 4, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 5, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 6, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
          { id: 7, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 8, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 9, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 10, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 11, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 12, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
          { id: 13, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 14, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 15, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 16, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 17, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 18, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
          { id: 19, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 20, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 21, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 22, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 23, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 24, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
          { id: 25, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 26, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 27, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 28, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 29, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 30, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
          { id: 31, title: '观看直播奖励', value: '+50', positive: true, type: 1, time: '2026-08-22 21:30', balance: 3280 },
          { id: 32, title: '兑换商品 · 无线蓝牙耳机', value: '-12800', positive: false, type: 1, time: '2026-08-20 14:32', balance: 3230 },
          { id: 33, title: '每日签到', value: '+10', positive: true, type: 1, time: '2026-08-18 08:00', balance: 16030 },
          { id: 34, title: '直播间互动任务', value: '+30', positive: true, type: 0, time: '2026-08-21 20:10', balance: 1260 },
          { id: 35, title: '活动奖励 · 周榜', value: '+200', positive: true, type: 1, time: '2026-08-12 12:00', balance: 16020 },
          { id: 36, title: '兑换商品 · 智能手环', value: '-360', positive: false, type: 0, time: '2026-07-30 20:05', balance: 1230 },
        ]
        const list = assetLogs.filter(o => o.type === type)
        resolve({
          code: 0,
          data: {
            total: list.length * 2,
            pageData: list,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
}
