import request from '@/static/request'

export default {
  getAddressList: type => request.post('/api/shop/address/list', { type }),
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
