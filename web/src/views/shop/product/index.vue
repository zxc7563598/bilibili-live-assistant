<template>
  <CommonPage>
    <template #action>
      <NButton v-permission="'AddUser'" type="primary" @click="handleAdd()">
        <i class="i-material-symbols:add mr-4 text-18" />
        添加商品
      </NButton>
    </template>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="800">
      <MeQueryItem label="商品名称">
        <n-input v-model:value="query.name" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="积分类型">
        <n-select v-model:value="query.credit_type" :options="creditTypeOptions" placeholder="请指定积分类型" />
      </MeQueryItem>
    </MeCrud>
  </CommonPage>
</template>

<script setup>
import { NButton, NSwitch } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import api from './api'

const loadingMap = reactive({})
const router = useRouter()

// 列表信息
const $table = ref(null)
const columns = [
  { title: '商品名称', key: 'name', minWidth: 180, sorter: true },
  { title: '商品标价', key: 'price', width: 160, sorter: true, render(row) {
    return row.price + getCreditTypeLabel(row.credit_type)
  } },
  { title: '已售', key: 'sold', width: 120, sorter: true },
  { title: '库存', key: 'stock', width: 120, sorter: true },
  { title: '上架', key: 'enable', width: 120, sorter: true, render(row) {
    return [
      h(
        NSwitch,
        {
          'size': 'small',
          'round': false,
          'value': row.enable,
          'loading': !!loadingMap[row.id],
          'onUpdate:value': (val) => {
            handleEnable(row.id, val)
          },
        },
      ),
    ]
  } },
  { title: '排序', key: 'sort_order', width: 120, sorter: true },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    align: 'right',
    fixed: 'right',
    render(row) {
      return [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            secondary: true,
            onClick: () => router.push({ path: '/shop/product/details', query: { id: row.id } }),
          },
          {
            default: () => '商品管理',
            icon: () => h('i', { class: 'i-fe:calendar text-14' }),
          },
        ),
      ]
    },
  },
]

const query = ref({
  name: null,
  credit_type: null,
})
const creditTypeOptions = ref([
  {
    label: '星光',
    value: 0,
  },
  {
    label: '积分',
    value: 1,
  },
])

function getCreditTypeLabel(value) {
  const item = creditTypeOptions.value.find(item => item.value === value)
  return item ? item.label : ''
}

function handleAdd() {
  router.push({ path: '/shop/product/details' })
}

async function handleEnable(id, enable) {
  loadingMap[id] = true
  try {
    await api.productEnable(id, enable)
  }
  catch {
    // 失败提示由 http 拦截器统一处理，这里吞掉 rejection 避免冒到控制台
  }
  finally {
    // 成功与否都刷新列表，让开关显示回到后端的真实状态
    delete loadingMap[id]
    $table.value?.handleSearch(true)
  }
}

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
