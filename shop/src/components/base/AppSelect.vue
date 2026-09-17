<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number, Array], default: '' },
  label: { type: String, default: '' },
  placeholder: { type: String, default: '请选择' },
  options: { type: Array, default: () => [] }, // 扁平 [{label,value}] 或级联 [{label,value,children}]
  searchable: { type: Boolean, default: true }, // 仅扁平模式生效
  clearable: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'selected'])

const open = ref(false)
const query = ref('')
const crumbPath = ref([]) // 级联：导航路径，其 children 即当前展示的列表
const selectedPath = ref([]) // 级联：已提交的完整节点路径（含末级叶子），用于各级高亮
const restoreDone = ref(false)

const isCascade = computed(() => props.options.some(o => o.children?.length))

const hasValue = computed(() => {
  if (isCascade.value)
    return Array.isArray(props.modelValue) && props.modelValue.length > 0
  return props.modelValue !== '' && props.modelValue != null
})

// 触发字段展示文本
const displayValue = computed(() => {
  if (isCascade.value) {
    if (Array.isArray(props.modelValue) && props.modelValue.length) {
      const p = findPathByValue(props.options, props.modelValue)
      if (p.length === props.modelValue.length)
        return p.map(n => n.label).join(' ')
    }
    return typeof props.modelValue === 'string' && props.modelValue ? props.modelValue : ''
  }
  return props.options.find(o => o.value === props.modelValue)?.label ?? ''
})

// ---- 级联：下钻 / 面包屑 / 回填 ----
// 当前列表 = crumbPath 最末容器的 children（面包屑即导航路径本身）
const currentOptions = computed(() => crumbPath.value.length ? crumbPath.value.at(-1).children ?? [] : props.options)
const currentStepLabel = computed(() => ['选择省份', '选择城市', '选择区县'][currentOptions.value[0]?.level] ?? '选择')

// 当前展示层级上，已提交节点中同级的那个 → "点击什么就高亮什么"
function isSelected(o) {
  return selectedPath.value[crumbPath.value.length]?.value === o.value
}

function onTap(item) {
  if (item.children?.length) {
    crumbPath.value.push(item)
    return
  }
  selectedPath.value = [...crumbPath.value, item]
  const codes = selectedPath.value.map(n => n.value)
  const display = selectedPath.value.map(n => n.label).join(' ')
  emit('update:modelValue', codes)
  emit('selected', { value: codes, display })
  open.value = false
}

// 面包屑点第 i 级 → 展示该级选项（已选高亮），而非钻进其下级
function jumpTo(i) {
  crumbPath.value = crumbPath.value.slice(0, i)
}

// 按 value(codes) 逐层查，只在已匹配父节点子树内找，找不到即停
function findPathByValue(list, values) {
  const found = []
  let cur = list
  for (const v of values) {
    const n = cur.find(x => x.value === v)
    if (!n)
      break
    found.push(n)
    cur = n.children ?? []
  }
  return found
}

// 按 label 逐层查（旧数据只有展示文本时回填），限定子树避免全国重名区县
function findPathByLabel(list, labels) {
  const found = []
  let cur = list
  for (const l of labels) {
    const n = cur.find(x => x.label === l)
    if (!n)
      break
    found.push(n)
    cur = n.children ?? []
  }
  return found
}

function restoreFromValue() {
  const v = props.modelValue
  let nodes = []
  if (Array.isArray(v) && v.length)
    nodes = findPathByValue(props.options, v)
  else if (typeof v === 'string' && v.includes(' '))
    nodes = findPathByLabel(props.options, v.split(' ').map(s => s.trim()).filter(Boolean))
  selectedPath.value = nodes
  // 导航到最深一级的兄弟列表（末级已选在列表里高亮，便于改选任意一级）
  crumbPath.value = nodes.length ? nodes.slice(0, -1) : []
}

// options 可能异步到货（regions 懒加载），open + options 就绪后再回填
watch(() => [props.open, props.options], () => {
  if (props.open && isCascade.value && props.options.length && !restoreDone.value) {
    restoreFromValue()
    restoreDone.value = true
  }
  else if (!props.open) {
    restoreDone.value = false
    crumbPath.value = []
    selectedPath.value = []
    query.value = ''
  }
})

// 扁平：搜索 / 单选
const filteredOptions = computed(() => {
  const q = query.value.trim()
  return q ? props.options.filter(o => o.label.includes(q)) : props.options
})

function pickFlat(o) {
  emit('update:modelValue', o.value)
  emit('selected', { value: o.value, display: o.label })
  open.value = false
}

function clear() {
  emit('update:modelValue', isCascade.value ? [] : '')
  open.value = false
}
</script>

<template>
  <label class="block">
    <span v-if="label" class="mb-1.5 block text-sm font-medium text-fg-2">{{ label }}</span>
    <div class="relative">
      <button type="button" class="flex h-12 w-full items-center justify-between gap-2 rounded-2xl border bg-surface px-4 text-[15px] transition press" :class="[open ? 'border-primary' : 'border-line', hasValue ? 'text-fg' : 'text-fg-3']" :disabled="disabled" :aria-expanded="open" @click="open = true">
        <span class="truncate">{{ displayValue || placeholder }}</span>
        <AppIcon v-if="!(clearable && hasValue)" name="chevron-down" :size="18" class="shrink-0 text-fg-3" />
      </button>
      <button v-if="clearable && hasValue" type="button" class="absolute right-2 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-full text-fg-3 press" aria-label="清除" @click="clear">
        <AppIcon name="close" :size="14" />
      </button>
    </div>
    <AppBottomSheet v-model="open" :title="isCascade ? (label || '选择地区') : (label || '请选择')">
      <template v-if="!isCascade">
        <div v-if="searchable" class="sticky top-0 z-10 -mx-5 border-b border-line bg-surface px-5 pb-3">
          <input v-model="query" type="text" class="h-11 w-full rounded-xl border border-line bg-surface-2 px-4 text-[15px] outline-none transition placeholder:text-fg-3 focus:border-primary" placeholder="搜索选项">
        </div>
        <div class="pt-2">
          <button v-for="o in filteredOptions" :key="o.value" type="button" class="flex w-full items-center justify-between py-3 text-[15px] press" :class="o.value === modelValue ? 'font-medium text-primary' : 'text-fg'" @click="pickFlat(o)">
            <span class="truncate">{{ o.label }}</span>
            <AppIcon v-if="o.value === modelValue" name="check" :size="18" class="shrink-0" />
          </button>
          <p v-if="!filteredOptions.length" class="py-8 text-center text-sm text-fg-3">
            未找到相关选项
          </p>
        </div>
      </template>
      <template v-else>
        <div class="sticky top-0 z-10 -mx-5 border-b border-line bg-surface px-5 pb-2">
          <div class="no-scrollbar flex items-center gap-0.5 overflow-x-auto whitespace-nowrap pb-1">
            <template v-for="(n, i) in crumbPath" :key="n.value">
              <button type="button" class="shrink-0 rounded-md px-1 py-0.5 text-sm text-fg-3 press" @click="jumpTo(i)">
                {{ n.label }}
              </button>
              <AppIcon name="chevron-right" :size="14" class="shrink-0 text-fg-3" />
            </template>
            <span class="shrink-0 px-1 text-sm font-medium text-primary">{{ currentStepLabel }}</span>
          </div>
        </div>
        <div class="pt-2">
          <button v-for="o in currentOptions" :key="o.value" type="button" class="flex w-full items-center justify-between py-3 text-[15px] press" :class="isSelected(o) ? 'font-medium text-primary' : 'text-fg'" @click="onTap(o)">
            <span class="truncate">{{ o.label }}</span>
            <AppIcon v-if="o.children?.length" name="chevron-right" :size="18" class="shrink-0 text-fg-3" />
            <AppIcon v-else-if="isSelected(o)" name="check" :size="18" class="shrink-0" />
          </button>
          <p v-if="!currentOptions.length" class="py-8 text-center text-sm text-fg-3">
            暂无数据
          </p>
        </div>
      </template>
    </AppBottomSheet>
  </label>
</template>
