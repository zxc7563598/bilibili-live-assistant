<template>
  <div class="min-w-0 w-full flex items-center">
    <n-input class="min-w-0 flex-1" :value="value" clearable :placeholder="placeholder" :disabled="disabled || busy" @update:value="v => emit('update:value', v)" />
    <n-button v-if="upload" size="medium" class="rounded-l-none px-10 -ml-px" title="上传图片" :loading="uploading" :disabled="disabled || syncing" @click="openPicker">
      <template #icon>
        <i class="i-mdi:image-plus text-16" />
      </template>
      上传
    </n-button>
    <div class="h-34 w-34 f-c-c shrink-0 overflow-hidden border border-gray-200 bg-white transition -ml-px dark:border-gray-700 dark:bg-gray-800/60" :class="{ 'pointer-events-none opacity-60': disabled || busy }">
      <n-image v-if="value" width="32" height="32" object-fit="cover" :src="value" alt="图片预览" />
      <i v-else class="i-mdi:image text-15 text-gray-300 dark:text-gray-500" />
    </div>
    <n-button v-if="syncOss" size="medium" class="rounded-l-none px-10 -ml-px" :loading="syncing" :disabled="disabled || busy || !canSync" @click="handleSync">
      <template #icon>
        <i class="i-mdi:cloud-upload-outline text-16" />
      </template>
      同步OSS
    </n-button>
    <input ref="fileRef" type="file" class="hidden" tabindex="-1" aria-hidden="true" :accept="accept" @change="onFileChange">
  </div>
</template>

<script setup>
defineOptions({ name: 'UploadInput' })

const props = defineProps({
  /** 绑定的图片路径（本地路径 / blob / http(s) 地址） */
  value: { type: String, default: '' },
  /** 上传方法：upload(file: File) → Promise<{ data: { path } }>；不传则隐藏上传按钮 */
  upload: { type: Function, default: null },
  /** 同步到 OSS：syncOss({ path }) → Promise<{ data: { path: http(s)地址 } }>；不传则隐藏该按钮 */
  syncOss: { type: Function, default: null },
  /** 允许选择的文件类型，透传给隐藏的 file input */
  accept: { type: String, default: 'image/*' },
  /** 上传大小上限（MB），0 表示不限制 */
  maxSize: { type: Number, default: 5 },
  /** 路径为空时的占位提示 */
  placeholder: { type: String, default: '请选择或输入图片路径' },
  /** 是否禁用整个组件 */
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:value'])

const HTTP_URL_RE = /^https?:\/\//i

const fileRef = ref(null)
const uploading = ref(false)
const syncing = ref(false)

const busy = computed(() => uploading.value || syncing.value)
// 仅 http(s) 视为远端地址；其余（相对路径/data:/blob:）视为本地、可同步
const isRemote = computed(() => HTTP_URL_RE.test(props.value))
const canSync = computed(() => !!props.syncOss && !!props.value && !isRemote.value)

// 输入框本身只负责编辑/手填路径，上传仅通过「上传」按钮触发
function openPicker() {
  if (props.disabled || busy.value || !props.upload)
    return
  fileRef.value?.click()
}

async function onFileChange(e) {
  const file = e.target.files?.[0]
  // 及时重置，保证再次选择同一文件也能触发 change
  e.target.value = ''
  if (!file || props.disabled || busy.value || !props.upload)
    return

  // 文件类型校验
  const isImageAccept = props.accept.startsWith('image/')
  if (isImageAccept && file.type && !file.type.startsWith('image/')) {
    $message.warning('请选择图片文件')
    return
  }
  // 大小校验
  if (props.maxSize > 0 && file.size > props.maxSize * 1024 * 1024) {
    $message.warning(`图片不能超过 ${props.maxSize}MB`)
    return
  }

  uploading.value = true
  $loadingBar.start()
  try {
    const res = await props.upload(file)
    const path = res?.data?.path
    if (typeof path === 'string' && path)
      emit('update:value', path)
    $loadingBar.finish()
  }
  catch {
    // 失败提示由 http 拦截器统一 toast，这里只停掉 loading、保留旧值
    $loadingBar.error()
  }
  finally {
    uploading.value = false
  }
}

async function handleSync() {
  if (props.disabled || busy.value || !canSync.value || !props.syncOss)
    return
  syncing.value = true
  $loadingBar.start()
  try {
    const res = await props.syncOss({ path: props.value })
    const url = res?.data?.path
    if (typeof url === 'string' && url) {
      emit('update:value', url)
      $message.success('已同步到 OSS')
    }
    $loadingBar.finish()
  }
  catch {
    $loadingBar.error()
  }
  finally {
    syncing.value = false
  }
}
</script>
