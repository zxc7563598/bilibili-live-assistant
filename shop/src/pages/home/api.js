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
      }, 300)
    })
  },
  getShopList: (page, search) => {
    console.warn('getShopList', page, search)
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            total: 15,
            pageData: [
              {
                id: 1,
                name: '无线蓝牙耳机 主动降噪',
                cover: '',
                amount: 12800,
                type: 1,
              },
              {
                id: 2,
                name: '316 不锈钢保温杯 500ml',
                cover: '',
                amount: 6800,
                type: 1,
              },
              {
                id: 3,
                name: '定制帆布手提包',
                cover: 'https://shub.points.xin/attachment/goods/cover_image/image_681ad8392a90b5.30457608.png',
                amount: 4200,
                type: 0,
              },
              {
                id: 4,
                name: '智能运动手环',
                cover: '',
                amount: 360,
                type: 1,
              },
              {
                id: 5,
                name: '智能运动手环',
                cover: '',
                amount: 360,
                type: 1,
              },
              {
                id: 6,
                name: '智能运动手环',
                cover: '',
                amount: 360,
                type: 1,
              },
              {
                id: 7,
                name: '智能运动手环',
                cover: '',
                amount: 360,
                type: 1,
              },
              {
                id: 8,
                name: '智能运动手环',
                cover: '',
                amount: 360,
                type: 1,
              },
            ],
          },
          msg: 'success',
        })
      }, 2000)
    })
  },
}
