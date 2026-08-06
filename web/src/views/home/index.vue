<template>
  <AppPage show-footer full>
    <!-- 已登录：登录信息 + 直播间 双栏布局 -->
    <div v-if="isLoggedIn" class="mb-4 flex gap-4">
      <!-- 登录信息卡片 -->
      <n-card size="small" :bordered="false" class="min-w-200 w-auto flex-shrink-0">
        <div class="flex flex-col items-center">
          <!-- 头像 -->
          <div class="relative">
            <n-avatar round :size="64" :src="loginStatus.face" :bordered="true" />

            <!-- 在线状态 -->
            <div class="absolute bottom-1 right-1 h-3 w-3 rounded-full bg-green-500 ring-2 ring-white" />
          </div>

          <!-- 用户信息 -->
          <div class="mt-3 w-full text-center">
            <div class="truncate text-15 font-medium">
              {{ loginStatus.username }}
            </div>

            <div class="mt-1 text-12 text-gray-400">
              UID {{ loginStatus.uid }}
            </div>
          </div>

          <!-- 退出 -->
          <n-button
            secondary type="error" size="small" class="mt-4 w-full" :loading="logoutLoading"
            @click="handleLogout"
          >
            退出登录
          </n-button>
        </div>
      </n-card>
      <!-- 直播间管理卡片 -->
      <n-card title="直播间管理" :bordered="false" size="small" class="min-w-300 flex-1">
        <!-- 加载中 -->
        <div v-if="listenerLoading && !listenerStatus" class="f-c-c py-6">
          <n-spin size="small" />
        </div>

        <!-- 未绑定房间 -->
        <div v-else-if="!hasRoom" class="flex items-center gap-3">
          <n-input
            v-model:value="roomIdInput" placeholder="请输入直播间 ID" size="small" :disabled="roomBindLoading"
            style="max-width: 240px" @keyup.enter="handleRoomUpdate"
          />
          <n-button size="small" type="primary" :loading="roomBindLoading" @click="handleRoomUpdate">
            确认绑定
          </n-button>
        </div>

        <!-- 已绑定房间 -->
        <template v-else>
          <!-- 直播间标题 + 直播状态 -->
          <div class="mb-3 flex items-center gap-2">
            <span class="truncate text-16 font-semibold">{{ listenerStatus.title || '直播间' }}</span>
            <n-tag :type="liveStatusTagType" size="small" round>
              {{ liveStatusLabel }}
            </n-tag>
            <n-tag :type="isRunning ? 'success' : 'default'" size="small" round>
              <template #icon>
                <i
                  :class="isRunning ? 'i-material-symbols:play-circle-outline' : 'i-material-symbols:pause-circle-outline'"
                />
              </template>
              {{ isRunning ? '监听中' : '已停止' }}
            </n-tag>
          </div>

          <!-- 直播间基本信息 -->
          <div class="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-12 text-gray-400">
            <span>房间号 <span class="text-gray-700 font-medium dark:text-gray-300">{{ listenerStatus.roomId
            }}</span></span>
            <span>主播 UID <span class="text-gray-700 font-medium dark:text-gray-300">{{ listenerStatus.uid
            }}</span></span>
            <span>人气 <span class="text-gray-700 font-medium dark:text-gray-300">{{ formatNumber(listenerStatus.online)
            }}</span></span>
            <span>关注 <span class="text-gray-700 font-medium dark:text-gray-300">{{
              formatNumber(listenerStatus.attention)
            }}</span></span>
            <span v-if="listenerStatus.liveTime">开播 {{ listenerStatus.liveTime }}</span>
            <span v-if="isRunning" class="text-gray-400">｜ 监听运行 {{ listenerStatus.uptime || '-' }}</span>
          </div>
          <!-- 数据统计卡片 -->
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
          <!-- 运行详情 -->
          <div v-if="isRunning" class="mb-3 text-12 text-gray-400">
            启动于 {{ listenerStatus.startTime || '-' }}
          </div>
          <!-- 操作按钮 -->
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

          <!-- 修改直播间 -->
          <div v-if="showRoomEdit" class="mt-3 flex items-center gap-3">
            <n-input
              v-model:value="roomIdInput" placeholder="请输入新的直播间 ID" size="small" :disabled="roomBindLoading"
              style="max-width: 240px"
            />
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

    <!-- 未登录：独立登录卡片 -->
    <n-card v-else title="登录管理" size="small" class="mb-4">
      <div class="f-c-c py-8">
        <n-button type="primary" size="large" :loading="loginLoading" @click="openLoginModal">
          <i class="i-material-symbols:login mr-2" />
          扫码登录 B站 账号
        </n-button>
      </div>
    </n-card>

    <!-- 消息模块（全宽，监听运行后可见） -->
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
      <div
        ref="chatContainerRef" class="cus-scroll mb-3 h-320 overflow-y-auto rounded-8 bg-gray-50/60 p-3 dark:bg-gray-900/40"
      >
        <div
          v-if="messages.length === 0" class="h-full flex flex-col items-center justify-center gap-2 text-14 text-gray-400"
        >
          <div class="i-carbon-chat text-32 opacity-40" />
          <div>
            暂无消息，等待直播数据...
          </div>
        </div>

        <div
          v-for="(msg, idx) in messages" :key="idx" class="mb-2 rounded-6 px-3 py-2 transition hover:bg-white dark:hover:bg-gray-800"
        />
      </div>

      <!-- 输入区域 -->
      <div
        class="flex gap-2 rounded-8 bg-gray-50 p-2 dark:bg-gray-900/40"
      >
        <n-input
          v-model:value="danmuText" placeholder="输入弹幕内容..." size="small" :maxlength="40" show-count
          :disabled="sendLoading" class="flex-1" @keyup.enter="handleSendDanmu"
        />

        <n-button
          size="small" type="primary" :loading="sendLoading" :disabled="!danmuText.trim()"
          @click="handleSendDanmu"
        >
          发送
        </n-button>
      </div>
    </n-card>

    <!-- 登录弹窗 -->
    <n-modal
      v-model:show="showLoginModal" title="扫码登录 B站" preset="card" style="width: 440px" :mask-closable="false"
      @after-leave="clearPollTimer"
    >
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
const loginStatus = ref({
  isLoggedIn: false,
  uid: 0,
  username: '',
  face: '',
  buvid: '',
})
const loginLoading = ref(false)

// QR 弹窗状态
const showLoginModal = ref(false)
const qrcodeUrl = ref('')
const qrcodeKey = ref('')
const qrPollStatus = ref('waiting') // 'waiting' | 'scanned' | 'expired' | 'success'
const qrMessage = ref('')
const qrLoading = ref(false)
let qrPollTimer = null

// ==================== 直播间状态 ====================
const listenerStatus = ref({
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
const listenerLoading = ref(false)
const roomIdInput = ref('')
const roomBindLoading = ref(false)
const startLoading = ref(false)
const stopLoading = ref(false)
const showRoomEdit = ref(false)

// ==================== 消息状态 ====================
const messages = ref([])
const danmuText = ref('')
const sendLoading = ref(false)
const chatContainerRef = ref(null)
const MAX_MESSAGES = 500

// ==================== WebSocket ====================
let wsClient = null
let reconnectTimer = null
const logoutLoading = ref(false)

// ==================== 计算属性 ====================
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
