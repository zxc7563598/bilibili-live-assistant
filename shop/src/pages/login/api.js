export default {
  getConfig: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            loginBg: '', // 'https://shub.points.xin/attachment/shop-config/login-background-image/image_67aea68f6a9343.86099094.jpeg',
            logo: '', // 'https://danmusuite.hejunjie.life/dist/avatar.jpg',
            title: '积分商城',
            slogan: '登录后可兑换积分好礼',
          },
          message: 'success',
        })
      }, 1500)
    })
  },
  getAccount: (account) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            account: account === 1234567,
          },
          message: 'success',
        })
      }, 1500)
    })
  },
  login: (uid, password) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            uid,
            password,
          },
          message: 'success',
        })
      }, 1500)
    })
  },
}
