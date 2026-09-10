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
                积分商城是一个可「安装」到手机桌面的网页站点，本页用于配置它的名称、外观与登录页等基础信息
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「安装」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  积分商城本身是一个网页站点。用户用支持 PWA 的手机浏览器（如 Safari、Chrome、Edge）打开商城时，可以把网站「安装」到手机桌面，之后像普通 App 一样，点击桌面图标即可直接进入；微信等不支持 PWA 的浏览器里打开时仍是普通网页。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">站点名称</span>：显示在浏览器标签栏、收藏夹，以及安装后的手机桌面图标下方。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">站点说明</span>：用一句话介绍这个商城。用户在手机浏览器里执行「添加到主屏幕」时会看到这段说明，用来确认要安装的内容。
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                站点名称
              </div>
              <n-input v-model:value="appConfigForm.site_name" type="text" placeholder="例如：积分商城" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                站点说明
              </div>
              <n-input v-model:value="appConfigForm.site_description" type="text" placeholder="介绍商城的一句话，例如：欢迎来到 XX 的积分商城" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「颜色」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">应用启动背景色</span>：只会在用户从桌面图标打开商城、首页内容渲染完成前的一瞬间显示，停留时间极短，通常无需修改。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">网站主题色</span>：商城的主色调，用于按钮、边框、链接、选中状态等界面元素；调整后打开积分商城即可看到整体效果。
                </div>
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
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">网站图标（Favicon）</span>：显示在浏览器标签栏、收藏夹里的小图标；网站被安装为 App 后，同样用作手机桌面的应用图标。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">网站 Logo</span>：展示在商城页面内（例如登录页顶部）的品牌图，通常比图标更大、更醒目。
                </div>
                <div>
                  两者都要求 1:1 的正方形图片，建议不超过 512×512 像素；文件过大或分辨率过高都会拖慢用户打开网页的速度。
                </div>
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
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  积分商城需要登录才能访问。通过签到、送礼等行为产生过积分的用户会被系统自动注册，直接登录即可；还没有注册的用户在登录时如何处理，由下面的开关决定：
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">关闭</span>：未注册的用户登录时会被提示「尚未注册」，无法进入商城；需要先产生积分、被系统登记为商城用户后才能登录。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">开启（默认）</span>：未注册的用户直接登录即可，系统会自动为其注册。即使没有任何积分（进去后暂时换不了东西），也可以先浏览商城。
                </div>
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
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">上传背景图</span>：商城登录页的背景会替换为这张图片，建议使用宽高比约 20:9 的横向大图，在手机上展示效果最好。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">留空时</span>：登录页会以上面配置的「网站主题色」自动生成纯色背景，此配置项非必填。
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页背景图
              </div>
              <UploadInput v-model:value="appConfigForm.login_bg" :upload="file => api.uploadImage(file, 'login_bg')" :sync-oss="api.syncOss" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页标题
              </div>
              <n-input v-model:value="appConfigForm.login_title" type="text" placeholder="显示在登录页顶部的大标题，例如：积分商城" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                登录页Slogan
              </div>
              <n-input v-model:value="appConfigForm.login_slogan" type="text" placeholder="显示在主标题下方的一句话宣传语，例如：签到送礼，好礼不断" />
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
        <n-card v-show="appConfigShow" id="module-basic" size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                阿里云OSS配置
              </div>
              <div class="mt-2 text-13 text-gray-400">
                配置阿里云OSS，以便将图片托管到阿里云，加快用户图片访问速度
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「OSS」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  商城用到的图片默认存放在这台服务器上，加载会占用本机带宽。配置阿里云 OSS 后，图片会托管到阿里云，由阿里云分布在全国的节点就近分发，用户在手机、电脑上打开商城时图片加载更快、更稳定，也能减轻这台服务器的压力。
                </div>
                <div>
                  开启需要在你的阿里云账号下准备两样东西：一个存储空间（Bucket）和一对访问密钥（AccessKey），都可以在下方指引的阿里云控制台页面免费创建、申请，再按对应位置填到本页即可。
                </div>
              </div>
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「bucket 名 与 OSS 地址」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  这两项都来自你创建的存储空间：在阿里云 OSS 控制台的 <a class="text-primary hover:underline" href="https://oss.console.aliyun.com/bucket" target="_blank" rel="noreferrer">Bucket 列表页</a> 点击「创建 Bucket」，填入一个英文名称、选择一个离你的用户较近的地域，直接创建即可，无需其它额外配置。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">bucket 名</span>：就是创建 Bucket 时填写的那个英文名称，例如 mall-images。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">OSS 地址</span>：进入刚创建的 Bucket，在「概览」页里找到 Endpoint（地域节点），形如 oss-cn-hangzhou.aliyuncs.com，将其填入即可，前面带不带 https:// 都可以。
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                OSS 地址
              </div>
              <n-input v-model:value="appConfigForm.oss_endpoint" type="text" placeholder="例如：oss-cn-hangzhou.aliyuncs.com" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                bucket 名
              </div>
              <n-input v-model:value="appConfigForm.oss_bucket" type="text" placeholder="例如：mall-images" />
            </div>
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「AccessKey」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  AccessKey 是阿里云的访问密钥，系统凭它来读写你的 OSS。在阿里云 RAM 控制台的 <a class="text-primary hover:underline" href="https://ram.console.aliyun.com/profile/access-keys" target="_blank" rel="noreferrer">访问密钥管理页</a> 点击「创建 AccessKey」，会生成一对 <span class="text-gray-700 font-medium dark:text-gray-200">AccessKey ID</span> 与 <span class="text-gray-700 font-medium dark:text-gray-200">AccessKey Secret</span>，分别填入下面两项。
                </div>
                <div>
                  两点提醒：Secret 只在创建时完整展示一次，请立即复制并妥善保存，忘记后只能重新创建；建议使用 RAM 子账号的密钥并只授予该 Bucket 的上传权限，不要直接填主账号的密钥，以免泄露后影响整个账号的安全。
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                AccessKey ID
              </div>
              <n-input v-model:value="appConfigForm.oss_access_key_id" type="text" placeholder="创建 AccessKey 时生成，例如 LTAI5t…" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                AccessKey Secret
              </div>
              <n-input v-model:value="appConfigForm.oss_access_key_secret" type="password" show-password-on="mousedown" placeholder="创建 AccessKey 时生成，仅展示一次" />
            </div>
          </div>
          <template #footer>
            <div class="flex justify-end">
              <n-button size="small" type="primary" :loading="appOssLoading" @click="applyOss()">
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
const appOssLoading = ref(false)
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
  oss_endpoint: '',
  oss_access_key_id: '',
  oss_access_key_secret: '',
  oss_bucket: '',
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

function applyOss() {
  if (appConfigForm.value.oss_endpoint.trim() === '') {
    return $message.warning('OSS 地址 不可以为空')
  }
  if (appConfigForm.value.oss_access_key_id.trim() === '') {
    return $message.warning('AccessKey ID 不可以为空')
  }
  if (appConfigForm.value.oss_access_key_secret.trim() === '') {
    return $message.warning('AccessKey Secret 不可以为空')
  }
  if (appConfigForm.value.oss_bucket.trim() === '') {
    return $message.warning('bucket 名不可以为空')
  }
  appOssLoading.value = true
  api.applyOss(
    appConfigForm.value.oss_endpoint,
    appConfigForm.value.oss_access_key_id,
    appConfigForm.value.oss_access_key_secret,
    appConfigForm.value.oss_bucket,
  ).then(() => {
    $message.success('保存成功')
  }).finally(() => {
    appOssLoading.value = false
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
