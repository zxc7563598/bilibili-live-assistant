import axios from 'axios'
import { encryptRequest } from 'hejunjie-encrypted-request'
import Cookies from 'js-cookie'
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

// 添加请求拦截器
service.interceptors.request.use(
  async (config) => {
    const accessToken = Cookies.get('accessToken')
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    const method = (config.method || 'get').toUpperCase()
    const hasBody = config.data != null && config.data !== '' && !(config.data instanceof FormData)
    const canEncrypt = globalThis.crypto?.subtle != null
    const isDev = import.meta.env.DEV
    if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS' && hasBody) {
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
