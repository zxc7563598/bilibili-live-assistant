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
                regionCode: ['440000', '440300', '440305'],
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
                regionCode: ['440000', '440100', '440106'],
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
  getAddressDetails: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            id: 1,
            type: 1,
            name: '张伟',
            phone: '138****8888',
            regionCode: ['370000', '370100', '370116'],
            detail: '科技园路 1 号 A 座 1201 室',
            email: '',
            isDefault: true,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  savedAddress: (id, type, name, phone, regionCode, detail, email, isDefault) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            id,
            type,
            name,
            phone,
            regionCode,
            detail,
            email,
            isDefault,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
}
