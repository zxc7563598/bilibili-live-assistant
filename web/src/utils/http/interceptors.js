// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import api from '@/api'
import { useAuthStore } from '@/store'
import { handleAuthExpired, resolveResError } from './helpers'

const SUCCESS_CODES = [0, 200]
const EXPIRED_CODES = [10002, 10003, 10004, 10005, 10006, 10007, 10008, 20001]

// 👉 全局 refresh promise（替代 queue）
let refreshPromise = null

export function setupInterceptors(axiosInstance) {
  axiosInstance.interceptors.request.use(reqResolve, reqReject)
  axiosInstance.interceptors.response.use(resResolve, resReject)
  // 响应前
  function resResolve(response) {
    const { data, status, config, statusText, headers } = response
    const isJSON = headers['content-type']?.includes('application/json')
    if (!isJSON) {
      return Promise.resolve(data ?? response)
    }
    const code = data?.code ?? status
    if (SUCCESS_CODES.includes(code)) {
      return data
    }
    // 重新登录
    const isRefreshRequest = config?.skipAuthRefresh
    if (EXPIRED_CODES.includes(code) && !isRefreshRequest) {
      return replayRequest(axiosInstance, config)
    }
    // 处理失败
    const needTip = config?.needTip !== false
    const message = resolveResError(code, data?.msg ?? statusText, needTip)
    return Promise.reject({
      code,
      message,
      error: data ?? response,
    })
  }
}

// 请求前
function reqResolve(config) {
  if (config.needToken === false) {
    return config
  }
  const { accessToken } = useAuthStore()
  if (accessToken) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
}

// 请求失败
function reqReject(error) {
  return Promise.reject(error)
}

// 响应失败
async function resReject(error) {
  if (!error || !error.response) {
    const code = error?.code
    const message = resolveResError(code, error.message)
    return Promise.reject({ code, message, error })
  }
  const { data, status, config } = error.response
  const code = data?.code ?? status
  const needTip = config?.needTip !== false
  const message = resolveResError(code, data?.msg ?? error.message, needTip)
  return Promise.reject({
    code,
    message,
    error: data ?? error.response,
  })
}

// refresh & replay
async function replayRequest(axiosInstance, config) {
  if (config.__isRetryRequest) {
    return Promise.reject(new Error('请求已重试'))
  }
  config.__isRetryRequest = true
  try {
    const token = await getRefreshToken()
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
    return axiosInstance(config)
  }
  catch (err) {
    return Promise.reject(err)
  }
}

// 获取新 token（带 promise 缓存）
async function getRefreshToken() {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const authStore = useAuthStore()
      try {
        const res = await api.refreshToken(authStore.refreshToken, {
          skipAuthRefresh: true,
        })
        if (!SUCCESS_CODES.includes(res?.code)) {
          handleAuthExpired?.('登录已过期，是否重新登录？', true)
          throw new Error('刷新令牌失败')
        }
        const { accessToken, refreshToken } = res.data
        authStore.setToken(accessToken, refreshToken)
        return accessToken
      }
      catch (err) {
        handleAuthExpired?.('登录已过期，是否重新登录？', true)
        throw err
      }
      finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}
