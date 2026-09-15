// Copyright © 2023 Ronnie Zhang (大脸怪). MIT License.

import dayjs from 'dayjs'

/**
 * @param {(object | string | number)} time
 * @param {string} format
 * @returns {string | null} 格式化后的时间字符串
 *
 */
export function formatDateTime(time = undefined, format = 'YYYY-MM-DD HH:mm:ss') {
  return dayjs(time).format(format)
}

export function formatDate(date = undefined, format = 'YYYY-MM-DD') {
  return formatDateTime(date, format)
}

/**
 * 按 value 取 options 里对应的 label，取不到返回「未知」。
 *
 * @param {Array | import('vue').Ref<Array>} options 选项数组，直接传数组或 ref 都可以
 * @param {number | string} value 目标值，字符串会转成数字再比
 * @returns {string} 命中的 label，未命中为「未知」
 */
export function getOptionsLabel(options, value) {
  const list = Array.isArray(options) ? options : options?.value
  if (!Array.isArray(list))
    return '未知'
  // null / undefined / 空串统一按「没有值」处理：Number(null) === 0 会把缺失值错认成第一项
  if (value === null || value === undefined || value === '')
    return '未知'
  const item = list.find(o => o.value === Number(value))
  return item ? item.label : '未知'
}

/**
 * @param {Function} fn
 * @param {number} wait
 * @returns {Function}  节流函数
 *
 */
export function throttle(fn, wait) {
  let context, args
  let previous = 0

  return function (...argArr) {
    const now = Date.now()
    context = this
    args = argArr
    if (now - previous > wait) {
      fn.apply(context, args)
      previous = now
    }
  }
}

/**
 * @param {Function} method
 * @param {number} wait
 * @param {boolean} immediate
 * @return {*} 防抖函数
 */
export function debounce(method, wait, immediate) {
  let timeout
  return function (...args) {
    const context = this
    if (timeout) {
      clearTimeout(timeout)
    }
    // 立即执行需要两个条件，一是immediate为true，二是timeout未被赋值或被置为null
    if (immediate) {
      /**
       * 如果定时器不存在，则立即执行，并设置一个定时器，wait毫秒后将定时器置为null
       * 这样确保立即执行后wait毫秒内不会被再次触发
       */
      const callNow = !timeout
      timeout = setTimeout(() => {
        timeout = null
      }, wait)
      if (callNow) {
        method.apply(context, args)
      }
    }
    else {
      // 如果immediate为false，则函数wait毫秒后执行
      timeout = setTimeout(() => {
        /**
         * args是一个类数组对象，所以使用fn.apply
         * 也可写作method.call(context, ...args)
         */
        method.apply(context, args)
      }, wait)
    }
  }
}

/**
 * @param {number} time 毫秒数
 * @returns 睡一会儿，让子弹暂停一下
 */
export function sleep(time) {
  return new Promise(resolve => setTimeout(resolve, time))
}

/**
 * @param {HTMLElement} el
 * @param {Function} cb
 * @return {ResizeObserver}
 */
export function useResize(el, cb) {
  const observer = new ResizeObserver((entries) => {
    cb(entries[0].contentRect)
  })
  observer.observe(el)
  return observer
}
