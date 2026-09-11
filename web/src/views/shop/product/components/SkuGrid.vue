<template>
  <div class="space-y-10">
    <div class="flex flex-wrap items-center justify-between gap-8">
      <div class="text-13 text-gray-500 dark:text-gray-400">
        共 <span class="text-primary font-medium">{{ rows.length }}</span> 个组合，已上架 <span class="text-primary font-medium">{{ enabledCount }}</span> 个
      </div>
      <div class="flex items-center gap-6">
        <n-tooltip>
          <template #trigger>
            <n-button size="small" secondary @click="fillPrice">
              同步展示价
            </n-button>
          </template>
          把所有已上架组合的价格设为商品展示价
        </n-tooltip>
        <n-button size="small" secondary @click="setAll(true)">
          全部上架
        </n-button>
        <n-button size="small" secondary @click="setAll(false)">
          全部下架
        </n-button>
      </div>
    </div>
    <div v-if="!rows.length" class="border border-gray-200 rounded-4 border-dashed px-10 py-16 text-center text-13 text-gray-400 dark:border-gray-700">
      请先在上方添加规格并填写规格值，系统会自动生成规格组合。
    </div>
    <div v-else class="grid grid-cols-1 gap-8 lg:grid-cols-2 xl:grid-cols-3">
      <div v-for="row in rows" :key="row.key" class="border rounded-4 px-10 py-8 transition" :class="row.enabled ? 'border-primary/40 bg-primary/4 dark:bg-primary/8' : 'border-gray-200 bg-gray-50 opacity-70 dark:border-gray-700 dark:bg-gray-800/40'">
        <div class="flex items-center justify-between gap-8">
          <div class="min-w-0 flex-1 truncate text-13 font-medium" :title="propsLabel(row.specProperties)">
            {{ propsLabel(row.specProperties) }}
          </div>
          <n-switch v-model:value="row.enabled" size="small" />
        </div>
        <div v-if="row.enabled" class="mt-8 space-y-6">
          <div class="flex items-center gap-6">
            <div class="w-60 shrink-0 text-12 text-gray-500 dark:text-gray-400">
              价格
            </div>
            <n-input-number v-model:value="row.price" class="min-w-0 flex-1" size="small" :min="0" :precision="0" placeholder="不小于 0 的整数">
              <template #suffix>
                {{ creditType === 0 ? '星光' : '积分' }}
              </template>
            </n-input-number>
          </div>
          <div class="flex items-center gap-6">
            <div class="w-60 shrink-0 text-12 text-gray-500 dark:text-gray-400">
              成本价
            </div>
            <n-input-number v-model:value="row.costPrice" class="min-w-0 flex-1" size="small" :min="0" :precision="0" placeholder="不小于 0 的整数" />
          </div>
          <div class="flex items-center gap-6">
            <div class="w-60 shrink-0 text-12 text-gray-500 dark:text-gray-400">
              库存
            </div>
            <n-input-number v-model:value="row.stock" class="min-w-0 flex-1" size="small" :min="0" :precision="0" placeholder="不小于 0 的整数">
              <template #suffix>
                件
              </template>
            </n-input-number>
          </div>
        </div>
        <div v-else class="mt-8 text-12 text-gray-400 dark:text-gray-500">
          未上架，用户无法选择该组合
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { propsLabel } from '../spec-utils'

defineOptions({ name: 'SkuGrid' })

const props = defineProps({
  defaultPrice: { type: Number, default: 0 },
  creditType: { type: Number, default: 0 },
})

const rows = defineModel('rows', { type: Array, default: () => [] })

const enabledCount = computed(() => rows.value.filter(r => r.enabled).length)

function setAll(enabled) {
  for (const row of rows.value)
    row.enabled = enabled
}

function fillPrice() {
  const price = Number(props.defaultPrice) || 0
  for (const row of rows.value) {
    if (row.enabled)
      row.price = price
  }
  $message.success('已按商品展示价填充')
}
</script>
