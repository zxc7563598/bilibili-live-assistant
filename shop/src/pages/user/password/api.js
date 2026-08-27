export default {
  savedPassword: (oldPassword, newPassword) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            oldPassword,
            newPassword,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
  logout: () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {},
          msg: 'success',
        })
      }, 1500)
    })
  },
}
