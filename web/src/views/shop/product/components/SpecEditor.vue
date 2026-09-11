<template>
  <div class="space-y-10">
    <div v-for="(spec, specIndex) in specs" :key="spec.uid" class="border border-gray-200 rounded-4 px-10 py-8 dark:border-gray-700">
      <div class="flex items-center gap-8">
        <span class="shrink-0 text-12 text-gray-400 dark:text-gray-500">规格 {{ specIndex + 1 }}</span>
        <n-input v-model:value="spec.key_name" class="min-w-0 flex-1" :maxlength="100" placeholder="规格名称，如「颜色」「尺寸」" />
        <n-button quaternary type="error" @click="removeSpec(specIndex)">
          <template #icon>
            <i class="i-fe:trash-2" />
          </template>
          删除规格
        </n-button>
      </div>
      <div class="mt-8 pl-4">
        <div class="mb-5 text-12 text-gray-400 dark:text-gray-500">
          规格值
        </div>
        <div class="flex flex-wrap gap-6">
          <div v-for="(value, valueIndex) in spec.values" :key="value.uid" class="w-200 flex items-center">
            <n-input v-model:value="value.value_name" :maxlength="100" size="small" placeholder="规格值" />
            <n-button quaternary size="small" type="error" class="ml-2 shrink-0" :disabled="spec.values.length <= 1" @click="removeValue(spec, valueIndex)">
              <template #icon>
                <i class="i-fe:x" />
              </template>
            </n-button>
          </div>
          <n-button size="small" secondary type="primary" @click="addValue(spec)">
            <template #icon>
              <i class="i-fe:plus" />
            </template>
            添加规格值
          </n-button>
        </div>
      </div>
    </div>
    <div v-if="!specs.length" class="border border-gray-200 rounded-4 border-dashed px-10 py-16 text-center text-13 text-gray-400 dark:border-gray-700">
      暂无规格。不添加规格时，商品会按「默认规格」单个 SKU 上架。
    </div>
    <n-button secondary block type="primary" @click="addSpec">
      <template #icon>
        <i class="i-fe:plus" />
      </template>
      添加规格类型
    </n-button>
  </div>
</template>

<script setup>
import { nextUid } from '../spec-utils'

defineOptions({ name: 'SpecEditor' })

const specs = defineModel('specs', { type: Array, default: () => [] })

function addSpec() {
  specs.value.push({
    uid: nextUid('s'),
    key_name: '',
    // 新建规格默认带一个空值，避免出现「有规格类型却没有值」的中间态
    values: [{ uid: nextUid('v'), value_name: '' }],
  })
}

function removeSpec(index) {
  specs.value.splice(index, 1)
}

function addValue(spec) {
  spec.values.push({ uid: nextUid('v'), value_name: '' })
}

function removeValue(spec, index) {
  spec.values.splice(index, 1)
}
</script>
