import request from '@/static/request'

export default {
  getUserInfo: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            avatar: 'https://danmusuite.hejunjie.life/dist/avatar.jpg',
            uid: '4325051',
            name: '哎呀又胖啦',
            points: 3280,
            stars: 1260,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  getRoomID: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            roomID: 22384516,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  logout: () => request.post('/api/shop/liveuser/logout'),
}
