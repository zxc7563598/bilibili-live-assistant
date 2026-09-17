<!--
  省 / 市 / 区三级联动。v-model 是行政区划 code 数组，如 ['370000','370100','370116']，
  与 live_user_orders.receiver_region_code / live_user_addresses.region_code 的存储形式一一对应。

  地区数据来源：src/data/regions.json，是 internal/region/regions.json 的副本
  （该文件本身又是 shop/src/data/regions.json 的副本，见 internal/region/embed.go 的注释）。
  三份内容必须保持一致，更新行政区划时一起替换。

  之所以要在前端再放一份：internal/region 只导出了 Resolve(codes) (string, bool) 用于服务端校验，
  没有任何 HTTP 接口能把整棵树发给前端。

  树的字段名 { value, label, children } 与 NCascader 的默认取值字段完全一致，所以零映射直接喂。

  只能选到叶子节点：后端 internal/region.Resolve 要求 code 链至少两段且逐级是直接子节点，
  只选到省（甚至只有省）会被打回 11302「地区信息不正确」。
  列里点选时 NCascader 自己就挡住了非叶子（handleClick 里 if (isLeaf) 才 toggleCheckbox），
  但搜索框与键盘 Enter 走的是 CascaderSelectMenu.doCheck，对任意节点都会调 doCheck，
  所以这里显式指定 check-strategy="child"，并在 handleUpdate 里再兜一层只认叶子。
-->
<template>
  <n-cascader :value="leafValue" :options="options" :disabled="!ready" :placeholder="placeholder" check-strategy="child" filterable @update:value="handleUpdate" />
</template>

<script setup>
const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
})
const emit = defineEmits(['update:modelValue'])

const options = ref([])
const ready = ref(false)
const loadFailed = ref(false)

const placeholder = computed(() => {
  if (loadFailed.value)
    return '地区数据加载失败，请刷新页面重试'
  if (!ready.value)
    return '地区数据加载中…'
  return '请选择省 / 市 / 区'
})

// NCascader 单选的 value 就是叶子节点的 value，所以这里取 code 链的最后一段即可。
const leafValue = computed(() => {
  const codes = Array.isArray(props.modelValue) ? props.modelValue : []
  return codes.length ? codes.at(-1) : null
})

/**
 * 在树里自顶向下找 value 所在的完整 code 链，且只有叶子节点才算命中；找不到返回 null。
 *
 * 这里刻意不使用 NCascader on-update:value 的第 3 个 path 参数：自己 DFS 只有一条代码路径，
 * 不受组件 emit 签名在各版本间的差异影响，filterable 与逐个下钻两条交互也走同一处逻辑。
 */
function findPath(nodes, value) {
  for (const node of nodes) {
    if (node.value === value)
      return node.children?.length ? null : [node.value]
    if (node.children?.length) {
      const sub = findPath(node.children, value)
      if (sub)
        return [node.value, ...sub]
    }
  }
  return null
}

function handleUpdate(value) {
  const path = findPath(options.value, value)
  // 解不出路径时不 emit：绝不能把 code 链清成 []，否则提交时地区为空会被后端打回 11306；
  // 同理，命中的不是叶子（只选到省 / 市）时也不 emit，链停在中间会被后端打回 11302
  if (path)
    emit('update:modelValue', path)
}

onMounted(async () => {
  try {
    options.value = (await import('@/data/regions.json')).default
    ready.value = true
  }
  catch {
    // 地区数据是本地打包资源，加载失败只可能是构建/部署问题，不需要 toast，
    // 把状态写进 placeholder 让管理员看得见即可
    loadFailed.value = true
  }
})
</script>
