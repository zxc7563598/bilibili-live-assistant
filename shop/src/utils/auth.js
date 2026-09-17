// 登录态 cookie 清理：供登出与鉴权失败兜底共用，避免 cookie 名散落多处
import Cookies from 'js-cookie'

export function clearAuthCookies() {
  Cookies.remove('user_access_token')
  Cookies.remove('user_refresh_token')
}
