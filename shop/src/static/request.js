import axios from 'axios'
import { encryptRequest } from 'hejunjie-encrypted-request'
import Cookies from 'js-cookie'
import { clearAuthCookies } from '@/utils/auth'
import toast from '@/utils/toast'

// 加密版本号，与后端约定一致
const ENCRYPT_VERSION = 1

const service = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 10000, // 请求超时时间毫秒
  headers: {
    'Content-Type': 'application/json',
  },
})

// RSA 公钥缓存（Promise 防抖，仅拉取一次；失败自动复位允许重试）
let publicKeyPromise = null

// base64 DER → SPKI PEM（hejunjie-encrypted-request 需要 PEM 格式公钥）
const DER_LINE_RE = /.{1,64}/g
function derToPem(base64Der) {
  const lines = base64Der.match(DER_LINE_RE) || []
  return `-----BEGIN PUBLIC KEY-----\n${lines.join('\n')}\n-----END PUBLIC KEY-----\n`
}

// 校验公钥响应签名
async function verifyPubkeySign({ key_id: keyId, public_key: publicKey, timestamp, sign }) {
  const enc = new TextEncoder()
  const key = await crypto.subtle.importKey(
    'raw',
    enc.encode(import.meta.env.VITE_SIGN_SECRET),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const sig = await crypto.subtle.sign('HMAC', key, enc.encode(`pubkey:${keyId}${publicKey}${timestamp}`))
  const hex = Array.from(new Uint8Array(sig), b => b.toString(16).padStart(2, '0')).join('')
  if (hex !== sign) {
    toast.error('公钥签名校验失败, 调整配置')
    throw new Error('公钥签名校验失败')
  }
}

// 获取用于加密的 RSA 公钥（SPKI PEM）
function fetchPublicKey() {
  if (!publicKeyPromise) {
    publicKeyPromise = service
      .get('/api/shop/public-key')
      .then(async (res) => {
        if (res.code !== 0) {
          toast.error(`获取公钥失败: ${res.msg}`)
          throw new Error(`获取公钥失败: ${res.msg}`)
        }
        await verifyPubkeySign(res.data)
        return derToPem(res.data.public_key)
      })
      .catch((err) => {
        publicKeyPromise = null // 允许下次请求重试
        throw err
      })
  }
  return publicKeyPromise
}

// 登录态失效错误码（i18n 均表示"登录状态异常，请重新登录"）
const AUTH_EXPIRED_CODES = new Set(['10002', '10003', '10004', '10005', '10006', '10007', '10008', '20001'])

// 刷新 Promise：并发请求同时鉴权失败时只发一次刷新，其余等待同一 Promise（与 fetchPublicKey 同模式）
let refreshPromise = null

// 刷新成功 → 重新写入两个 cookie；无论成败都复位 Promise，下次失效时重新发起刷新
function refreshTokens() {
  const refreshToken = Cookies.get('user_refresh_token')
  if (!refreshToken) {
    return Promise.reject(new Error('缺少 refresh_token'))
  }
  if (!refreshPromise) {
    refreshPromise = service.post('/api/shop/liveuser/refresh', { token: refreshToken }, { _isRefresh: true }).then((res) => {
      if (res.code !== 0) {
        throw new Error(res.msg || '刷新 token 失败')
      }
      Cookies.set('user_access_token', res.data.access_token)
      Cookies.set('user_refresh_token', res.data.refresh_token)
      return res.data
    }).finally(() => {
      refreshPromise = null // 无论成败都复位，避免缓存已轮换的旧 token
    })
  }
  return refreshPromise
}

// 清理登录态并跳转登录页
function handleAuthFail() {
  clearAuthCookies()
  if (window.location.pathname === '/shop/login')
    return
  window.location.href = '/shop/login'
}

// 添加请求拦截器
service.interceptors.request.use(
  async (config) => {
    // 刷新请求本身不携带旧的 access_token，避免传递已过期凭证
    if (!config._isRefresh) {
      const accessToken = Cookies.get('user_access_token')
      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`
      }
    }
    const method = (config.method || 'get').toUpperCase()
    const hasBody = config.data != null && config.data !== '' && !(config.data instanceof FormData)
    const canEncrypt = globalThis.crypto?.subtle != null
    const isDev = import.meta.env.DEV
    // 重试时 config.data 已是第一遍加密后的密文，跳过二次加密，仅重新附加新的 access_token
    if (!config._retried && method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS' && hasBody) {
      // dev 或纯 HTTP（无 Web Crypto）→ 明文发送
      if (isDev || !canEncrypt) {
        if (!isDev && !canEncrypt)
          console.warn('[request] 当前为纯 HTTP 环境，无法使用 Web Crypto 加密，请求以明文发送')
      }
      else {
        let data = config.data
        if (typeof data === 'string')
          data = JSON.parse(data) // axios 对字符串会原样透传，先还原为对象
        const rsaPublicKey = await fetchPublicKey()
        config.data = await encryptRequest({ data, rsaPublicKey, signSecret: import.meta.env.VITE_SIGN_SECRET }, ENCRYPT_VERSION)
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 添加响应拦截器
service.interceptors.response.use(
  async (response) => {
    const config = response.config
    const code = String(response.data?.code ?? '')
    // 非登录态错误 → 正常返回响应包
    if (!AUTH_EXPIRED_CODES.has(code)) {
      return response.data
    }
    // 刷新请求本身失败（refresh_token 过期等）→ 抛错，由等待方统一兜底，防止递归刷新
    if (config._isRefresh) {
      throw new Error(response.data?.msg || '登录状态异常')
    }
    // 已重试一次仍失败 → 会话确实无效，清理登录态并跳转，不再递归
    if (config._retried) {
      handleAuthFail()
      throw new Error(response.data?.msg || '登录状态异常')
    }
    // 无 refresh_token 无法刷新 → 直接跳转登录
    if (!Cookies.get('user_refresh_token')) {
      handleAuthFail()
      throw new Error(response.data?.msg || '登录状态异常')
    }
    // 首次失效且具备刷新条件 → 刷新 + 重放原请求
    try {
      await refreshTokens()
      config._retried = true
      // 重新走请求拦截器：自动附加新的 access_token，_retried 跳过二次加密
      return service(config)
    }
    catch (err) {
      // 刷新失败（refresh_token 已失效）→ 清理登录态并跳转登录
      handleAuthFail()
      throw err
    }
  },
  (error) => {
    return Promise.reject(error)
  },
)
export default service
