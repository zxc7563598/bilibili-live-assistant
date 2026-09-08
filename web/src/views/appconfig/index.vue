<template>
  <AppPage show-footer full>
    <div class="flex gap-10">
      <div class="min-w-0 flex-1 space-y-10">
        <n-card v-show="appConfigShow" id="module-basic" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                积分商城配置
              </div>
              <div class="mt-2 text-13 text-gray-400">
                配置积分商城的基本信息
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「安装」
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                积分商城通常为一个网站；
                在支持 PWA 的浏览器引擎中访问时（火狐/谷歌等非国产浏览器），允许用户将该网站作为一个 App 安装在手机中
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                站点名称
              </div>
              <n-input v-model:value="appConfigForm.site_name" type="text" placeholder="站点以及安装后展示的名称" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                站点说明
              </div>
              <n-input v-model:value="appConfigForm.site_description" type="text" placeholder="用户在浏览器安装本应用的说明" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「颜色」
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                应用启动背景色：当用户在支持 PWA 的浏览器作为应用安装到手机后，启动的瞬间在内容加载完成前展示的背景色，通常无需做任何变更
                网站主题色：站点中各种UI（边框/按钮等）通用的颜色，可以自行调整后访问积分商城查看效果
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                应用启动背景色
              </div>
              <n-color-picker v-model:value="appConfigForm.site_background_color" :show-alpha="false" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                网站主题色
              </div>
              <n-color-picker v-model:value="appConfigForm.site_theme_color" :show-alpha="false" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于网站图标与logo
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                网站图标：用户在浏览器中打开网站时，在网站标签页旁展示的图片
                网站logo：在网站中，例如登录页展示的logo
                图片均为1:1的方形图片，尺寸建议不超过 512 x 512，过大的图片会影响用户加载速度
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                网站图标
              </div>
              <UploadInput v-model:value="appConfigForm.site_icon" :upload="file => api.uploadImage(file, 'site_icon')" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                网站logo
              </div>
              <UploadInput v-model:value="appConfigForm.logo" :upload="file => api.uploadImage(file, 'logo')" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「用户自动注册」
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                积分商城需要登录才允许访问。
                通常情况下，用户通过签到/送礼等行为产生积分后才会被注册
                开启后，任意用户均可注册，不论他们是否产生了积分
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                无积分用户注册
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in enabledEnums" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="appConfigForm.register === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="appConfigForm.register = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于登录页背景图
              </div>
              <div class="mt-2 text-12 text-gray-500 dark:text-gray-400">
                非必须配置项，上传后登录页背景图将会被替换为上传的图片，通常建议配置 20:9 的图片作为背景
                不上传时登录页会通过配置的「网站主题色」生成背景
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页背景图
              </div>
              <UploadInput v-model:value="appConfigForm.login_bg" :upload="file => api.uploadImage(file, 'login_bg')" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页标题
              </div>
              <n-input v-model:value="appConfigForm.login_title" type="text" placeholder="登录页顶部显示的主标题" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页Slogan
              </div>
              <n-input v-model:value="appConfigForm.login_slogan" type="text" placeholder="登录页副标题或宣传语" />
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="appLoading" @click="apply()">
                保存配置
              </n-button>
            </div>
          </template>
        </n-card>
      </div>
    </div>
  </AppPage>
</template>

<script setup>
import api from './api'

// 基本配置模块
const appLoading = ref(false)
const appConfigShow = ref(false)
const appConfigForm = ref({
  site_name: '',
  site_description: '',
  site_background_color: '#f5f6f8',
  site_theme_color: '#965bff',
  site_icon: '',
  register: '1',
  logo: '',
  login_bg: '',
  login_title: '',
  login_slogan: '',
})

// 通用部分
const enabledEnums = [
  {
    label: '关闭',
    value: '0',
  },
  {
    label: '开启',
    value: '1',
  },
]

// 各模块配置存储
function apply() {
  if (appConfigForm.value.site_name.trim() === '') {
    return $message.warning('站点名称不可以为空')
  }
  if (appConfigForm.value.site_description.trim() === '') {
    return $message.warning('站点说明不可以为空')
  }
  if (appConfigForm.value.site_background_color.trim() === '') {
    return $message.warning('应用启动背景色不可以为空')
  }
  if (appConfigForm.value.site_theme_color.trim() === '') {
    return $message.warning('网站主题色')
  }
  if (appConfigForm.value.register.trim() === '') {
    return $message.warning('无积分用户注册')
  }
  if (appConfigForm.value.login_title.trim() === '') {
    return $message.warning('登录页标题')
  }
  appLoading.value = true
  api.applyData(
    appConfigForm.value.site_name,
    appConfigForm.value.site_description,
    appConfigForm.value.site_background_color,
    appConfigForm.value.site_theme_color,
    appConfigForm.value.site_icon,
    appConfigForm.value.register,
    appConfigForm.value.logo,
    appConfigForm.value.login_bg,
    appConfigForm.value.login_title,
    appConfigForm.value.login_slogan,
  ).then(() => {
    $message.success('保存成功')
  }).finally(() => {
    appLoading.value = false
  })
}

onMounted(() => {
  api.getData().then((res) => {
    Object.assign(appConfigForm.value, res.data)
    appConfigShow.value = true
  }).catch(() => {
    appConfigShow.value = false
  })
})
</script>
