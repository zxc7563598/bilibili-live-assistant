import request from '@/static/request'

export default {
  getAddressList: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            list: [
              {
                id: 1,
                type: 1,
                name: '张伟',
                phone: '138****8888',
                region: '广东省 深圳市 南山区',
                detail: '科技园路 1 号 A 座 1201 室',
                email: '',
                isDefault: true,
              },
              {
                id: 2,
                type: 1,
                name: '张伟',
                phone: '138****8888',
                region: '广东省 广州市 天河区',
                detail: '体育西路 100 号',
                email: '',
                isDefault: false,
              },
              {
                id: 3,
                type: 0,
                name: '张伟',
                phone: '',
                region: '',
                detail: '',
                email: 'zhangwei@example.com',
                isDefault: false,
              },
            ],
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  getConfirm: () => request.post('/api/shop/order/confirm'),
  reOrder: (ConfirmID) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            id: ConfirmID,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  confirmPayment: (ConfirmID, addressID) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            id: ConfirmID,
            address: addressID,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
}
