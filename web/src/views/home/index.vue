<template>
  <AppPage show-footer>
    <!-- 已登录：登录信息 + 直播间 双栏布局 -->
    <div v-if="isLoggedIn" class="mb-4 flex gap-4">
      <!-- 登录信息卡片 -->
      <n-card title="登录信息" size="small" class="w-240 flex-shrink-0">
        <div class="flex items-center gap-3">
          <n-avatar round :size="36" class="flex-shrink-0 bg-primary">
            <span class="text-14 text-white font-bold">
              {{ loginStatus.username?.charAt(0) || 'U' }}
            </span>
          </n-avatar>
          <div class="min-w-0 flex-1">
            <div class="truncate text-14 font-medium">
              {{ loginStatus.username }}
            </div>
            <div class="mt-1 text-12 text-gray-400">
              UID {{ loginStatus.uid }}
            </div>
          </div>
        </div>
        <n-button
          type="error"
          size="small"
          text
          class="mt-3"
          :loading="logoutLoading"
          @click="handleLogout"
        >
          退出登录
        </n-button>
      </n-card>

      <!-- 直播间管理卡片 -->
      <n-card title="直播间管理" size="small" class="flex-1">
        <!-- 加载中 -->
        <div v-if="listenerLoading && !listenerStatus" class="f-c-c py-6">
          <n-spin size="small" />
        </div>

        <!-- 未绑定房间 -->
        <div v-else-if="!hasRoom" class="flex items-center gap-3">
          <n-input
            v-model:value="roomIdInput"
            placeholder="请输入直播间 ID"
            size="small"
            :disabled="roomBindLoading"
            style="max-width: 240px"
            @keyup.enter="handleRoomUpdate"
          />
          <n-button size="small" type="primary" :loading="roomBindLoading" @click="handleRoomUpdate">
            确认绑定
          </n-button>
        </div>

        <!-- 已绑定房间 -->
        <template v-else>
          <!-- 房间信息行 -->
          <div class="mb-3 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-14 text-gray-400">直播间</span>
              <span class="text-16 font-semibold">{{ listenerStatus.roomId }}</span>
              <n-tag :type="isRunning ? 'success' : 'default'" size="small" round>
                <template #icon>
                  <i :class="isRunning ? 'i-material-symbols:play-circle-outline' : 'i-material-symbols:pause-circle-outline'" />
                </template>
                {{ isRunning ? '运行中' : '已停止' }}
              </n-tag>
            </div>
            <span v-if="isRunning" class="text-12 text-gray-400">
              已运行 {{ listenerStatus.uptime || '-' }}
            </span>
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
            <n-button
              v-if="!isRunning"
              size="small"
              type="primary"
              :loading="startLoading"
              @click="handleStart"
            >
              <i class="i-material-symbols:play-arrow mr-1" />
              启动监听
            </n-button>
            <n-button
              v-else
              size="small"
              type="warning"
              :loading="stopLoading"
              @click="handleStop"
            >
              <i class="i-material-symbols:stop mr-1" />
              停止监听
            </n-button>
            <n-button
              v-if="!isRunning"
              size="small"
              :loading="roomBindLoading"
              @click="showRoomEdit = true"
            >
              <i class="i-material-symbols:edit-outline mr-1" />
              修改直播间
            </n-button>
          </div>

          <!-- 修改直播间 -->
          <div v-if="showRoomEdit" class="mt-3 flex items-center gap-3">
            <n-input
              v-model:value="roomIdInput"
              placeholder="请输入新的直播间 ID"
              size="small"
              :disabled="roomBindLoading"
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
    <n-card v-if="isRunning" title="实时消息" size="small" class="mb-4">
      <!-- 消息展示区 -->
      <div
        ref="chatContainerRef"
        class="cus-scroll mb-3 h-320 overflow-y-auto card-border rounded-6 auto-bg p-3"
      >
        <div v-if="messages.length === 0" class="h-full f-c-c text-14 text-gray-400">
          暂无消息，等待直播数据...
        </div>
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="mb-1 flex items-start gap-2"
        >
          <span class="mt-1 flex-shrink-0 text-12 text-gray-400">
            {{ formatMsgTime(msg.timestamp) }}
          </span>
          <n-tag :type="getCmdTagType(msg.cmd)" size="small" class="flex-shrink-0">
            {{ getCmdLabel(msg.cmd) }}
          </n-tag>
          <span class="break-all text-14">{{ formatMsgContent(msg) }}</span>
        </div>
      </div>

      <!-- 发送弹幕 -->
      <div class="flex gap-2">
        <n-input
          v-model:value="danmuText"
          placeholder="输入弹幕内容（1-40 字符）"
          size="small"
          :maxlength="40"
          show-count
          :disabled="sendLoading"
          class="flex-1"
          @keyup.enter="handleSendDanmu"
        />
        <n-button
          size="small"
          type="primary"
          :loading="sendLoading"
          :disabled="!danmuText.trim()"
          @click="handleSendDanmu"
        >
          发送
        </n-button>
      </div>
    </n-card>

    <!-- 登录弹窗 -->
    <n-modal
      v-model:show="showLoginModal"
      title="扫码登录 B站"
      preset="card"
      style="width: 440px"
      :mask-closable="false"
      @after-leave="clearPollTimer"
    >
      <div class="f-c-c flex-col py-4">
        <n-spin v-if="qrLoading" size="medium" />

        <div v-else-if="qrcodeUrl" class="f-c-c flex-col">
          <div class="rounded-8 bg-white p-4">
            <QrcodeVue :value="qrcodeUrl" :size="200" level="M" />
          </div>

          <div class="mt-4 f-c-c gap-2">
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
          </div>

          <p v-if="qrMessage && qrPollStatus !== 'expired'" class="mt-2 text-14 text-gray-400">
            {{ qrMessage }}
          </p>

          <n-button v-if="qrPollStatus === 'expired'" type="primary" class="mt-4" @click="refreshQrcode">
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

// ==================== 登录状态 ====================
const loginStatus = ref(null)
const loginLoading = ref(false)

// ==================== QR 弹窗状态 ====================
const showLoginModal = ref(false)
const qrcodeUrl = ref('')
const qrcodeKey = ref('')
const qrPollStatus = ref('') // 'waiting' | 'scanned' | 'expired' | 'success'
const qrMessage = ref('')
const qrLoading = ref(false)
let qrPollTimer = null

// ==================== 直播间状态 ====================
const listenerStatus = ref(null)
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

// ==================== 消息类型映射 ====================
const CMD_LABELS = {
  DANMU_MSG: '弹幕',
  SEND_GIFT: '礼物',
  INTERACT_WORD: '进房',
  GUARD_BUY: '舰长',
  SUPER_CHAT_MESSAGE: 'SC',
  ENTRY_EFFECT: '进场',
  WATCHED_CHANGE: '看过',
  ROOM_REAL_TIME_MESSAGE_UPDATE: '系统',
}

function getCmdLabel(cmd) {
  return CMD_LABELS[cmd] || cmd || '消息'
}

function getCmdTagType(cmd) {
  switch (cmd) {
    case 'DANMU_MSG': return 'info'
    case 'SEND_GIFT': return 'success'
    case 'INTERACT_WORD': return 'default'
    case 'GUARD_BUY': return 'warning'
    case 'SUPER_CHAT_MESSAGE': return 'error'
    default: return 'default'
  }
}

function formatMsgContent(msg) {
  const { cmd, data } = msg
  if (!data)
    return ''

  switch (cmd) {
    case 'DANMU_MSG':
      return `${data.username || '未知用户'}：${data.content || ''}`
    case 'SEND_GIFT':
      return `${data.username || '未知用户'} 送出 ${data.giftName || '礼物'}${data.num ? ` x${data.num}` : ''}`
    case 'INTERACT_WORD':
      return `${data.username || '未知用户'} 进入了直播间`
    case 'GUARD_BUY':
      return `${data.username || '未知用户'} 开通了 ${data.giftName || '舰长'}`
    case 'SUPER_CHAT_MESSAGE':
      return `${data.username || '未知用户'}：${data.message || ''}`
    default:
      return JSON.stringify(data)
  }
}

function formatMsgTime(ts) {
  if (!ts)
    return ''
  const d = new Date(ts)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

// ==================== 登录相关 ====================
async function fetchLoginStatus() {
  loginLoading.value = true
  try {
    const res = await api.getLoginStatus()
    loginStatus.value = res.data
  }
  catch {
    loginStatus.value = null
  }
  finally {
    loginLoading.value = false
  }
}

function openLoginModal() {
  showLoginModal.value = true
  fetchQRCode()
}

async function fetchQRCode() {
  qrLoading.value = true
  qrPollStatus.value = ''
  qrMessage.value = ''
  try {
    const res = await api.getQRCode()
    qrcodeUrl.value = res.data.url
    qrcodeKey.value = res.data.qrcodeKey
    startPolling()
  }
  catch {
    qrcodeUrl.value = ''
  }
  finally {
    qrLoading.value = false
  }
}

function startPolling() {
  clearPollTimer()
  qrPollTimer = setInterval(async () => {
    try {
      const res = await api.pollQRCode(qrcodeKey.value)
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
    catch {
      // 轮询失败不中断，继续尝试
    }
  }, 4000)
}

function refreshQrcode() {
  clearPollTimer()
  fetchQRCode()
}

function clearPollTimer() {
  if (qrPollTimer) {
    clearInterval(qrPollTimer)
    qrPollTimer = null
  }
}

async function refreshAfterLogin() {
  await fetchLoginStatus()
  if (isLoggedIn.value) {
    await fetchListenerStatus()
  }
}

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
  catch {
    // interceptor 已处理错误提示
  }
  finally {
    logoutLoading.value = false
  }
}

// ==================== 直播间相关 ====================
async function fetchListenerStatus() {
  if (!isLoggedIn.value)
    return

  listenerLoading.value = true
  try {
    const res = await api.getListenerStatus()
    const wasRunning = listenerStatus.value?.isRunning
    listenerStatus.value = res.data

    // 状态变化时管理 WebSocket
    if (res.data.isRunning && !wasRunning) {
      connectWebSocket()
    }
    else if (!res.data.isRunning && wasRunning) {
      disconnectWebSocket()
    }
  }
  catch {
    listenerStatus.value = null
    disconnectWebSocket()
  }
  finally {
    listenerLoading.value = false
  }
}

async function handleRoomUpdate() {
  const id = Number.parseInt(roomIdInput.value, 10)
  if (!id || id < 1) {
    $message.warning('请输入有效的直播间 ID')
    return
  }

  roomBindLoading.value = true
  try {
    await api.updateRoom(id)
    $message.success('直播间绑定成功')
    roomIdInput.value = ''
    showRoomEdit.value = false
    await fetchListenerStatus()
  }
  catch {
    // interceptor 已处理错误提示
  }
  finally {
    roomBindLoading.value = false
  }
}

async function handleStart() {
  startLoading.value = true
  try {
    await api.startListener()
    $message.success('监听已启动')
    await fetchListenerStatus()
  }
  catch {
    // interceptor 已处理错误提示
  }
  finally {
    startLoading.value = false
  }
}

async function handleStop() {
  stopLoading.value = true
  try {
    await api.stopListener()
    $message.success('监听已停止')
    disconnectWebSocket()
    await fetchListenerStatus()
  }
  catch {
    // interceptor 已处理错误提示
  }
  finally {
    stopLoading.value = false
  }
}

// ==================== WebSocket ====================
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

function disconnectWebSocket() {
  clearTimeout(reconnectTimer)
  if (wsClient) {
    wsClient.close()
    wsClient = null
  }
}

// ==================== 发送弹幕 ====================
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
    await api.sendDanmu(text)
    danmuText.value = ''
    $message.success('弹幕发送成功')
  }
  catch {
    // interceptor 已处理错误提示
  }
  finally {
    sendLoading.value = false
  }
}

// ==================== 辅助函数 ====================
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

// ==================== 生命周期 ====================
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
