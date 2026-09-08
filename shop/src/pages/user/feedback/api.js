import request from '@/static/request'

export default {
  savedFeedback: (type, content, contact) => request.post('/api/shop/feedback/submit', { type, content, contact }),
}
