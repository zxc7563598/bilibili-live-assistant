import { createApp } from 'vue'
import AppBottomSheet from '@/components/base/AppBottomSheet.vue'
import AppButton from '@/components/base/AppButton.vue'
import AppCarousel from '@/components/base/AppCarousel.vue'
import AppIcon from '@/components/base/AppIcon.vue'
import AppImage from '@/components/base/AppImage.vue'
import AppNavBar from '@/components/base/AppNavBar.vue'
import AppSwitch from '@/components/base/AppSwitch.vue'
import Tag from '@/components/base/Tag.vue'
import App from './App.vue'
import router from './router/index.js'
import { applyPrimaryColor } from './utils/color'
import { initTheme } from './utils/theme'
import './style.css'

// 主题初始化
initTheme()

// 商户主题色：示例默认色
applyPrimaryColor('#965bff')

const app = createApp(App)

// 在路由切换时修改页面的title和meta标签
router.beforeEach((to) => {
  document.title = to.meta.title
  const metaTags = to.meta.metaTags || []
  metaTags.forEach((tag) => {
    const tagElement = document.createElement('meta')
    Object.keys(tag).forEach((key) => {
      tagElement.setAttribute(key, tag[key])
    })
    document.head.appendChild(tagElement)
  })
})

app.component('AppBottomSheet', AppBottomSheet)
app.component('AppCarousel', AppCarousel)
app.component('AppNavBar', AppNavBar)
app.component('AppSwitch', AppSwitch)
app.component('AppButton', AppButton)
app.component('AppImage', AppImage)
app.component('AppIcon', AppIcon)
app.component('Tag', Tag)

app.use(router)
app.mount('#app')
