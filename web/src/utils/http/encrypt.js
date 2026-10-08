// 请求体加密：与商城（shop/src/static/request.js）共用同一套机制与同一个密钥
//
// 线格式 { en_data, enc_payload, timestamp, sign }，后端由
// internal/middleware/requestDecrypt.go 解密，算法与密钥约定见 pkg/crypto。
import { encryptRequest } from 'hejunjie-encrypted-request'

// 加密版本号，与后端约定一致
const ENCRYPT_VERSION = 1

// RSA 公钥缓存（Promise 防抖，仅拉取一次；失败自动复位允许重试）
let publicKeyPromise = null

// base64 DER → SPKI PEM（hejunjie-encrypted-request 需要 PEM 格式公钥）
const DER_LINE_RE = /.{1,64}/g
function derToPem(base64Der) {
  const lines = base64Der.match(DER_LINE_RE) || []
  return `-----BEGIN PUBLIC KEY-----\n${lines.join('\n')}\n-----END PUBLIC KEY-----\n`
}

// 校验公钥响应签名，防止公钥在下发途中被替换
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
    window.$message?.error('公钥签名校验失败，请检查 VITE_SIGN_SECRET 与后端 crypto.sign_secret 是否一致')
    throw new Error('公钥签名校验失败')
  }
}

// 获取用于加密的 RSA 公钥（SPKI PEM）
//
// 公钥接口在全局 /api/public-key 上，而本实例的 baseURL 是 /api/admin，
// 相对路径会被 axios 拼成 /api/admin/api/public-key，所以这里单独覆盖 baseURL。
// 传入 axiosInstance 而不是直接 import request，避免与 index.js 形成循环依赖。
export function fetchPublicKey(axiosInstance) {
  if (!publicKeyPromise) {
    publicKeyPromise = axiosInstance
      .get('/api/public-key', { baseURL: import.meta.env.VITE_AXIOS_BASE_URL })
      .then(async (res) => {
        if (res.code !== 0) {
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

// 按需加密请求体。不需要加密的情况直接返回，config 原样发出。
async function encryptConfigData(config, axiosInstance) {
  const method = (config.method || 'get').toUpperCase()
  if (method === 'GET' || method === 'HEAD' || method === 'OPTIONS') {
    return
  }
  if (config.data == null || config.data === '') {
    return
  }
  // 上传走 multipart，后端按非 JSON 直通，不参与加密
  if (config.data instanceof FormData) {
    return
  }
  // axios 对字符串 body 原样透传，加密后后端解出来仍是字符串、绑定必然失败：
  // 能解析成 JSON 的还原成对象，解析不了的（非 JSON 文本）干脆不加密
  if (typeof config.data === 'string') {
    try {
      config.data = JSON.parse(config.data)
    }
    catch {
      return
    }
  }
  // 开发环境一律明文：本地调试时看得到 payload，也不必为每个 dev 环境配密钥
  if (import.meta.env.DEV) {
    return
  }
  // Web Crypto 只在安全上下文可用（HTTPS 或 localhost），纯 HTTP 访问时只能明文
  if (globalThis.crypto?.subtle == null) {
    console.warn('[http] 当前非安全上下文（需 HTTPS 或 localhost），无法使用 Web Crypto 加密，请求以明文发送')
    return
  }
  const signSecret = import.meta.env.VITE_SIGN_SECRET
  if (!signSecret) {
    console.warn('[http] 未配置 VITE_SIGN_SECRET，请求以明文发送')
    return
  }
  const rsaPublicKey = await fetchPublicKey(axiosInstance)
  config.data = await encryptRequest({ data: config.data, rsaPublicKey, signSecret }, ENCRYPT_VERSION)
}

export { encryptConfigData }
