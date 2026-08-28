import { createApp } from 'vue'
import AppBottomSheet from '@/components/base/AppBottomSheet.vue'
import AppButton from '@/components/base/AppButton.vue'
import AppCarousel from '@/components/base/AppCarousel.vue'
import AppEmpty from '@/components/base/AppEmpty.vue'
import AppIcon from '@/components/base/AppIcon.vue'
import AppImage from '@/components/base/AppImage.vue'
import AppInput from '@/components/base/AppInput.vue'
import AppNavBar from '@/components/base/AppNavBar.vue'
import AppSegmentedControl from '@/components/base/AppSegmentedControl.vue'
import AppSelect from '@/components/base/AppSelect.vue'
import AppSkeleton from '@/components/base/AppSkeleton.vue'
import AppSwitch from '@/components/base/AppSwitch.vue'
import Tag from '@/components/base/Tag.vue'
import App from './App.vue'
import router from './router/index.js'
import { applyPrimaryColor } from './utils/color'
import { applySiteConfig, loadSiteConfig, siteConfig } from './utils/pwa'
import { initTheme } from './utils/theme'
import './style.css'

// 主题初始化
initTheme()

// 商户主题色：示例默认色
applyPrimaryColor('#965bff')

// 加载后台站点配置，动态设置 title / favicon 等（theme-color 由 utils/theme.js 接管）
loadSiteConfig().then(applySiteConfig)

const app = createApp(App)

// 在路由切换时修改页面的title和meta标签
router.beforeEach((to) => {
  document.title = siteConfig.value.name || to.meta.title
  const metaTags = to.meta.metaTags || []
  metaTags.forEach((tag) => {
    const tagElement = document.createElement('meta')
    Object.keys(tag).forEach((key) => {
      tagElement.setAttribute(key, tag[key])
    })
    document.head.appendChild(tagElement)
  })
})

app.component('AppSegmentedControl', AppSegmentedControl)
app.component('AppBottomSheet', AppBottomSheet)
app.component('AppCarousel', AppCarousel)
app.component('AppNavBar', AppNavBar)
app.component('AppSwitch', AppSwitch)
app.component('AppInput', AppInput)
app.component('AppSelect', AppSelect)
app.component('AppButton', AppButton)
app.component('AppImage', AppImage)
app.component('AppIcon', AppIcon)
app.component('AppEmpty', AppEmpty)
app.component('AppSkeleton', AppSkeleton)
app.component('Tag', Tag)

app.use(router)
app.mount('#app')
