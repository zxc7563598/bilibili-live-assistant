// 颜色工具：将商户主题色注入 CSS 变量，并派生暗夜模式调亮色与对比文字色。
// 派生色（hover/active/soft/ring）在 CSS 里用 color-mix() 基于 --primary 计算，
const HEX_RE = /^[0-9a-f]{6}$/i

export function normalizeHex(hex) {
  if (!hex)
    return '#2563eb'
  let h = String(hex).trim().replace('#', '')
  if (h.length === 3)
    h = h.split('').map(c => c + c).join('')
  if (!HEX_RE.test(h))
    return '#2563eb'
  return `#${h.toLowerCase()}`
}

export function hexToRgb(hex) {
  const h = normalizeHex(hex).slice(1)
  return {
    r: Number.parseInt(h.slice(0, 2), 16),
    g: Number.parseInt(h.slice(2, 4), 16),
    b: Number.parseInt(h.slice(4, 6), 16),
  }
}

export function rgbToHex({ r, g, b }) {
  const to = n => Math.round(Math.min(255, Math.max(0, n))).toString(16).padStart(2, '0')
  return `#${to(r)}${to(g)}${to(b)}`
}

// amount ∈ [0,1]：0 = 原色，1 = 纯白
export function lighten(hex, amount) {
  const { r, g, b } = hexToRgb(hex)
  return rgbToHex({
    r: r + (255 - r) * amount,
    g: g + (255 - g) * amount,
    b: b + (255 - b) * amount,
  })
}

export function darken(hex, amount) {
  const { r, g, b } = hexToRgb(hex)
  return rgbToHex({ r: r * (1 - amount), g: g * (1 - amount), b: b * (1 - amount) })
}

// WCAG 相对亮度 0..1
export function luminance(hex) {
  const { r, g, b } = hexToRgb(hex)
  const f = (c) => {
    const s = c / 255
    return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4
  }
  return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b)
}

// 根据亮度返回可读文字色（白或深色）
export function onColor(hex) {
  return luminance(hex) > 0.5 ? '#1b1f2a' : '#ffffff'
}

// 注入商户主题色，会同时计算暗夜模式下的调亮色与对比文字色，保证深色背景依然可读
export function applyPrimaryColor(hex) {
  const base = normalizeHex(hex)
  const dark = lighten(base, 0.22) // 暗夜模式下适当调亮
  const root = document.documentElement
  root.style.setProperty('--primary-base', base)
  root.style.setProperty('--primary', base) // 日间
  root.style.setProperty('--primary-dark', dark) // 暗夜
  root.style.setProperty('--on-primary', onColor(base))
  root.style.setProperty('--on-primary-dark', onColor(dark))
}
