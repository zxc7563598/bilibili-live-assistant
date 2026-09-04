import request from '@/static/request'

export default {
  getAddressList: type => request.post('/api/shop/address/list', { type }),
  getConfirm: () => request.post('/api/shop/order/confirm'),
  reOrder: id => request.post('/api/shop/order/again', { id }),
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
