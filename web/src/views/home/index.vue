<template>
  <AppPage show-footer full>
    <div v-if="isLoggedIn" class="mb-4 flex gap-4">
      <n-card size="small" :bordered="false" class="min-w-200 w-auto flex-shrink-0">
        <div class="flex flex-col items-center">
          <div class="relative">
            <n-avatar round :size="64" :src="loginStatus.face" :bordered="true" />
            <div class="absolute bottom-1 right-1 h-3 w-3 rounded-full bg-green-500 ring-2 ring-white" />
          </div>
          <div class="mt-3 w-full text-center">
            <div class="truncate text-15 font-medium">
              {{ loginStatus.username }}
            </div>
            <div class="mt-1 text-12 text-gray-400">
              UID {{ loginStatus.uid }}
            </div>
          </div>
          <n-button secondary type="error" size="small" class="mt-4 w-full" :loading="logoutLoading" @click="handleLogout">
            退出登录
          </n-button>
        </div>
      </n-card>
      <n-card title="直播间管理" :bordered="false" size="small" class="min-w-300 flex-1">
        <div v-if="listenerLoading && !listenerStatus" class="f-c-c py-6">
          <n-spin size="small" />
        </div>
        <div v-else-if="!hasRoom" class="flex items-center gap-3">
          <n-input v-model:value="roomIdInput" placeholder="请输入直播间 ID" size="small" :disabled="roomBindLoading" style="max-width: 240px" @keyup.enter="handleRoomUpdate" />
          <n-button size="small" type="primary" :loading="roomBindLoading" @click="handleRoomUpdate">
            确认绑定
          </n-button>
        </div>
        <template v-else>
          <div class="mb-3 flex items-center gap-2">
            <span class="truncate text-16 font-semibold">{{ listenerStatus.title || '直播间' }}</span>
            <n-tag :type="liveStatusTagType" size="small" round>
              {{ liveStatusLabel }}
            </n-tag>
            <n-tag :type="isRunning ? 'success' : 'default'" size="small" round>
              <template #icon>
                <i :class="isRunning ? 'i-material-symbols:play-circle-outline' : 'i-material-symbols:pause-circle-outline'" />
              </template>
              {{ isRunning ? '监听中' : '已停止' }}
            </n-tag>
          </div>
          <div class="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-12 text-gray-400">
            <span>房间号 <span class="text-gray-700 font-medium dark:text-gray-300">{{ listenerStatus.roomId }}</span></span>
            <span>主播 UID <span class="text-gray-700 font-medium dark:text-gray-300">{{ listenerStatus.uid }}</span></span>
            <span>人气 <span class="text-gray-700 font-medium dark:text-gray-300">{{ formatNumber(listenerStatus.online) }}</span></span>
            <span>关注 <span class="text-gray-700 font-medium dark:text-gray-300">{{ formatNumber(listenerStatus.attention) }}</span></span>
            <span v-if="listenerStatus.liveTime">开播 {{ listenerStatus.liveTime }}</span>
            <span v-if="isRunning" class="text-gray-400">｜ 监听运行 {{ listenerStatus.uptime || '-' }}</span>
          </div>
          <div class="grid grid-cols-3 mb-3 gap-3">
            <div class="card-border rounded-6 auto-bg-highlight p-3 text-center">
              <div class="text-16 text-primary font-semibold">
                {{ listenerStatus.msgCount ?? 0 }}
              </div>
              <div class="mt-1 text-12 text-gray-400">
                消息总数
              </div>
            </div>
            <div class="card-border rounded-6 auto-bg-highlight p-3 text-center">
              <div class="text-16 text-[#2080f0] font-semibold">
                {{ listenerStatus.danmuCount ?? 0 }}
              </div>
              <div class="mt-1 text-12 text-gray-400">
                弹幕
              </div>
            </div>
            <div class="card-border rounded-6 auto-bg-highlight p-3 text-center">
              <div class="text-16 text-[#18a058] font-semibold">
                {{ listenerStatus.giftCount ?? 0 }}
              </div>
              <div class="mt-1 text-12 text-gray-400">
                礼物
              </div>
            </div>
          </div>
          <div v-if="isRunning" class="mb-3 text-12 text-gray-400">
            启动于 {{ listenerStatus.startTime || '-' }}
          </div>
          <div class="flex items-center gap-2">
            <n-button v-if="!isRunning" size="small" type="primary" :loading="startLoading" @click="handleStart">
              <i class="i-material-symbols:play-arrow mr-1" />
              启动监听
            </n-button>
            <n-button v-else size="small" type="warning" :loading="stopLoading" @click="handleStop">
              <i class="i-material-symbols:stop mr-1" />
              停止监听
            </n-button>
            <n-button v-if="!isRunning" size="small" :loading="roomBindLoading" @click="showRoomEdit = true">
              <i class="i-material-symbols:edit-outline mr-1" />
              修改直播间
            </n-button>
          </div>
          <div v-if="showRoomEdit" class="mt-3 flex items-center gap-3">
            <n-input v-model:value="roomIdInput" placeholder="请输入新的直播间 ID" size="small" :disabled="roomBindLoading" style="max-width: 240px" />
            <n-button size="small" type="primary" :loading="roomBindLoading" @click="handleRoomUpdate">
              确认修改
            </n-button>
            <n-button size="small" @click="showRoomEdit = false">
              取消
            </n-button>
          </div>
        </template>
      </n-card>
    </div>
    <n-card v-else title="登录管理" size="small" class="mb-4">
      <div class="f-c-c py-8">
        <n-button type="primary" size="large" :loading="loginLoading" @click="openLoginModal">
          <i class="i-material-symbols:login mr-2" />
          扫码登录 B站 账号
        </n-button>
      </div>
    </n-card>
    <n-card v-if="isRunning" size="small" class="mb-4 flex-1 overflow-hidden" :bordered="false">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="h-2 w-2 animate-pulse rounded-full bg-green-500" />
            <span class="text-15 font-medium">实时消息</span>
          </div>
          <span class="text-12 text-gray-400">
            {{ messages.length }} 条
          </span>
        </div>
      </template>
      <!-- 消息区域 -->
      <div ref="chatContainerRef" class="cus-scroll mb-3 h-320 overflow-y-auto rounded-8 bg-gray-50/60 p-3 dark:bg-gray-900/40">
        <div v-if="messages.length === 0" class="h-full flex flex-col items-center justify-center gap-2 text-14 text-gray-400">
          <div class="i-carbon-chat text-32 opacity-40" />
          <div>
            暂无消息，等待直播数据...
          </div>
        </div>
        <div v-for="(msg, idx) in messages" :key="idx" class="mb-2">
          <!-- 普通弹幕 -->
          <template v-if="msg.cmd === 'DANMU_MSG'">
            <div class="mb-5 rounded-8 bg-white px-3 py-2 transition dark:bg-gray-900 hover:bg-gray-50 dark:hover:bg-gray-800">
              <div class="flex items-center gap-2 overflow-hidden whitespace-nowrap">
                <span v-if="msg.data.badge_name" class="mr-5 inline-flex shrink-0 items-center border rounded-4 px-8 py-2 text-10 font-medium" :class="getBadgeClass(msg.data.badge_type)">
                  {{ msg.data.badge_name }}
                  Lv{{ msg.data.badge_level }}
                </span>
                <a class="shrink-0 font-medium no-underline transition hover:underline" :class="getUserNameClass(msg.data.badge_type)" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                  {{ msg.data.username }}
                </a>
                <span class="ml-2 mr-4 shrink-0 text-gray-400">:</span>
                <span class="min-w-0 overflow-hidden text-ellipsis text-14 text-gray-700 dark:text-gray-200">
                  {{ msg.data.content }}
                </span>
              </div>
            </div>
          </template>
          <!-- 礼物 -->
          <template v-else-if="msg.cmd === 'SEND_GIFT' || msg.cmd === 'SEND_GIFT_V2'">
            <div class="mb-2 rounded-8 bg-orange-50 px-3 py-2 transition dark:bg-orange-900/10 hover:bg-orange-100/60 dark:hover:bg-orange-900/20">
              <div class="flex items-center gap-2 overflow-hidden whitespace-nowrap">
                <span v-if="msg.data.badge_name" class="mr-5 inline-flex shrink-0 items-center rounded-4 px-8 py-2 text-10 font-medium" :class="getBadgeClass(msg.data.badge_type)">
                  {{ msg.data.badge_name }}
                  Lv{{ msg.data.badge_level }}
                </span>
                <a class="shrink-0 font-medium no-underline transition hover:underline" :class="getUserNameClass(msg.data.badge_type)" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                  {{ msg.data.username }}
                </a>
                <span class="ml-3 shrink-0 text-gray-500 dark:text-gray-400">
                  送出
                </span>
                <span class="min-w-0 overflow-hidden text-ellipsis text-orange-600 font-medium dark:text-orange-400">
                  {{ msg.data.gift_name }} x {{ msg.data.num }}
                </span>
                <span v-if="msg.data.blind_gift" class="ml-5 min-w-0 overflow-hidden text-ellipsis text-10 text-gray-600 font-medium dark:text-gray-400">
                  {{ msg.data.blind_gift.original_gift_name }} - {{ blindGiftAnalyze(msg.data.num, msg.data.price, msg.data.blind_gift.original_gift_price) }}
                </span>
                <span class="ml-auto shrink-0 rounded-4 bg-orange-100 px-8 py-2 text-12 text-orange-600 font-medium dark:bg-orange-900/30 dark:text-orange-300">
                  ¥{{ formatPrice(msg.data.price * msg.data.num) }}
                </span>
              </div>
            </div>
          </template>
          <!-- 醒目留言 -->
          <template v-else-if="msg.cmd === 'SUPER_CHAT_MESSAGE'">
            <div class="mb-5 border border-orange-200 rounded-4 bg-orange-50 px-3 py-3 dark:border-orange-800 dark:bg-orange-950/30">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 flex items-center gap-2 overflow-hidden">
                  <span v-if="msg.data.badge_name" class="mr-3 inline-flex shrink-0 items-center border rounded-4 px-8 py-2 text-10 font-medium" :class="getBadgeClass(msg.data.badge_type)">
                    {{ msg.data.badge_name }}
                    Lv{{ msg.data.badge_level }}
                  </span>
                  <a class="shrink-0 font-medium no-underline transition hover:underline" :class="getUserNameClass(msg.data.badge_type)" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                    {{ msg.data.username }}
                  </a>
                </div>
                <span class="shrink-0 text-orange-600 font-bold dark:text-orange-400">
                  ¥{{ formatPrice(msg.data.price) }}
                </span>
              </div>
              <div class="mt-3 border border-orange-100 rounded-6 bg-white px-3 py-2 text-14 text-gray-700 leading-normal dark:border-orange-900 dark:bg-gray-900/50 dark:text-gray-200">
                {{ msg.data.message }}
              </div>
            </div>
          </template>
          <!-- 舰长开通 -->
          <template v-else-if="msg.cmd === 'GUARD_BUY'">
            <div class="mb-5 rounded-8 px-3 py-3 transition" :class="getGuardCardClass(msg.data.guard_level)">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 overflow-hidden">
                  <span class="shrink-0 rounded-full px-8 py-2 text-11 font-medium" :class="getGuardBadgeClass(msg.data.guard_level)">
                    {{ getGuardName(msg.data.guard_level) }}
                  </span>
                  <a class="shrink-0 font-medium no-underline hover:underline" :class="getGuardUserClass(msg.data.guard_level)" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                    {{ msg.data.username }}
                  </a>
                  <span class="text-13 text-gray-500 dark:text-gray-400">
                    开通了
                  </span>
                  <span class="font-medium" :class="getGuardUserClass(msg.data.guard_level)">
                    {{ msg.data.gift_name }}
                  </span>
                  <span v-if="msg.data.num > 1" class="text-13 text-gray-500 dark:text-gray-400">
                    x {{ msg.data.num }}
                  </span>
                </div>
                <span v-if="msg.data.price" class="shrink-0 text-13 font-medium" :class="getUserNameClass(msg.data.guard_level)">
                  ¥{{ formatPrice(msg.data.price * msg.data.num) }}
                </span>
              </div>
            </div>
          </template>
          <!-- 互动事件 -->
          <template v-else-if="msg.cmd === 'INTERACT_WORD_V2'">
            <div class="mb-5 flex items-center gap-2 rounded-8 px-3 py-2 text-13 text-gray-500 dark:text-gray-400">
              <span class="mr-5 h-20 w-20 inline-flex shrink-0 items-center justify-center rounded-full bg-gray-100 text-gray-400 dark:bg-gray-800">
                <span v-if="msg.data.msg_type === 1" class="i-carbon-login text-12" />
                <span v-else-if="msg.data.msg_type === 2" class="i-carbon-favorite text-12" />
                <span v-else-if="msg.data.msg_type === 3" class="i-carbon-share text-12" />
              </span>
              <span v-if="msg.data.badge_name" class="mr-5 inline-flex shrink-0 items-center border rounded-4 px-8 py-2 text-10 font-medium" :class="getBadgeClass(msg.data.badge_type)">
                {{ msg.data.badge_name }}
                Lv{{ msg.data.badge_level }}
              </span>
              <a class="shrink-0 font-medium no-underline transition hover:underline" :class="getUserNameClass(msg.data.badge_type)" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                {{ msg.data.username }}
              </a>
              <span class="shrink-0">
                <template v-if="msg.data.msg_type === 1">
                  进入直播间
                </template>
                <template v-else-if="msg.data.msg_type === 2">
                  关注了直播间
                </template>
                <template v-else-if="msg.data.msg_type === 3">
                  分享了直播间
                </template>
              </span>
            </div>
          </template>
          <!-- PK -->
          <template v-else-if="msg.cmd === 'PK_BATTLE_PRE_NEW'">
            <div class="mb-5 border border-indigo-200 rounded-8 bg-indigo-50 px-3 py-2 dark:border-indigo-900 dark:bg-indigo-950/30">
              <div class="flex items-center gap-2 text-13">
                <span class="shrink-0 text-indigo-600 font-medium dark:text-indigo-300">
                  PK 对决
                </span>
                <span class="text-gray-400"> | </span>
                <span class="text-gray-500 dark:text-gray-400">
                  即将开始
                </span>
              </div>
              <div class="mt-2 flex items-center gap-2 text-14">
                <a class="text-gray-700 font-medium no-underline dark:text-gray-200 hover:underline" :href="`https://space.bilibili.com/${msg.data.uid}`" target="_blank">
                  {{ msg.data.username }}
                </a>
                <span class="text-gray-400">
                  ·
                </span>
                <a class="text-gray-500 no-underline dark:text-gray-400 hover:text-indigo-500 hover:underline" :href="`https://live.bilibili.com/${msg.data.room_id}`" target="_blank">
                  房间 {{ msg.data.room_id }}
                </a>
              </div>
            </div>
          </template>
        </div>
      </div>
      <!-- 输入区域 -->
      <div class="flex gap-2 rounded-8 bg-gray-50 p-2 dark:bg-gray-900/40">
        <n-input v-model:value="danmuText" placeholder="输入弹幕内容..." size="small" :maxlength="40" show-count :disabled="sendLoading" class="flex-1" @keyup.enter="handleSendDanmu" />
        <n-button size="small" type="primary" :loading="sendLoading" :disabled="!danmuText.trim()" @click="handleSendDanmu">
          发送
        </n-button>
      </div>
    </n-card>
    <!-- 登录弹窗 -->
    <n-modal v-model:show="showLoginModal" title="扫码登录 B站" preset="card" style="width: 440px" :mask-closable="false" @after-leave="clearPollTimer">
      <div class="f-c-c flex-col py-4">
        <n-spin v-if="qrLoading" size="medium" />
        <div v-else-if="qrcodeUrl" class="f-c-c flex-col">
          <div class="rounded-8 bg-white p-4">
            <QrcodeVue :value="qrcodeUrl" :size="200" level="M" />
          </div>
          <div class="mt-12 f-c-c gap-2">
            <n-tag v-if="qrPollStatus === 'waiting'" type="info" size="large">
              <i class="i-material-symbols:qr-code-scanner mr-2" />
              等待扫描
            </n-tag>
            <n-tag v-else-if="qrPollStatus === 'scanned'" type="warning" size="large">
              <i class="i-material-symbols:check-circle-outline mr-2" />
              已扫描，请在手机上确认
            </n-tag>
            <n-tag v-else-if="qrPollStatus === 'expired'" type="error" size="large">
              <i class="i-material-symbols:error-outline mr-2" />
              二维码已过期
            </n-tag>
            <n-tag v-else-if="qrPollStatus === 'success'" type="success" size="large">
              <i class="i-material-symbols:check-circle mr-2" />
              登录成功
            </n-tag>
            <p v-else class="mt-2 text-14 text-gray-400">
              {{ qrMessage }}
            </p>
          </div>
          <n-button v-if="qrPollStatus === 'expired'" type="primary" class="mt-12" @click="refreshQrcode">
            <i class="i-material-symbols:refresh mr-2" />
            刷新二维码
          </n-button>
        </div>
        <div v-else class="f-c-c flex-col py-6">
          <p class="mb-4 text-14 text-gray-400">
            获取二维码失败
          </p>
          <n-button type="primary" @click="refreshQrcode">
            重试
          </n-button>
        </div>
      </div>
    </n-modal>
  </AppPage>
</template>

<script setup>
import QrcodeVue from 'qrcode.vue'
import { useAuthStore } from '@/store'
import api, { buildWsUrl } from './api'

// 登录状态
const loginLoading = ref(false)
const loginStatus = ref({
  isLoggedIn: false, // 是否登录
  uid: 0, // 用户uid
  username: '', // 用户昵称
  face: '', // 头像URL
  buvid: '', // buvid
})

// 登录二维码弹窗
const qrLoading = ref(false)
const showLoginModal = ref(false)
const qrcodeUrl = ref('')
const qrcodeKey = ref('')
const qrPollStatus = ref('waiting') // 'waiting' | 'scanned' | 'expired' | 'success'
const qrMessage = ref('')
let qrPollTimer = null

// 直播间状态
const listenerLoading = ref(false)
const roomBindLoading = ref(false)
const startLoading = ref(false)
const stopLoading = ref(false)
const showRoomEdit = ref(false)
const listenerStatus = ref({
  isRunning: false, // 是否正在监听
  roomId: 0, // 房间号
  startTime: '', // 开始监听时间
  uptime: '', // 已监听时长
  msgCount: 0, // 监听到消息数量
  danmuCount: 0, // 监听到弹幕数量
  giftCount: 0, // 监听到礼物数量
  uid: 0, // 直播间主播ID
  title: '', // 直播间标题
  liveStatus: 0, // 直播状态：0=未开播, 1=直播中, 2=轮播中
  online: 0, // 在线观众数（人气值，非真实人数）
  attention: 0, // 关注数
  liveTime: '', // 开播时间
})
const roomIdInput = ref('')

// WebSocket 消息状态
const MAX_MESSAGES = 500
const messages = ref([])
const messagesShowType = ref({
  DANMU_MSG: true,
  SEND_GIFT: true,
  SEND_GIFT_V2: true,
  SUPER_CHAT_MESSAGE: true,
  GUARD_BUY: true,
  INTERACT_WORD_V2: true,
  PK_BATTLE_PRE_NEW: true,
})
const chatContainerRef = ref(null)
const danmuText = ref('')
const sendLoading = ref(false)

// WebSocket 连接信息
let wsClient = null
let reconnectTimer = null
const logoutLoading = ref(false)

// 计算属性
const isLoggedIn = computed(() => loginStatus.value?.isLoggedIn === true)
const hasRoom = computed(() => !!listenerStatus.value?.roomId)
const isRunning = computed(() => listenerStatus.value?.isRunning === true)
const liveStatusLabel = computed(() => {
  const map = { 0: '未开播', 1: '直播中', 2: '轮播中' }
  return map[listenerStatus.value?.liveStatus] || '未知'
})
const liveStatusTagType = computed(() => {
  const map = { 0: 'default', 1: 'success', 2: 'info' }
  return map[listenerStatus.value?.liveStatus] || 'default'
})

// 登录模块 - 获取登录状态
async function fetchLoginStatus() {
  loginLoading.value = true
  try {
    const res = await api.getLoginStatus()
    if (res.code === 0) {
      Object.assign(loginStatus.value, res.data)
    }
  }
  catch {
    loginStatus = ref({
      isLoggedIn: false,
      uid: 0,
      username: '',
      face: '',
      buvid: '',
    })
  }
  finally {
    loginLoading.value = false
  }
}

// 登录模块 - 打开扫码登录弹窗
function openLoginModal() {
  showLoginModal.value = true
  fetchQRCode()
}

// 登录模块 - 获取登录二维码信息
async function fetchQRCode() {
  qrLoading.value = true
  qrPollStatus.value = 'waiting'
  qrMessage.value = ''
  try {
    const res = await api.getQRCode()
    if (res.code === 0) {
      qrcodeUrl.value = res.data.url
      qrcodeKey.value = res.data.qrcodeKey
      if (qrcodeKey.value) {
        startPolling()
      }
    }
  }
  catch {
    qrcodeUrl.value = ''
    qrcodeKey.value = ''
  }
  finally {
    qrLoading.value = false
  }
}

// 登录模块 - 轮询二维码登录状态
function startPolling() {
  clearPollTimer()
  qrPollTimer = setInterval(async () => {
    try {
      const res = await api.pollQRCode(qrcodeKey.value)
      if (res.code === 0) {
        if (res.data.isSuccess) {
          qrPollStatus.value = 'success'
          clearPollTimer()
          // 延迟关闭弹窗，让用户看到成功状态
          setTimeout(() => {
            showLoginModal.value = false
            refreshAfterLogin()
          }, 1000)
        }
        else if (res.data.isExpired) {
          qrPollStatus.value = 'expired'
          qrMessage.value = res.data.message || ''
          clearPollTimer()
        }
        else if (res.data.isScanned) {
          qrPollStatus.value = 'scanned'
          qrMessage.value = res.data.message || ''
        }
        else {
          qrPollStatus.value = 'waiting'
          qrMessage.value = res.data.message || ''
        }
      }
    }
    catch {
      // 轮询失败不中断，继续尝试
    }
  }, 4000)
}

// 登录模块 - 重新获取登录二维码
function refreshQrcode() {
  clearPollTimer()
  fetchQRCode()
}

// 登录模块 - 清空二维码登录状态轮询
function clearPollTimer() {
  if (qrPollTimer) {
    clearInterval(qrPollTimer)
    qrPollTimer = null
  }
}

// 登录模块 - 刷新登录状态
async function refreshAfterLogin() {
  await fetchLoginStatus()
  if (isLoggedIn.value) {
    await fetchListenerStatus()
  }
}

// 登录模块 - 退出登录
async function handleLogout() {
  logoutLoading.value = true
  try {
    await api.logout()
    disconnectWebSocket()
    loginStatus.value = null
    listenerStatus.value = null
    messages.value = []
    showRoomEdit.value = false
    roomIdInput.value = ''
    $message.success('已退出登录')
  }
  finally {
    logoutLoading.value = false
  }
}

// 直播间模块 - 获取直播间状态
async function fetchListenerStatus() {
  if (!isLoggedIn.value)
    return
  listenerLoading.value = true
  try {
    const res = await api.getListenerStatus()
    if (res.code === 0) {
      const wasRunning = listenerStatus.value?.isRunning
      Object.assign(listenerStatus.value, res.data)
      // 状态变化时管理 WebSocket
      if (res.data.isRunning && !wasRunning) {
        connectWebSocket()
      }
      else if (!res.data.isRunning && wasRunning) {
        disconnectWebSocket()
      }
    }
  }
  catch {
    listenerStatus = ref({
      isRunning: false,
      roomId: 0,
      startTime: '',
      uptime: '',
      msgCount: 0,
      danmuCount: 0,
      giftCount: 0,
      uid: 0,
      title: '',
      liveStatus: 0,
      online: 0,
      attention: 0,
      liveTime: '',
    })
    disconnectWebSocket()
  }
  finally {
    listenerLoading.value = false
  }
}

// 直播间模块 - 绑定直播间
async function handleRoomUpdate() {
  const id = Number.parseInt(roomIdInput.value, 10)
  if (!id || id < 1) {
    $message.warning('请输入有效的直播间 ID')
    return
  }
  roomBindLoading.value = true
  try {
    const res = await api.updateRoom(id)
    if (res.code === 0) {
      $message.success('直播间绑定成功')
      roomIdInput.value = ''
      showRoomEdit.value = false
      await fetchListenerStatus()
    }
  }
  finally {
    roomBindLoading.value = false
  }
}

// 直播间模块 - 启动监听
async function handleStart() {
  startLoading.value = true
  try {
    const res = await api.startListener()
    if (res.code === 0) {
      $message.success('监听已启动')
      await fetchListenerStatus()
    }
  }
  finally {
    startLoading.value = false
  }
}

// 直播间模块 - 停止监听
async function handleStop() {
  stopLoading.value = true
  try {
    const res = await api.stopListener()
    if (res.code === 0) {
      $message.success('监听已停止')
      disconnectWebSocket()
      await fetchListenerStatus()
    }
  }
  finally {
    stopLoading.value = false
  }
}

// 消息模块 - 连接消息流
function connectWebSocket() {
  if (!listenerStatus.value?.isRunning)
    return
  if (wsClient?.readyState === WebSocket.OPEN)
    return
  clearTimeout(reconnectTimer)
  const token = useAuthStore().accessToken
  if (!token)
    return
  const url = buildWsUrl(token)
  wsClient = new WebSocket(url)
  wsClient.onopen = () => {
    console.warn('[Live] WebSocket 已连接')
  }
  wsClient.onmessage = (event) => {
    try {
      const { cmd, data } = JSON.parse(event.data)
      if (messagesShowType.value?.[cmd] === true) {
        messages.value.push({
          cmd,
          data,
          timestamp: Date.now(),
        })
        // FIFO 内存控制
        if (messages.value.length > MAX_MESSAGES) {
          messages.value.splice(0, messages.value.length - MAX_MESSAGES)
        }
        // 自动滚动到底部
        nextTick(() => scrollToBottom())
      }
    }
    catch (e) {
      console.error('[Live] WebSocket 消息解析失败:', e)
    }
  }
  wsClient.onerror = (e) => {
    console.error('[Live] WebSocket 错误:', e)
  }
  wsClient.onclose = () => {
    console.warn('[Live] WebSocket 已断开')
    wsClient = null
    // 非预期断开时自动重连
    if (listenerStatus.value?.isRunning) {
      reconnectTimer = setTimeout(() => {
        connectWebSocket()
      }, 3000)
    }
  }
}

// 消息模块 - 关闭消息流
function disconnectWebSocket() {
  clearTimeout(reconnectTimer)
  if (wsClient) {
    wsClient.close()
    wsClient = null
  }
}

// 消息模块 - 发送弹幕
async function handleSendDanmu() {
  const text = danmuText.value.trim()
  if (!text)
    return
  if (text.length > 40) {
    $message.warning('弹幕内容不能超过 40 个字符')
    return
  }
  sendLoading.value = true
  try {
    const res = await api.sendDanmu(text)
    if (res.code === 0) {
      danmuText.value = ''
      $message.success('弹幕发送成功')
    }
  }
  finally {
    sendLoading.value = false
  }
}

// 辅助方法 - 数字转换
function formatNumber(num) {
  if (num == null)
    return '-'
  if (num >= 10000)
    return `${(num / 10000).toFixed(1)}万`
  return String(num)
}

// 辅助方法 - 滚动到底部
function scrollToBottom() {
  const el = chatContainerRef.value
  if (!el)
    return
  const threshold = 50
  const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - threshold
  if (nearBottom) {
    el.scrollTop = el.scrollHeight
  }
}

// 辅助方法 - 格式化金额
function formatPrice(price) {
  return (price / 100).toFixed(2)
}

// 辅助方法 - 获取用户名称颜色
function getUserNameClass(type) {
  switch (type) {
    case 1: // 总督
      return 'text-yellow-600 dark:text-yellow-400'
    case 2: // 提督
      return 'text-purple-600 dark:text-purple-400'
    case 3: // 舰长
      return 'text-blue-600 dark:text-blue-400'
    default: // 普通用户
      return 'text-gray-600 dark:text-gray-300'
  }
}

// 辅助方法 - 获取牌子颜色
function getBadgeClass(type) {
  switch (type) {
    case 1: // 总督
      return `
        border-yellow-300
        bg-yellow-50
        text-yellow-700
        dark:border-yellow-700
        dark:bg-yellow-900/30
        dark:text-yellow-300
      `
    case 2: // 提督
      return `
        border-purple-300
        bg-purple-50
        text-purple-700
        dark:border-purple-700
        dark:bg-purple-900/30
        dark:text-purple-300
      `
    case 3: // 舰长
      return `
        border-blue-300
        bg-blue-50
        text-blue-700
        dark:border-blue-700
        dark:bg-blue-900/30
        dark:text-blue-300
      `
    default: // 普通用户
      return `
        border-gray-300
        bg-gray-50
        text-gray-600
        dark:border-gray-700
        dark:bg-gray-800
        dark:text-gray-300
      `
  }
}

// 辅助方法 - 盲盒盈亏分析
function blindGiftAnalyze(num, price, original_gift_price) {
  if (price > original_gift_price) {
    return `赚了 ${(((price - original_gift_price) * num) / 100).toFixed(2)} 元`
  }
  else {
    return `亏了 ${(((original_gift_price - price) * num) / 100).toFixed(2)} 元`
  }
}

// 辅助方法 - 用户身份名称
function getGuardName(level) {
  switch (level) {
    case 1:
      return '总督'
    case 2:
      return '提督'
    case 3:
      return '舰长'
    default:
      return '航海'
  }
}

// 辅助方法 - 获取用户身份背景
function getGuardCardClass(level) {
  switch (level) {
    case 1: // 总督
      return `
        bg-yellow-50
        dark:bg-yellow-900/20
      `
    case 2: // 提督
      return `
        bg-purple-50
        dark:bg-purple-900/20
      `
    case 3: // 舰长
      return `
        bg-blue-50
        dark:bg-blue-900/20
      `
    default: // 普通用户
      return `
        bg-gray-50
        dark:bg-gray-900
      `
  }
}

// 辅助方法 - 获取用户开通舰长模块的牌子背景色
function getGuardBadgeClass(level) {
  switch (level) {
    case 1: // 总督
      return `
        bg-yellow-100
        text-yellow-700
        dark:bg-yellow-800/40
        dark:text-yellow-300
      `
    case 2: // 提督
      return `
        bg-purple-100
        text-purple-700
        dark:bg-purple-800/40
        dark:text-purple-300
      `
    case 3: // 舰长
      return `
        bg-blue-100
        text-blue-700
        dark:bg-blue-800/40
        dark:text-blue-300
      `
  }
}

// 辅助方法 - 获取用户开通舰长模块的文字颜色
function getGuardUserClass(level) {
  switch (level) {
    case 1: // 总督
      return `
        text-yellow-700
        dark:text-yellow-300
      `
    case 2: // 提督
      return `
        text-purple-700
        dark:text-purple-300
      `
    case 3: // 舰长
      return `
        text-blue-700
        dark:text-blue-300
      `
  }
}

// 生命周期
onMounted(async () => {
  await fetchLoginStatus()
  if (isLoggedIn.value) {
    await fetchListenerStatus()
  }
})

onUnmounted(() => {
  clearPollTimer()
  disconnectWebSocket()
})
</script>
