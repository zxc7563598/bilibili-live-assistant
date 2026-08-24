import Vue3WordCloudBase from 'vue3-word-cloud'

// vue3-word-cloud 挂载后会用 ~1s 轮询检测容器尺寸变化，但停止条件检查的是 Vue 2
// 的 _isDestroyed（Vue 3 里永远不会被置位），导致组件每次卸载都会残留一个定时器。
// 这里在 beforeUnmount 手动置位 _isDestroyed，下一次轮询 tick 即停止，避免累积。
const Vue3WordCloud = {
  ...Vue3WordCloudBase,
  beforeUnmount() {
    this._isDestroyed = true
  },
}

export default Vue3WordCloud
