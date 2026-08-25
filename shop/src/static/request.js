import axios from 'axios'
import Cookies from 'js-cookie'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 10000, // 请求超时时间毫秒
  headers: {
    'Content-Type': 'application/json',
  },
})

// 添加请求拦截器
service.interceptors.request.use(
  (config) => {
    const accessToken = Cookies.get('accessToken')
    if (accessToken)
      config.headers.Authorization = `Bearer ${accessToken}`
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 添加响应拦截器
service.interceptors.response.use(
  (response) => {
    switch (String(response.data.code)) {
      case '900006':
        window.location.href = '/shop/login'
        break
    }
    return response.data
  },
  (error) => {
    return Promise.reject(error)
  },
)
export default service
