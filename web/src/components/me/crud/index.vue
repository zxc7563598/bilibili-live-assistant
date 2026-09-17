<!-- Copyright © 2023 Ronnie Zhang (大脸怪). MIT License. -->

<template>
  <div class="h-full flex flex-col overflow-hidden">
    <AppCard v-if="$slots.default" bordered bg="#fafafc dark:black" class="mb-30 min-h-60 rounded-4">
      <form class="flex justify-between p-16" @submit.prevent="handleSearch()">
        <n-scrollbar x-scrollable>
          <n-space :wrap="!expand || isExpanded" :size="[32, 16]" class="p-10">
            <slot />
          </n-space>
        </n-scrollbar>
        <div class="shrink-0 p-10">
          <n-button ghost type="primary" @click="handleReset">
            <i class="i-fe:rotate-ccw mr-4" />
            重置
          </n-button>
          <n-button attr-type="submit" class="ml-20" type="primary">
            <i class="i-fe:search mr-4" />
            搜索
          </n-button>
          <template v-if="expand">
            <n-button v-if="!isExpanded" type="primary" text @click="toggleExpand">
              <i class="i-fe:chevrons-down ml-4" />
              展开
            </n-button>
            <n-button v-else text type="primary" @click="toggleExpand">
              <i class="i-fe:chevrons-up ml-4" />
              收起
            </n-button>
          </template>
        </div>
      </form>
    </AppCard>

    <AppCard v-if="$slots.statistic" bordered bg="#fafafc dark:black" class="mb-30 min-h-60 rounded-4">
      <slot name="statistic" />
    </AppCard>

    <NDataTable
      ref="nTableRef"
      :remote="remote"
      :loading="loading"
      :scroll-x="scrollX"
      :columns="columns"
      :data="tableData"
      :row-key="(row) => row[rowKey]"
      :pagination="isPagination ? pagination : false"
      flex-height
      class="flex-1"
      @update:checked-row-keys="onChecked"
      @update:page="onPageChange"
      @update:sorter="onSorterChange"
    />
  </div>
</template>

<script setup>
import { NDataTable } from 'naive-ui'
import { utils, writeFile } from 'xlsx'

const props = defineProps({
  /**
   * @remote true: 后端分页  false： 前端分页
   */
  remote: {
    type: Boolean,
    default: true,
  },
  /**
   * @isPagination 是否分页
   */
  isPagination: {
    type: Boolean,
    default: true,
  },
  scrollX: {
    type: Number,
    default: 1200,
  },
  rowKey: {
    type: String,
    default: 'id',
  },
  columns: {
    type: Array,
    required: true,
  },
  /** queryBar中的参数 */
  queryItems: {
    type: Object,
    default() {
      return {}
    },
  },
  /**
   * ! 约定接口入参出参
   * 分页模式需约定分页接口入参
   *    @pageSize 分页参数：一页展示多少条，默认10
   *    @pageNo   分页参数：页码，默认1
   * 远程排序，列配置中给数据列加 sorter:true 即启用：
   *    @sortField DB 排序列名（默认取列 key，可用列 sortField 覆盖）
   *    @sortOrder ascend|descend（Naive UI 表头原值，直接透传），无排序时不携带这两项；后端需按白名单校验并映射 SQL 方向
   * 需约定接口出参
   *    @pageData 分页模式必须,非分页模式如果没有pageData则取上一层data
   *    @total    分页模式必须，非分页模式如果没有total则取上一层data.length
   */
  getData: {
    type: Function,
    required: true,
  },
  /** 是否支持展开 */
  expand: Boolean,
})

const emit = defineEmits(['update:queryItems', 'onChecked', 'onDataChange'])
const loading = ref(false)
const initQuery = { ...props.queryItems }
const tableData = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 20,
  prefix({ itemCount }) {
    return `共 ${itemCount} 条数据`
  },
})

const nTableRef = ref(null) // NDataTable 实例，重置时清除排序箭头
const sort = reactive({ field: null, order: null })

// 是否展开
const isExpanded = ref(false)

function toggleExpand() {
  isExpanded.value = !isExpanded.value
}

async function handleQuery() {
  try {
    loading.value = true
    let paginationParams = {}
    // 如果非分页模式或者使用前端分页,则无需传分页参数
    if (props.isPagination && props.remote) {
      paginationParams = { pageNo: pagination.page, pageSize: pagination.pageSize }
    }
    // 无活动排序时不带排序参数，后端回退默认排序
    const sortParams = props.remote && sort.order
      ? { sortField: sort.field, sortOrder: sort.order }
      : {}
    const { data } = await props.getData({
      ...props.queryItems,
      ...paginationParams,
      ...sortParams,
    })
    tableData.value = data?.pageData || data
    pagination.itemCount = data.total ?? data.length
    if (pagination.itemCount && !tableData.value.length && pagination.page > 1) {
      // 如果当前页数据为空，且总条数不为0，则返回上一页数据
      onPageChange(pagination.page - 1)
    }
  }
  catch (error) {
    console.error(error)
    tableData.value = []
    pagination.itemCount = 0
  }
  finally {
    emit('onDataChange', tableData.value)
    loading.value = false
  }
}

function handleSearch(keepCurrentPage = false) {
  if (keepCurrentPage || !props.remote) {
    handleQuery()
  }
  else {
    onPageChange(1)
  }
}
async function handleReset() {
  sort.field = null
  sort.order = null
  const queryItems = { ...props.queryItems }
  for (const key in queryItems) {
    queryItems[key] = null
  }
  emit('update:queryItems', { ...queryItems, ...initQuery })
  await nextTick()
  // 清除表头排序箭头；其发出的 update:sorter(null) 在回调里被忽略，避免二次请求
  nTableRef.value?.clearSorter()
  pagination.page = 1
  handleQuery()
}
function onPageChange(currentPage) {
  pagination.page = currentPage
  if (props.remote) {
    handleQuery()
  }
}
function onChecked(rowKeys) {
  if (props.columns.some(item => item.type === 'selection')) {
    emit('onChecked', rowKeys)
  }
}
// 递归查找列配置
function findColumn(cols, key) {
  for (const col of cols) {
    if (col.children) {
      const hit = findColumn(col.children, key)
      if (hit)
        return hit
    }
    if (col.key === key)
      return col
  }
  return null
}

function onSorterChange(sorterState) {
  // sorterState: SortState | SortState[] | null（本项目为单列排序，未启用 multiple）
  if (Array.isArray(sorterState))
    sorterState = sorterState[0]
  // 程序化 clearSorter() 会发 null，由 handleReset 统一处理，避免重复请求
  if (!sorterState)
    return
  const col = findColumn(props.columns, sorterState.columnKey)
  if (!col)
    return
  // 排序原值（ascend/descend/false）直接透传给后端校验映射，前端不做翻译
  const order = sorterState.order || null
  // 默认 DB 排序列名取列 key；key 与 DB 列不一致时可用列上 sortField 覆盖
  const field = order ? (col.sortField ?? sorterState.columnKey) : null
  if (sort.field === field && sort.order === order)
    return // 状态未变化则忽略
  sort.field = field
  sort.order = order
  if (props.remote) {
    pagination.page = 1
    handleQuery()
  }
}
function handleExport(columns = props.columns, data = tableData.value) {
  if (!data?.length)
    return $message.warning('没有数据')
  const columnsData = columns.filter(item => !!item.title && !item.hideInExcel)
  const thKeys = columnsData.map(item => item.key)
  const thData = columnsData.map(item => typeof item.title === 'function' ? item.key : item.title)
  const trData = data.map(item => thKeys.map(key => item[key]))
  const sheet = utils.aoa_to_sheet([thData, ...trData])
  const workBook = utils.book_new()
  utils.book_append_sheet(workBook, sheet, '数据报表')
  writeFile(workBook, '数据报表.xlsx')
}

defineExpose({
  handleSearch,
  handleReset,
  handleExport,
})
</script>
