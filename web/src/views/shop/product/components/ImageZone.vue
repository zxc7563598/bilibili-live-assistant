<template>
  <div class="space-y-8">
    <div v-for="(image, index) in images" :key="image.uid" class="border border-gray-200 rounded-4 px-10 py-8 dark:border-gray-700" :class="image.enabled ? '' : 'opacity-60'">
      <div class="mb-6 flex items-center justify-between">
        <div class="flex items-center gap-6 text-12 text-gray-500 dark:text-gray-400">
          <span class="rounded-3 bg-gray-100 px-6 py-1 font-medium dark:bg-gray-800">{{ index + 1 }}</span>
          <span v-if="!image.enabled" class="text-warning">已停用，商城端不展示</span>
        </div>
        <div class="flex items-center gap-4">
          <n-tooltip>
            <template #trigger>
              <n-button quaternary size="tiny" :disabled="index === 0" @click="move(index, -1)">
                <template #icon>
                  <i class="i-fe:arrow-up" />
                </template>
              </n-button>
            </template>
            上移
          </n-tooltip>
          <n-tooltip>
            <template #trigger>
              <n-button quaternary size="tiny" :disabled="index === images.length - 1" @click="move(index, 1)">
                <template #icon>
                  <i class="i-fe:arrow-down" />
                </template>
              </n-button>
            </template>
            下移
          </n-tooltip>
          <n-tooltip>
            <template #trigger>
              <n-switch v-model:value="image.enabled" size="small" />
            </template>
            是否启用
          </n-tooltip>
          <n-button quaternary size="tiny" type="error" @click="remove(index)">
            <template #icon>
              <i class="i-fe:trash-2" />
            </template>
          </n-button>
        </div>
      </div>
      <UploadInput v-model:value="image.path" :upload="upload" :sync-oss="syncOss" :placeholder="placeholder" />
    </div>
    <div v-if="!images.length" class="border border-gray-200 rounded-4 border-dashed px-10 py-16 text-center text-13 text-gray-400 dark:border-gray-700">
      {{ emptyTip }}
    </div>
    <n-button secondary block type="primary" @click="add">
      <template #icon>
        <i class="i-fe:plus" />
      </template>
      添加图片
    </n-button>
  </div>
</template>

<script setup>
import { nextUid } from '../spec-utils'

defineOptions({ name: 'ImageZone' })

defineProps({
  /** 上传方法，透传给 UploadInput */
  upload: { type: Function, default: null },
  /** 同步 OSS 方法，透传给 UploadInput */
  syncOss: { type: Function, default: null },
  /** 路径输入框占位提示 */
  placeholder: { type: String, default: '请选择或输入图片路径' },
  /** 空列表时的提示文案 */
  emptyTip: { type: String, default: '还没有图片' },
})

const images = defineModel('images', { type: Array, default: () => [] })

function add() {
  images.value.push({ uid: nextUid('img'), path: '', enabled: true })
}

function remove(index) {
  images.value.splice(index, 1)
}

/** 上移 / 下移：sort_order 在提交时按数组下标派生，这里直接换位置即可 */
function move(index, offset) {
  const target = index + offset
  if (target < 0 || target >= images.value.length)
    return
  const [item] = images.value.splice(index, 1)
  images.value.splice(target, 0, item)
}
</script>
