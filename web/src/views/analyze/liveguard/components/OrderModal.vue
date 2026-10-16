<template>
  <n-modal v-model:show="visible" title="帮用户下单" preset="card" style="width: 720px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「帮用户下单」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            与<span class="text-gray-700 font-medium dark:text-gray-200">用户自己下单完全一致</span>的流程：锁定库存、扣积分/星光、生成订单并快照收货地址，相当于模拟他自己兑换一次。
          </div>
          <div>
            商品有大航海档位限购时按该用户<span class="text-gray-700 font-medium dark:text-gray-200">当前身份</span>判定：历史月份是总督、现在已经过期的用户可能买不了限购商品。
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          用户
        </div>
        <div class="text-13 text-gray-700 dark:text-gray-300">
          {{ form.uname }}（UID {{ form.uid }}）
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          商品
        </div>
        <n-select v-model:value="form.product_id" filterable remote :loading="productLoading" :options="productOptions" placeholder="输入商品名称搜索，只列启用中的商品" class="w-full" @search="searchProductsDebounced" @update:value="handleProductChange" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          规格
        </div>
        <n-select v-model:value="form.sku_id" :options="skuOptions" :disabled="!form.product_id" placeholder="请选择规格" class="w-full" @update:value="handleSkuChange" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          数量
        </div>
        <n-input-number v-model:value="form.count" :min="1" :max="maxCount" :show-button="false" :precision="0" placeholder="兑换数量" class="w-full" />
        <div v-if="selectedSku" class="shrink-0 text-12 text-gray-400 dark:text-gray-500">
          该规格库存 {{ selectedSku.stock }}
        </div>
      </div>
      <div class="flex items-start gap-5">
        <div class="w-110 shrink-0 pt-4 text-right text-13 text-gray-500 dark:text-gray-400">
          收货地址
        </div>
        <div class="min-w-0 flex-1">
          <div v-if="!form.product_id" class="text-13 text-gray-400 dark:text-gray-500">
            先选商品，再按商品类型取该用户的地址
          </div>
          <div v-else-if="addressLoading" class="text-13 text-gray-400 dark:text-gray-500">
            地址加载中…
          </div>
          <template v-else>
            <n-radio-group v-if="addresses.length" v-model:value="form.address_id" class="flex flex-col gap-6">
              <n-radio v-for="item in addresses" :key="item.id" :value="item.id">
                <span class="text-13 text-gray-700 dark:text-gray-300">{{ addressLabel(item) }}</span>
              </n-radio>
            </n-radio-group>
            <div v-else class="text-13 text-gray-400 dark:text-gray-500">
              该用户没有{{ addressTypeLabel }}地址，请新建一个
            </div>
          </template>
          <div class="mt-8">
            <NButton size="small" secondary type="primary" :disabled="!form.product_id" @click="toggleAddressForm">
              {{ showAddressForm ? '收起新建地址' : '新建地址' }}
            </NButton>
          </div>
        </div>
      </div>
      <div v-if="showAddressForm" class="border border-gray-200 rounded-4 px-10 py-8 space-y-8 dark:border-gray-700">
        <div class="text-12 text-gray-500 dark:text-gray-400">
          新建的地址会存进该用户的地址簿，类型与商品一致（{{ addressTypeLabel }}），不影响他已设置的默认地址。
        </div>
        <div class="flex items-center gap-5">
          <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            收件人
          </div>
          <n-input v-model:value="addressForm.name" type="text" :maxlength="50" placeholder="收件人姓名" />
        </div>
        <template v-if="addressType === ADDRESS_TYPE_ACTUAL">
          <div class="flex items-center gap-5">
            <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
              手机号
            </div>
            <n-input v-model:value="addressForm.phone" type="text" :maxlength="20" placeholder="收件人手机号" />
          </div>
          <div class="flex items-center gap-5">
            <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
              所在地区
            </div>
            <RegionCascader v-model="addressForm.regionCodes" />
          </div>
          <div class="flex items-center gap-5">
            <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
              详细地址
            </div>
            <n-input v-model:value="addressForm.detail" type="text" :maxlength="200" placeholder="街道、门牌号等" />
          </div>
        </template>
        <div v-else class="flex items-center gap-5">
          <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            电子邮箱
          </div>
          <n-input v-model:value="addressForm.email" type="text" :maxlength="100" placeholder="虚拟商品发到这里" />
        </div>
        <div class="flex items-center gap-5">
          <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            设为默认
          </div>
          <n-switch v-model:value="addressForm.isDefault" />
        </div>
        <div class="flex justify-end">
          <NButton size="small" type="primary" :loading="addressSaving" @click="saveAddress">
            保存地址
          </NButton>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" :loading="saving" @click="save">
          确认下单
        </NButton>
        <NButton strong secondary class="ml-10" :disabled="saving" @click="close">
          取消
        </NButton>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { debounce, getOptionsLabel } from '@/utils'
import RegionCascader from '@/views/order/list/components/RegionCascader.vue'
import { parseSpecProperties, propsLabel } from '@/views/shop/product/spec-utils'
import api from '../api'

const emit = defineEmits(['success'])

// 与后端 enum 保持一致：地址类型 0 虚拟 / 1 实体，商品类型的数值刻意与它对齐
const ADDRESS_TYPE_VIRTUAL = 0
const ADDRESS_TYPE_ACTUAL = 1

// 与后端 enum.CreditType 一致，商品价格是积分/星光数量，不是金额
const creditTypeOptions = [
  { label: '星光', value: 0 },
  { label: '积分', value: 1 },
]

const visible = ref(false)
const saving = ref(false)
const addressSaving = ref(false)
const productLoading = ref(false)
const addressLoading = ref(false)

const productOptions = ref([])
const skuOptions = ref([])
const addresses = ref([])
const showAddressForm = ref(false)

function initAddressForm() {
  return { name: '', phone: '', regionCodes: [], detail: '', email: '', isDefault: false }
}

function initForm() {
  return {
    uid: null,
    uname: '',
    product_id: null,
    sku_id: null,
    count: 1,
    address_id: null,
  }
}

const form = ref(initForm())
const addressForm = ref(initAddressForm())
// 商品详情拿到的 SKU 列表，用来限制数量上限与展示库存
const product = ref(null)

const addressType = computed(() => product.value?.product_type ?? ADDRESS_TYPE_ACTUAL)
const addressTypeLabel = computed(() => (addressType.value === ADDRESS_TYPE_ACTUAL ? '实体' : '虚拟'))
const selectedSku = computed(() => (product.value?.skus || []).find(item => item.id === form.value.sku_id))
const maxCount = computed(() => Math.max(1, Number(selectedSku.value?.stock || 1)))

// row 来自名单列表，要的是 uid / uname
function open(row) {
  form.value = initForm()
  form.value.uid = row.uid
  form.value.uname = row.uname
  addressForm.value = initAddressForm()
  product.value = null
  skuOptions.value = []
  addresses.value = []
  showAddressForm.value = false
  visible.value = true
  searchProducts('')
}

function close() {
  visible.value = false
}

// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
  addressForm.value = initAddressForm()
  productOptions.value = []
  skuOptions.value = []
  addresses.value = []
  product.value = null
  showAddressForm.value = false
}

// 商品下拉走远程搜索，不认识缓存的旧结果：只认最后一次请求的返回
let productSeq = 0

async function searchProducts(name) {
  const seq = ++productSeq
  productLoading.value = true
  try {
    const res = await api.productList({ pageNo: 1, pageSize: 50, name: name || null, enable: 1 })
    if (seq !== productSeq)
      return
    productOptions.value = (res?.data?.pageData || []).map(item => ({
      label: `${item.name}（${item.price}${getOptionsLabel(creditTypeOptions, item.credit_type)}）`,
      value: item.id,
    }))
  }
  catch {
    if (seq === productSeq)
      productOptions.value = []
  }
  finally {
    if (seq === productSeq)
      productLoading.value = false
  }
}

// 停止输入后再查，避免每敲一个字打一次接口
const searchProductsDebounced = debounce(searchProducts, 300)

async function handleProductChange(productID) {
  form.value.sku_id = null
  form.value.count = 1
  form.value.address_id = null
  product.value = null
  skuOptions.value = []
  addresses.value = []
  showAddressForm.value = false
  if (!productID)
    return
  try {
    const res = await api.productDetails(productID)
    product.value = res?.data ?? null
  }
  catch {
    // 提示由 http 拦截器统一处理
    return
  }
  skuOptions.value = (product.value?.skus || []).map((item) => {
    const props = propsLabel(parseSpecProperties(item.spec_properties))
    const price = `${item.price}${getOptionsLabel(creditTypeOptions, product.value?.credit_type)}`
    return {
      label: [props, price, `库存 ${item.stock}`].filter(Boolean).join(' · '),
      value: item.id,
      disabled: Number(item.stock) <= 0,
      stock: item.stock,
    }
  })
  // 只有一个规格就替他选上，省一次点击
  const first = skuOptions.value.find(item => !item.disabled)
  if (first)
    form.value.sku_id = first.value
  loadAddresses()
}

// 数量上限随所选规格的库存走，换规格时把超出部分拉回来
function handleSkuChange() {
  if (form.value.count > maxCount.value)
    form.value.count = maxCount.value
}

async function loadAddresses() {
  addressLoading.value = true
  try {
    const res = await api.userAddressList({ uid: form.value.uid, type: addressType.value })
    addresses.value = res?.data?.list || []
    const picked = addresses.value.find(item => Number(item.is_default) === 1) || addresses.value[0]
    form.value.address_id = picked ? picked.id : null
  }
  catch {
    addresses.value = []
    form.value.address_id = null
  }
  finally {
    addressLoading.value = false
  }
}

function addressLabel(item) {
  const head = `${item.name}${Number(item.is_default) === 1 ? '（默认）' : ''}`
  if (Number(item.type) === ADDRESS_TYPE_VIRTUAL)
    return `${head} ${item.email}`
  return `${head} ${item.phone} ${item.region || ''} ${item.detail}`
}

function toggleAddressForm() {
  showAddressForm.value = !showAddressForm.value
  if (showAddressForm.value)
    addressForm.value = initAddressForm()
}

function validateAddressForm() {
  const { name, phone, regionCodes, detail, email } = addressForm.value
  if (!String(name ?? '').trim())
    return '请填写收件人姓名'
  if (addressType.value === ADDRESS_TYPE_ACTUAL) {
    if (!String(phone ?? '').trim())
      return '请填写手机号'
    if (!regionCodes.length)
      return '请选择所在地区'
    if (!String(detail ?? '').trim())
      return '请填写详细地址'
  }
  else if (!String(email ?? '').trim()) {
    return '请填写电子邮箱'
  }
  return ''
}

async function saveAddress() {
  const message = validateAddressForm()
  if (message)
    return $message.warning(message)

  addressSaving.value = true
  try {
    const { name, phone, regionCodes, detail, email, isDefault } = addressForm.value
    const res = await api.userAddressSave({
      uid: form.value.uid,
      type: addressType.value,
      name: String(name).trim(),
      phone: String(phone ?? '').trim(),
      // 后端按 JSON 数组串存，region 文案由它派生，不用传
      region_code: JSON.stringify(regionCodes),
      detail: String(detail ?? '').trim(),
      email: String(email ?? '').trim(),
      is_default: isDefault ? 1 : 0,
    })
    const newID = res?.data
    await loadAddresses()
    if (newID)
      form.value.address_id = newID
    showAddressForm.value = false
    addressForm.value = initAddressForm()
    $message.success('地址已保存', { key: 'liveguard.saveAddress' })
  }
  catch {
    // 提示由 http 拦截器统一处理，表单保持原样让管理员改
  }
  finally {
    addressSaving.value = false
  }
}

function validate() {
  if (!form.value.product_id)
    return '请选择商品'
  if (!form.value.sku_id)
    return '请选择规格'
  if (!Number.isInteger(Number(form.value.count)) || Number(form.value.count) < 1)
    return '请填写正确的数量'
  if (!form.value.address_id)
    return '请选择收货地址'
  return ''
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.placeForUser({
      uid: form.value.uid,
      sku_id: Number(form.value.sku_id),
      count: Number(form.value.count),
      address_id: Number(form.value.address_id),
    })
    $message.success('下单成功', { key: 'liveguard.placeForUser' })
    close()
    emit('success')
  }
  catch {
    // 失败提示由 http 拦截器统一处理；服务端会把锁住的库存放回去，这里保持弹窗让管理员改参数重试
  }
  finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>
