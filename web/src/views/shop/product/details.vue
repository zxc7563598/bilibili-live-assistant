<template>
  <CommonPage back :title="pageTitle">
    <template #action>
      <div class="flex items-center gap-8">
        <n-button secondary @click="handleCancel">
          取消
        </n-button>
        <n-button v-if="!loadFailed" type="primary" :loading="saving" @click="handleSubmit">
          保存商品
        </n-button>
      </div>
    </template>
    <div v-show="loaded" class="flex flex-col gap-12 lg:flex-row lg:items-start">
      <div class="min-w-0 flex-1 space-y-12">
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                基本信息
              </div>
              <div class="mt-2 text-13 text-gray-400">
                商品在商城列表和详情页顶部展示的内容
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「商品封面」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  商品封面是用户在商城列表里看到的第一眼内容，建议使用 1:1 正方形图片，主体居中、背景干净，避免出现文字或水印。
                </div>
                <div>
                  除了这里的封面，下方还有「轮播图」和「详情图」两组图片：轮播图会在商品详情页顶部横向滑动展示，详情图则按顺序纵向铺在商品说明下方。
                </div>
              </div>
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                商品名称
              </div>
              <n-input v-model:value="form.name" :maxlength="255" placeholder="展示在商城里的商品名字" />
            </div>
            <div class="flex items-center gap-5">
              <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
                商品封面
              </div>
              <UploadInput v-model:value="form.cover" :upload="file => api.uploadImage(file, 'cover')" :sync-oss="api.syncOss" :max-size="5" placeholder="请上传或填写封面图片地址" />
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 pt-6 text-right text-13 text-gray-500 dark:text-gray-400">
                商品说明
              </div>
              <n-input v-model:value="form.describe" type="textarea" :rows="3" :maxlength="1000" show-count placeholder="一句话介绍这个商品，例如：下单后 1-3 天发出哦~" />
            </div>
            <div class="flex items-start gap-5">
              <div class="w-100 shrink-0 pt-6 text-right text-13 text-gray-500 dark:text-gray-400">
                商品标签
              </div>
              <div class="min-w-0 flex-1 space-y-5">
                <TagEditor v-model:tags="tagList" />
                <div class="text-12 text-gray-400 dark:text-gray-500">
                  输入后回车即可添加，点标签上的 × 删除；用于商城里的筛选与展示，可以留空。
                </div>
              </div>
            </div>
          </div>
        </n-card>
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                规格设置
              </div>
              <div class="mt-2 text-13 text-gray-400">
                先在这里定义商品的规格类型与规格值，下方的 SKU 组合会根据它们自动生成
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「规格」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  规格是这个商品可以选择的维度，例如「类型」下有「镭射款 / 非镭射款」，「形象」下有「wink小蓝 / 傲娇小团子」。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">规格值</span>：每个规格至少要有一个值，值之间不要重复。同一个规格下的值会和其他规格的值两两组合，组合数等于各规格值数量的乘积，所以规格值不要放太多，否则组合会爆炸式增长。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">不需要规格的商品</span>：这里留空即可，商品会按一个「默认规格」的 SKU 上架。
                </div>
              </div>
            </div>
            <SpecEditor v-model:specs="specs" />
          </div>
        </n-card>
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                规格组合（SKU）
              </div>
              <div class="mt-2 text-13 text-gray-400">
                由上方规格自动组合生成。只有「上架」的组合才会保存为 SKU 对外售卖
              </div>
            </div>
          </template>
          <div class="space-y-10">
            <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
              <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
                关于「上架」与「库存」
              </div>
              <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
                <div>
                  系统会把所有规格组合都列出来。已经存在的组合会自动标记为<span class="text-gray-700 font-medium dark:text-gray-200">上架</span>并带上原来的价格和库存；新出现的组合默认<span class="text-gray-700 font-medium dark:text-gray-200">未上架</span>，需要手动打开开关并填写价格、库存。
                </div>
                <div>
                  未上架的组合同步保存后被移除，商城端不可购买。已上架的组合即使库存为 0 也会保留，只是显示售罄。
                </div>
                <div>
                  商品列表里显示的「库存」是这里所有已上架组合库存的总和，由系统自动汇总，不需要单独填写。
                </div>
                <div>
                  <span class="text-gray-700 font-medium dark:text-gray-200">成本价</span>：用于后台核算成本，商城端不会展示、也不影响成交。不填视为 0。
                </div>
              </div>
            </div>
            <SkuGrid v-model:rows="skuRows" :default-price="Number(form.price) || 0" :credit-type="Number(form.credit_type) || 0" />
          </div>
        </n-card>
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div>
              <div class="text-15 font-medium">
                商品图片
              </div>
              <div class="mt-2 text-13 text-gray-400">
                轮播图展示在商品详情页顶部，详情图纵向铺在商品说明下方
              </div>
            </div>
          </template>
          <div class="space-y-16">
            <div>
              <div class="mb-8 flex items-center gap-6">
                <i class="i-fe:image text-14 text-gray-400" />
                <span class="text-13 font-medium">轮播图</span>
                <span class="text-12 text-gray-400 dark:text-gray-500">按顺序横向滑动展示，建议 1:1</span>
              </div>
              <ImageZone v-model:images="carouselImages" :upload="file => api.uploadImage(file, 'carousel')" :sync-oss="api.syncOss" placeholder="请上传或填写轮播图地址" empty-tip="还没有轮播图" />
            </div>
            <n-divider class="m-0!" />
            <div>
              <div class="mb-8 flex items-center gap-6">
                <i class="i-fe:file-text text-14 text-gray-400" />
                <span class="text-13 font-medium">详情图</span>
                <span class="text-12 text-gray-400 dark:text-gray-500">按顺序纵向铺开，建议等宽长图</span>
              </div>
              <ImageZone v-model:images="detailImages" :upload="file => api.uploadImage(file, 'details')" :sync-oss="api.syncOss" placeholder="请上传或填写详情图地址" empty-tip="还没有详情图" />
            </div>
          </div>
        </n-card>
      </div>
      <div class="w-full shrink-0 lg:sticky lg:top-0 lg:w-300 space-y-12">
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div class="text-15 font-medium">
              发布状态
            </div>
          </template>
          <div class="space-y-10">
            <div class="flex items-center justify-between gap-5">
              <div class="shrink-0 text-13 text-gray-500 dark:text-gray-400">
                上架
              </div>
              <n-switch v-model:value="form.enable" />
            </div>
            <div class="flex items-center justify-between gap-5">
              <div class="shrink-0 text-13 text-gray-500 dark:text-gray-400">
                积分类型
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in creditTypeOptions" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="form.credit_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.credit_type = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="flex items-center justify-between gap-5">
              <div class="shrink-0 text-13 text-gray-500 dark:text-gray-400">
                商品类型
              </div>
              <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
                <div v-for="item in productTypeOptions" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="form.product_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.product_type = item.value">
                  {{ item.label }}
                </div>
              </div>
            </div>
            <div class="text-12 text-gray-400 dark:text-gray-500">
              商品类型决定了发货方式<br>
              <b>实体商品</b>：有实际的物品赠予用户，用户下单时需要填写自己的收货地址<br>
              <b>虚拟商品</b>：无实际物品，用户下单时需要填写自己的邮箱地址以便接受虚拟商品
            </div>
            <div class="flex items-center gap-5">
              <n-tooltip>
                <template #trigger>
                  <div class="w-60 shrink-0 text-13 text-gray-500 dark:text-gray-400">
                    展示价
                  </div>
                </template>
                商品列表里显示的价格，具体成交的金额以对应规格中设置的金额为准
              </n-tooltip>
              <n-input-number v-model:value="form.price" class="min-w-0 flex-1" :min="0" :precision="0" placeholder="单位：分">
                <template #suffix>
                  {{ Number(form.credit_type) === 0 ? '星光' : '积分' }}
                </template>
              </n-input-number>
            </div>
            <div class="text-12 text-gray-400 dark:text-gray-500">
              展示价只用于列表展示，实际成交以每个 SKU 的价格为准。
            </div>
            <div class="flex items-center gap-5">
              <n-tooltip>
                <template #trigger>
                  <div class="w-60 shrink-0 text-13 text-gray-500 dark:text-gray-400">
                    已售
                  </div>
                </template>
                通常由订单自动累计，手工改动只影响展示
              </n-tooltip>
              <n-input-number v-model:value="form.sold" class="min-w-0 flex-1" :min="0" :precision="0" placeholder="0" />
            </div>
            <div class="flex items-center gap-5">
              <n-tooltip>
                <template #trigger>
                  <div class="w-60 shrink-0 text-13 text-gray-500 dark:text-gray-400">
                    排序值
                  </div>
                </template>
                数字越大在商城列表里排得越靠前
              </n-tooltip>
              <n-input-number v-model:value="form.sort_order" class="min-w-0 flex-1" :precision="0" placeholder="0" />
            </div>
            <div class="text-12 text-gray-400 dark:text-gray-500">
              排序数字越高，商品展示越靠前
            </div>
          </div>
        </n-card>
        <n-card size="small" :bordered="false" class="border border-gray-200 rounded-3 dark:border-gray-700">
          <template #header>
            <div class="text-15 font-medium">
              库存汇总
            </div>
          </template>
          <div class="space-y-8">
            <div class="flex items-center justify-between">
              <span class="text-13 text-gray-500 dark:text-gray-400">已上架组合</span>
              <span class="text-15 font-medium">
                {{ stockSummary.enabledCount }}
                <span class="text-13 text-gray-400">/ {{ stockSummary.totalCount }}</span>
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-13 text-gray-500 dark:text-gray-400">可售库存</span>
              <span class="text-15 font-medium">{{ stockSummary.stock }}</span>
            </div>
            <div class="text-12 text-gray-400 dark:text-gray-500">
              由所有已上架组合的库存自动汇总，保存时写入商品库存。
            </div>
          </div>
        </n-card>
      </div>
    </div>
    <n-spin v-if="!loaded" class="block w-full py-40" size="large" />
  </CommonPage>
</template>

<script setup>
import { onBeforeRouteLeave } from 'vue-router'
import api from './api'
import ImageZone from './components/ImageZone.vue'
import SkuGrid from './components/SkuGrid.vue'
import SpecEditor from './components/SpecEditor.vue'
import TagEditor from './components/TagEditor.vue'
import { buildCombos, combosCount, MAX_COMBOS, MAX_IMAGES, MAX_SPECS, nextUid, normalizeProps, parseSpecProperties, propsKey, propsLabel, reconcileSkuRows } from './spec-utils'

defineOptions({ name: 'ShopProductDetails' })

const route = useRoute()
const router = useRouter()

const creditTypeOptions = [
  { label: '星光', value: 0 },
  { label: '积分', value: 1 },
]
const productTypeOptions = [
  { label: '虚拟', value: 0 },
  { label: '实体', value: 1 },
]

const form = ref({
  id: 0,
  name: '',
  cover: '',
  price: 0,
  credit_type: 0,
  sold: 0,
  describe: '',
  sort_order: 0,
  enable: true,
  product_type: 0,
})
const tagList = ref([])
const specs = ref([])
const skuRows = ref([])
const carouselImages = ref([])
const detailImages = ref([])

const loaded = ref(false)
// 详情加载失败时为 true：此时表单里是空数据，保存会按「新增」处理，
// 所以要把保存入口藏掉，避免误建出一条重复商品
const loadFailed = ref(false)
const dirty = ref(false)
const saving = ref(false)

const orphanStash = new Map()

const productID = computed(() => Number(route.query.id) || 0)
const pageTitle = computed(() => (productID.value > 0 ? '编辑商品' : '新增商品'))

const stockSummary = computed(() => {
  const enabled = skuRows.value.filter(r => r.enabled)
  return {
    enabledCount: enabled.length,
    totalCount: skuRows.value.length,
    stock: enabled.reduce((sum, r) => sum + (Number(r.stock) || 0), 0),
  }
})

async function load() {
  const id = productID.value
  if (!id) {
    skuRows.value = reconcileSkuRows(buildCombos([]), [], 0).rows
    return
  }

  const res = await api.getDetails(id)
  const data = res.data || {}
  Object.assign(form.value, {
    id: Number(data.id) || id,
    name: data.name ?? '',
    cover: data.cover ?? '',
    price: Number(data.price) || 0,
    credit_type: data.credit_type ?? 0,
    sold: Number(data.sold) || 0,
    describe: data.describe ?? '',
    sort_order: data.sort_order ?? 0,
    enable: data.enable ?? true,
    product_type: data.product_type ?? 0,
  })
  tagList.value = String(data.tags ?? '').split(',').map(t => t.trim()).filter(Boolean)

  specs.value = (data.specs ?? []).map(s => ({
    uid: nextUid('s'),
    key_name: s.key_name ?? '',
    values: (s.values ?? []).map(v => ({ uid: nextUid('v'), value_name: v.value_name ?? '' })),
  }))

  // 历史 SKU 全部先标为上架，交给对账去匹配（组合键 = 规格名 + 规格值名）；
  // 完全匹配上的组合保留上架与原价格库存，没匹配上的组合自然落到「未上架」等管理员处理
  const specKeys = specs.value.map(s => s.key_name.trim()).filter(Boolean)
  const prevRows = (data.skus ?? []).map((sku) => {
    const specProperties = normalizeProps(parseSpecProperties(sku.spec_properties), specKeys)
    const stock = Number(sku.stock) || 0
    return {
      key: propsKey(specProperties),
      uidKey: null,
      specProperties,
      price: Number(sku.price) || 0,
      costPrice: Number(sku.cost_price) || 0,
      stock,
      // 记录载入时的库存，用于判断管理员是否真的改过这个值
      originalStock: stock,
      enabled: true,
    }
  })

  const { rows, orphans } = reconcileSkuRows(buildCombos(specs.value), prevRows, form.value.price)
  skuRows.value = rows

  if (orphans.length)
    $message.warning(`有 ${orphans.length} 个历史 SKU 与当前规格不匹配，保存后将被移除`)

  const images = data.images ?? []
  carouselImages.value = images
    .filter(i => i.type === 0)
    .map(i => ({ uid: nextUid('img'), path: i.image_path ?? '', enabled: i.enable ?? true }))
  detailImages.value = images
    .filter(i => i.type === 1)
    .map(i => ({ uid: nextUid('img'), path: i.image_path ?? '', enabled: i.enable ?? true }))
}

function applyReconcile() {
  const { rows, orphans } = reconcileSkuRows(
    buildCombos(specs.value),
    skuRows.value,
    form.value.price,
    orphanStash,
  )
  skuRows.value = rows

  const lost = orphans.filter(o => o.enabled).length
  if (lost)
    $message.warning(`已移除 ${lost} 个规格组合，保存后对应 SKU 将被删除`)
}

// 规格是否处于「正在编辑」的中间态：新建的规格还没填名字，或某个规格暂时没有有效值。
// 此时组合列表是残缺的，对账会把已填好价格库存的行整批判成孤儿，所以先跳过
function specsEditing() {
  return specs.value.some(s => !s.key_name.trim() || !s.values.some(v => v.value_name.trim()))
}

watch(specs, () => {
  if (!loaded.value || specsEditing())
    return
  applyReconcile()
}, { deep: true })

function isNonNegInt(value) {
  if (value === null || value === undefined || value === '')
    return false
  const n = Number(value)
  return Number.isInteger(n) && n >= 0
}

function isInt(value) {
  if (value === null || value === undefined || value === '')
    return true
  return Number.isInteger(Number(value))
}

function validate() {
  const f = form.value
  if (!f.name.trim())
    return '请输入商品名称'
  if (f.name.trim().length > 255)
    return '商品名称不能超过 255 个字符'
  if (!f.cover.trim())
    return '请上传商品封面'
  if (f.cover.length > 255)
    return '封面地址过长'
  if (!isNonNegInt(f.price))
    return '商品价格必须为不小于 0 的整数（单位：分）'
  if (!isNonNegInt(f.sold))
    return '已售数量必须为不小于 0 的整数'
  if (!isInt(f.sort_order))
    return '排序值必须为整数'
  if (![0, 1].includes(f.credit_type))
    return '请选择积分类型'
  if (![0, 1].includes(f.product_type))
    return '请选择商品类型'
  if (f.describe.length > 1000)
    return '商品描述不能超过 1000 个字符'
  if (tagList.value.join(',').length > 255)
    return '标签总长度不能超过 255 个字符'

  if (specs.value.length > MAX_SPECS)
    return `规格类型不能超过 ${MAX_SPECS} 个`

  const keyNames = []
  for (const spec of specs.value) {
    const key = spec.key_name.trim()
    if (!key)
      return '规格名称不能为空'
    if (key.length > 100)
      return `规格名称「${key}」不能超过 100 个字符`
    if (keyNames.includes(key))
      return `规格名称「${key}」重复`
    keyNames.push(key)

    if (!spec.values.length)
      return `规格「${key}」至少需要添加一个规格值`

    const seen = new Set()
    for (const value of spec.values) {
      const name = value.value_name.trim()
      if (!name)
        return `规格「${key}」存在空的规格值`
      if (name.length > 100)
        return `规格值「${name}」不能超过 100 个字符`
      if (seen.has(name))
        return `规格「${key}」下的规格值「${name}」重复`
      seen.add(name)
    }
  }

  const count = combosCount(specs.value)
  if (count === 0)
    return '每个规格至少需要一个规格值'
  if (count > MAX_COMBOS)
    return `规格组合数（${count}）超过上限 ${MAX_COMBOS}，请减少规格值`

  const enabledRows = skuRows.value.filter(r => r.enabled)
  if (!enabledRows.length)
    return '请至少上架一个规格组合'
  for (const row of enabledRows) {
    const label = propsLabel(row.specProperties)
    if (row.price === null || row.price === undefined || row.price === '')
      return `组合「${label}」未填写价格`
    if (!isNonNegInt(row.price))
      return `组合「${label}」的价格必须为不小于 0 的整数`
    if (row.costPrice !== null && row.costPrice !== '' && !isNonNegInt(row.costPrice))
      return `组合「${label}」的成本价必须为不小于 0 的整数`
    if (!isNonNegInt(row.stock))
      return `组合「${label}」的库存必须为不小于 0 的整数`
  }

  const imageGroups = [[carouselImages.value, '轮播图'], [detailImages.value, '详情图']]
  const imageTotal = carouselImages.value.length + detailImages.value.length
  if (imageTotal > MAX_IMAGES)
    return `商品图片（轮播图 + 详情图）不能超过 ${MAX_IMAGES} 张`
  for (const [list, label] of imageGroups) {
    for (const image of list) {
      if (!String(image.path).trim())
        return `${label}存在空图片，请删除或补充`
      if (image.path.length > 1000)
        return '图片地址过长'
    }
  }

  return ''
}

function buildPayload() {
  const f = form.value
  return {
    id: Number(f.id) || 0,
    name: f.name.trim(),
    cover: f.cover,
    price: Number(f.price) || 0,
    credit_type: f.credit_type,
    sold: Number(f.sold) || 0,
    sort_order: Number(f.sort_order) || 0,
    enable: !!f.enable,
    product_type: f.product_type,
    tags: tagList.value.join(','),
    describe: f.describe,
    specs: specs.value.map(s => ({
      key_name: s.key_name.trim(),
      values: s.values.map(v => ({ value_name: v.value_name.trim() })),
    })),
    skus: skuRows.value
      .filter(r => r.enabled)
      .map((r) => {
        const sku = {
          price: Number(r.price) || 0,
          cost_price: Number(r.costPrice) || 0,
          spec_properties: r.specProperties,
        }
        // 库存没被改过就不提交，交给后端保留库里的最新值，
        // 免得把打开页面之后被订单扣掉的库存又覆盖回去
        const stock = Number(r.stock) || 0
        if (r.originalStock === null || r.originalStock === undefined || stock !== r.originalStock)
          sku.stock = stock
        return sku
      }),
    images: [
      ...carouselImages.value.map((image, index) => ({
        image_path: image.path,
        type: 0,
        sort_order: carouselImages.value.length - index,
        enable: !!image.enabled,
      })),
      ...detailImages.value.map((image, index) => ({
        image_path: image.path,
        type: 1,
        sort_order: detailImages.value.length - index,
        enable: !!image.enabled,
      })),
    ],
  }
}

async function handleSubmit() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  $loadingBar.start()
  try {
    const res = await api.productSave(buildPayload())
    $message.success('保存成功')
    const newID = Number(res?.data?.id) || 0
    if (!productID.value && newID) {
      // 地址带上新增的 ID，页面会按新的 fullPath 重新挂载并重新加载
      dirty.value = false
      router.replace({ query: { id: newID } })
      return
    }
    // 已存在的商品原地重载：让保存期间被订单改动的库存等字段回到真实值，
    // 同时重置库存基线，下一次保存仍然只提交被改过的库存
    orphanStash.clear()
    try {
      await load()
    }
    catch {
      // 重载失败不影响保存结果，错误提示由 http 拦截器统一处理
    }
    await nextTick()
    dirty.value = false
  }
  catch {
    $loadingBar.error()
  }
  finally {
    $loadingBar.finish()
    saving.value = false
  }
}

function handleCancel() {
  router.back()
}

watch([form, tagList, specs, skuRows, carouselImages, detailImages], () => {
  if (loaded.value)
    dirty.value = true
}, { deep: true })

onBeforeRouteLeave(() => {
  if (!dirty.value)
    return true
  return new Promise((resolve) => {
    $dialog.warning({
      title: '提示',
      content: '当前页面有未保存的修改，确定要离开吗？',
      positiveText: '离开',
      negativeText: '继续编辑',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
})

onMounted(async () => {
  try {
    await load()
  }
  catch {
    // 错误提示由 http 拦截器统一处理；标记加载失败，页面不再提供保存入口
    loadFailed.value = true
  }
  loaded.value = true
  // 等加载期间的 watch 回调都跑完，再清掉脏标记，避免刚进来就提示未保存
  await nextTick()
  dirty.value = false
})
</script>
