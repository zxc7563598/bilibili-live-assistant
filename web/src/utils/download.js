// 去掉结尾的斜杠，避免与后续路径拼出双斜杠
const TRAILING_SLASHES = /\/+$/

/**
 * API 基地址：开发环境由 .env.development 指向后端端口，生产环境为空串（同源）。
 * 与 ws 地址拼接（views/room/api.js）保持同一套解析方式。
 */
export function resolveApiBase() {
  return (import.meta.env.VITE_AXIOS_BASE_URL || window.location.origin).replace(TRAILING_SLASHES, '')
}

/**
 * 触发浏览器原生下载。
 *
 * 用隐藏的 <a> + click，而不是 window.open：后者在 await 之后会被弹窗拦截器挡掉；
 * 也不能用 window.location，那会在请求出错时把当前页面导航走。
 *
 * 文件名以响应的 Content-Disposition 为准 —— 本地开发是跨域，同源才生效的
 * download 属性会被浏览器忽略。
 */
export function downloadByUrl(url) {
  const a = document.createElement('a')
  a.href = url.startsWith('http') ? url : resolveApiBase() + url
  a.rel = 'noopener'
  a.style.display = 'none'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
