import request from '@/static/request'

export default {
  getConfig: () => request.get('/api/shop/login'),
  getAccount: account => request.post('/api/shop/liveuser/account', { account }),
  login: (account, password) => request.post('/api/shop/liveuser/login', { account, password }),
}
