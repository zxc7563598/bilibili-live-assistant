export default {
  savedFeedback: (type, content, contact) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          code: 0,
          data: {
            type,
            content,
            contact,
          },
          msg: 'success',
        })
      }, 1500)
    })
  },
}
