// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import axios from 'axios'
import { setupInterceptors } from './interceptors'

export function createAxios(options = {}) {
  const defaultOptions = {
    baseURL: `${import.meta.env.VITE_AXIOS_BASE_URL}/api/admin`,
    timeout: 12000,
    // 请求体加密开关，拦截器按此判断；mock 实例单独关掉
    encrypt: true,
  }
  const service = axios.create({
    ...defaultOptions,
    ...options,
  })
  setupInterceptors(service)
  return service
}

export const request = createAxios()

export const mockRequest = createAxios({
  baseURL: '/mock-api',
  encrypt: false,
})
