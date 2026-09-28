<template>
  <div class="card-border rounded-12 auto-bg p-16">
    <div class="text-13 text-gray-500 leading-relaxed dark:text-gray-400">
      之前用的是旧版系统的话，可以把那边导出的 <span class="text-gray-700 font-medium dark:text-gray-200">.gz</span> 数据文件在这里导入，
      弹幕、礼物、用户以及对应的积分/星光流水都会一并还原。只允许在弹幕、礼物、用户三张表均为空时导入。
    </div>

    <div class="mt-14 flex flex-wrap items-center gap-10">
      <n-button size="small" :disabled="busy" @click="openPicker">
        <template #icon>
          <i class="i-fe:file-plus" />
        </template>
        选择数据文件
      </n-button>
      <div class="min-w-0 flex-1 text-12 text-gray-400">
        <span v-if="file" class="flex items-center gap-5">
          <i class="i-fe:file shrink-0 text-13" />
          <span class="truncate text-gray-600 dark:text-gray-300">{{ file.name }}</span>
          <span class="shrink-0">{{ formatSize(file.size) }}</span>
        </span>
        <span v-else>还没有选择文件</span>
      </div>
      <n-button size="small" type="primary" :loading="uploading" :disabled="!file || busy" @click="handleImport">
        开始导入
      </n-button>
    </div>

    <input ref="fileRef" type="file" accept=".gz,application/gzip" class="hidden" tabindex="-1" aria-hidden="true" @change="onFileChange">

    <template v-if="started">
      <n-progress
        class="mt-16"
        type="line"
        :percentage="percent"
        :status="progressStatus"
        :processing="running"
        :height="8"
        :border-radius="4"
      />
      <div class="mt-6 flex items-center gap-5 text-12 text-gray-400">
        <i v-if="finished" class="i-fe:check-circle text-14 text-emerald-500" />
        <span>{{ progressText }}</span>
      </div>

      <StatisticsCard :stats="countStats" />

      <div v-if="finished && skippedText" class="text-12 text-gray-400 leading-relaxed">
        {{ skippedText }}
      </div>

      <div v-if="failed" class="mt-10 flex gap-8 border border-amber-200 rounded-8 bg-amber-50 px-12 py-10 dark:border-amber-500/25 dark:bg-amber-500/10">
        <i class="i-fe:alert-triangle mt-1 shrink-0 text-14 text-amber-500" />
        <div class="text-12 text-amber-700 leading-relaxed dark:text-amber-300">
          {{ errorText }}
        </div>
      </div>

      <div v-if="!running" class="mt-4 flex justify-end">
        <n-button size="tiny" @click="handleDismiss">
          知道了
        </n-button>
      </div>
    </template>
  </div>
</template>

<script setup>
import { useStorage } from '@vueuse/core'
import { StatisticsCard } from '@/components'
import api from '../api'

defineOptions({ name: 'HomeImportCard' })

// 与后端 upload.importFileLimit 保持一致
const IMPORT_FILE_MAX = 100 * 1024 * 1024
// 后台每导入一批会刷新一次进度，2 秒足够跟上
const POLL_INTERVAL = 2000

// 进度接口只返回状态与错误码，文案在前端映射，未命中的取值给兜底文案
const TABLE_LABEL = { bl_danmu_logs: '弹幕', bl_gift_records: '礼物', bl_user_vips: '用户' }
const ERROR_LABEL = {
  11408: '这个文件不是旧版导出的数据文件，请确认后再试。',
  11411: '文件来源不匹配，这个文件不是旧版导出的数据文件。',
  41403: '当前系统已经存在数据，无法导入。',
  51402: '导入任务已失效，请重新发起导入。',
}

const fileRef = ref(null)
const file = ref(null)
const uploading = ref(false)
const progress = ref(null)
// 任务标识持久化：刷新或切走再回来时，还能接上后台正在跑的任务
const taskId = useStorage('home.importTaskId', '')

let pollTimer = null

const started = computed(() => !!progress.value)
const running = computed(() => progress.value?.status === 0 || progress.value?.status === 1)
const finished = computed(() => progress.value?.status === 2)
const failed = computed(() => progress.value?.status === 3)
const busy = computed(() => uploading.value || running.value)
const percent = computed(() => Math.min(100, Number(progress.value?.percent ?? 0)))
const progressStatus = computed(() => (finished.value ? 'success' : failed.value ? 'error' : 'default'))

const progressText = computed(() => {
  if (finished.value)
    return '导入完成'
  if (running.value)
    return `导入中，当前：${TABLE_LABEL[progress.value?.current_table] || '准备中'}`
  return '导入失败'
})

const errorText = computed(() => ERROR_LABEL[progress.value?.error_code] || '导入失败，请稍后重试。')

const countStats = computed(() => [
  { label: '弹幕', value: progress.value?.danmu_count ?? 0, tooltip: '已导入的弹幕条数' },
  { label: '礼物', value: progress.value?.gift_count ?? 0, tooltip: '已导入的礼物条数' },
  { label: '用户', value: progress.value?.user_count ?? 0, tooltip: '已导入的用户条数' },
  { label: '资产流水', value: progress.value?.credit_log_count ?? 0, tooltip: '按导入值补录的积分/星光流水条数' },
])

const skippedText = computed(() => {
  const skipped = progress.value?.skipped_user_count ?? 0
  if (!skipped)
    return ''
  return `有 ${skipped} 条 uid 非法或重复的用户记录被跳过。`
})

function formatSize(size) {
  if (size >= 1024 * 1024)
    return `${(size / 1024 / 1024).toFixed(1)}MB`
  return `${Math.max(1, Math.round(size / 1024))}KB`
}

function openPicker() {
  if (busy.value)
    return
  fileRef.value?.click()
}

function onFileChange(e) {
  const picked = e.target.files?.[0]
  // 及时重置，保证再次选择同一文件也能触发 change
  e.target.value = ''
  if (!picked)
    return
  if (!picked.name.toLowerCase().endsWith('.gz')) {
    $message.warning('请选择旧版导出的 .gz 数据文件')
    return
  }
  if (picked.size > IMPORT_FILE_MAX) {
    $message.warning(`文件不能超过 ${IMPORT_FILE_MAX / 1024 / 1024}MB`)
    return
  }
  file.value = picked
}

function handleImport() {
  const picked = file.value
  if (!picked || busy.value)
    return
  const d = $dialog.warning({
    title: '确认导入数据',
    content: `导入会把「${picked.name}」里的历史数据写入弹幕、礼物、用户三张表，且要求这三张表当前为空。开始后无法中途取消，确认导入吗？`,
    positiveText: '开始导入',
    negativeText: '取消',
    async onPositiveClick() {
      try {
        d.loading = true
        uploading.value = true
        const { data } = await api.importData(picked)
        file.value = null
        taskId.value = data.task_id
        progress.value = null
        await pollProgress()
        startPolling()
        $message.success('已开始导入，进度会在这里实时更新')
        d.loading = false
      }
      catch (error) {
        // 失败提示由响应拦截器统一弹出（已有任务、已有数据、文件超限等）
        console.error('发起导入失败', error)
        d.loading = false
      }
      finally {
        uploading.value = false
      }
    },
  })
}

function handleDismiss() {
  clearPollTimer()
  progress.value = null
  taskId.value = ''
}

function startPolling() {
  clearPollTimer()
  pollTimer = setInterval(pollProgress, POLL_INTERVAL)
}

function clearPollTimer() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function pollProgress() {
  if (!taskId.value) {
    clearPollTimer()
    return
  }
  try {
    const { data } = await api.getImportProgress(taskId.value)
    progress.value = data
    // 已完成或已失败，停表保留结果
    if (data.status === 2 || data.status === 3)
      clearPollTimer()
  }
  catch (error) {
    // 任务已被淘汰：清掉本地记录回到初始态；其它错误（网络抖动）继续轮询
    if (error?.code === 51402) {
      clearPollTimer()
      handleDismiss()
    }
  }
}

onMounted(async () => {
  if (!taskId.value)
    return
  await pollProgress()
  if (running.value)
    startPolling()
})

onUnmounted(clearPollTimer)
</script>
