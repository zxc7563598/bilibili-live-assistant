<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="800">
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" />
      </MeQueryItem>
    </MeCrud>
  </CommonPage>

  <n-modal v-model:show="showMonthlyModal" title="每日分析" preset="card" style="width: 840px" :mask-closable="false">
    <n-spin :show="monthlyLoading">
      <n-calendar v-model:value="todayDate" #="{ year, month, date }" :on-panel-change="handleUpdateValue" class="mc-calendar">
        <div v-if="isPanelCell(year, month)" class="mc-cell">
          <span class="mc-live-dot" :class="monthlyInfo.live_days[date] ? 'is-live' : ''" :title="monthlyInfo.live_days[date] ? '已开播' : '未开播'" />
          <div v-if="getCellData(date).length" class="mc-stats">
            <span v-for="item in getCellData(date)" :key="item.key" class="mc-stat" :class="`is-${item.key}`" :title="item.title">
              <i :class="item.icon" />
              {{ item.label }}
            </span>
          </div>
        </div>
      </n-calendar>
    </n-spin>
  </n-modal>

  <n-modal v-model:show="showDanmuModal" title="弹幕分析" preset="card" style="width: 840px" :mask-closable="false">
    <n-spin :show="danmuLoading">
      <div class="mc-danmu">
        <div class="mc-wordcloud">
          <Vue3WordCloud :words="defaultWords" :color="colorHandler" :font-size-ratio="4" font-family="PingFang SC, Microsoft YaHei, sans-serif">
            <template #default="{ text, weight }">
              <span :title="`${text} × ${weight}`" style="cursor: pointer">
                {{ text }}
              </span>
            </template>
          </Vue3WordCloud>
        </div>

        <div v-if="hasReport" class="mc-report">
          <div class="mc-report-sec">
            <div class="mc-report-title">
              高频词
            </div>
            <div class="mc-chips">
              <span v-for="item in topWords" :key="item.word" class="mc-chip">
                {{ item.word }}
                <em>{{ item.count }}</em>
              </span>
            </div>
          </div>
          <div class="mc-report-sec">
            <div class="mc-report-title">
              高频词组
            </div>
            <div class="mc-chips">
              <span v-for="item in topPhrases" :key="item.word" class="mc-chip mc-chip-phrase">
                {{ item.word }}
                <em>{{ item.count }}</em>
              </span>
            </div>
          </div>
          <div class="mc-report-sec">
            <div class="mc-report-title">
              高频句子
            </div>
            <ol class="mc-sentences">
              <li v-for="(item, index) in topMessages" :key="item.word" :data-rank="index + 1">
                <span class="mc-sentence-text">「{{ item.word }}」</span>
                <em>{{ item.count }} 次</em>
              </li>
            </ol>
          </div>
        </div>

        <n-empty v-else-if="!danmuLoading" description="暂无弹幕数据" size="small" style="padding: 24px 0" />
      </div>
    </n-spin>
  </n-modal>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import api from './api'
import Vue3WordCloud from './vue3-word-cloud'

const $table = ref(null)
const columns = [
  { title: '用户UID', key: 'uid', minWidth: 170 },
  { title: '用户昵称', key: 'uname', minWidth: 120 },
  { title: '积分', key: 'points', width: 100 },
  { title: '星光', key: 'stars', width: 100 },
  { title: '发送弹幕', key: 'total_danmu_count', width: 100 },
  { title: '礼物金额', key: 'total_gift_amount', width: 100, render(row) {
    const total_gift_amount = (row.total_gift_amount / 100)
    return `¥${total_gift_amount.toFixed(2)}`
  } },
  {
    title: '操作',
    key: 'actions',
    width: 240,
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
            onClick: () => getMonthly(row.uid),
          },
          {
            default: () => '每日分析',
            icon: () => h('i', { class: 'i-fe:calendar text-14' }),
          },
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            style: 'margin-left: 12px;',
            onClick: () => getDanmu(row.uid),
          },
          {
            default: () => '弹幕分析',
            icon: () => h('i', { class: 'i-fe:bar-chart text-14' }),
          },
        ),
      ]
    },
  },
]

const query = ref({
  uid: null,
  uname: null,
})

const showMonthlyModal = ref(false)
const todayDate = ref(Date.now())
const monthlyLoading = ref(false)
const monthlyByUID = ref(0)
const monthlyInfo = ref({
  danmu_count: [],
  gift_amount: [],
  gift_count: [],
  live_days: [],
})
// 当前日历面板展示的年月，用于过滤前后月的补位单元格
const panelYM = ref({
  year: new Date(todayDate.value).getFullYear(),
  month: new Date(todayDate.value).getMonth() + 1,
})

function formatCompact(n) {
  if (n >= 100000000)
    return `${+(n / 100000000).toFixed(1)}亿`
  if (n >= 10000)
    return `${+(n / 10000).toFixed(1)}万`
  return String(n)
}

function formatYuan(cents) {
  const v = cents / 100
  if (v >= 10000)
    return formatCompact(v)
  return Number.isInteger(v) ? String(v) : String(Math.round(v * 100) / 100)
}

function isPanelCell(year, month) {
  return panelYM.value.year === year && panelYM.value.month === month
}

function getCellData(date) {
  const data = monthlyInfo.value
  const items = []
  if (data.gift_amount?.[date]) {
    const yuan = formatYuan(data.gift_amount[date])
    items.push({
      key: 'amount',
      icon: 'i-fe:hand-money-yuan',
      label: `¥${yuan}`,
      title: `礼物金额 ¥${yuan}`,
    })
  }
  if (data.danmu_count?.[date]) {
    items.push({
      key: 'danmu',
      icon: 'i-fe:message-square',
      label: formatCompact(data.danmu_count[date]),
      title: `弹幕 ${data.danmu_count[date]} 条`,
    })
  }
  if (data.gift_count?.[date]) {
    items.push({
      key: 'gift',
      icon: 'i-fe:gift',
      label: formatCompact(data.gift_count[date]),
      title: `礼物 ${data.gift_count[date]} 个`,
    })
  }
  return items
}

function getMonthly(uid) {
  showMonthlyModal.value = true
  const year = new Date(todayDate.value).getFullYear()
  const month = new Date(todayDate.value).getMonth() + 1
  monthlyByUID.value = uid
  panelYM.value = { year, month }
  loadMonthly(uid, year, month)
}

function handleUpdateValue({ year, month }) {
  panelYM.value = { year, month }
  loadMonthly(monthlyByUID.value, year, month)
}

function loadMonthly(uid, year, month) {
  monthlyInfo.value = {
    danmu_count: [],
    gift_amount: [],
    gift_count: [],
    live_days: [],
  }
  monthlyLoading.value = true
  api.getMonthlyByUID({ uid, year, month })
    .then((res) => {
      if (monthlyByUID.value === uid && panelYM.value.year === year && panelYM.value.month === month)
        Object.assign(monthlyInfo.value, res.data)
    })
    .catch(() => {})
    .finally(() => {
      monthlyLoading.value = false
    })
}

const showDanmuModal = ref(false)
const danmuLoading = ref(false)
const defaultWords = ref([])
const myColors = ['#1f77b4', '#629fc9', '#94bedb', '#c9e0ef']
function colorHandler(_word, index) {
  return myColors[index % myColors.length]
}
// 接口返回的四个维度原始数据，用于词云下方的文字报告
const danmuReport = ref({
  words: [],
  bigrams: [],
  trigrams: [],
  messages: [],
})
const hasReport = computed(() =>
  ['words', 'bigrams', 'trigrams', 'messages'].some(key => danmuReport.value[key].length > 0),
)
const topWords = computed(() => sortByCount(danmuReport.value.words).slice(0, 10))
const topPhrases = computed(() =>
  sortByCount([...danmuReport.value.bigrams, ...danmuReport.value.trigrams]).slice(0, 10),
)
const topMessages = computed(() => sortByCount(danmuReport.value.messages).slice(0, 8))
function sortByCount(list) {
  return [...list].sort((a, b) => b.count - a.count)
}

function getDanmu(uid) {
  showDanmuModal.value = true
  danmuLoading.value = true
  defaultWords.value = []
  danmuReport.value = { words: [], bigrams: [], trigrams: [], messages: [] }
  api.getDanmuByUID({ uid }).then((res) => {
    const { words = [], bigrams = [], trigrams = [], messages = [] } = res.data
    danmuReport.value = { words, bigrams, trigrams, messages }
    // 词云按四个维度合并去重，同名取最大出现次数
    const byText = new Map()
    ;[words, bigrams, trigrams, messages].forEach((group) => {
      group.forEach((item) => {
        if (!byText.has(item.word) || byText.get(item.word) < item.count)
          byText.set(item.word, item.count)
      })
    })
    defaultWords.value = [...byText.entries()]
  }).catch(() => {}).finally(() => {
    danmuLoading.value = false
  })
}

onMounted(() => {
  $table.value?.handleSearch()
})
</script>

<style>
.mc-calendar.n-calendar {
  height: 720px;
}

.mc-calendar .n-calendar-cell {
  padding: 15px 10px;
}

.mc-calendar .n-calendar-date {
  padding-bottom: 0.3em;
  width: 85%;
}

.mc-live-dot {
  position: absolute;
  top: 15px;
  right: 10px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background-color: #cfd3d8;
  flex: none;
}

.mc-live-dot.is-live {
  background-color: #18a058;
  box-shadow: 0 0 0 3px rgba(24, 160, 88, 0.15);
}

.mc-stats {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  margin-top: 4px;
}

.mc-stat {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 100%;
  padding: 2px 7px 2px 3px;
  border-radius: 999px;
  font-size: 10px;
  line-height: 1.3;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mc-stat i {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 13px;
  height: 13px;
  border-radius: 50%;
  color: #fff;
  font-size: 9px;
  flex: none;
}

.mc-stat.is-amount {
  font-size: 11px;
  font-weight: 700;
  background: #fef1e0;
  color: #b45309;
}

.mc-stat.is-amount i {
  background: #f59e0b;
}

.mc-stat.is-danmu {
  background: #eaf2fe;
  color: #2563eb;
}

.mc-stat.is-danmu i {
  background: #3b82f6;
}

.mc-stat.is-gift {
  background: #f2edfe;
  color: #7c3aed;
}

.mc-stat.is-gift i {
  background: #8b5cf6;
}

.dark .mc-live-dot {
  background-color: #565b61;
}

.dark .mc-stat.is-amount {
  background: rgba(245, 158, 11, 0.16);
  color: #fbbf24;
}

.dark .mc-stat.is-amount i {
  background: #f59e0b;
}

.dark .mc-stat.is-danmu {
  background: rgba(59, 130, 246, 0.16);
  color: #93c5fd;
}

.dark .mc-stat.is-danmu i {
  background: #3b82f6;
}

.dark .mc-stat.is-gift {
  background: rgba(139, 92, 246, 0.16);
  color: #c4b5fd;
}

.dark .mc-stat.is-gift i {
  background: #8b5cf6;
}

/* 弹幕词云：组件自身高度 100%，由父容器撑高 */
.mc-danmu {
  max-height: 70vh;
  overflow-y: auto;
}

.mc-wordcloud {
  height: 340px;
  width: 100%;
}

.mc-report {
  margin-top: 18px;
  padding-top: 4px;
  border-top: 1px dashed rgba(128, 128, 128, 0.35);
}

.mc-report-sec {
  margin-top: 16px;
}

.mc-report-title {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.mc-report-title::before {
  content: '';
  width: 4px;
  height: 14px;
  border-radius: 2px;
  background: #18a058;
  flex: none;
}

.mc-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mc-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1.4;
  background: #eaf2fe;
  color: #2563eb;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.mc-chip-phrase {
  background: #f2edfe;
  color: #7c3aed;
  border-color: rgba(139, 92, 246, 0.2);
}

.mc-chip em {
  font-style: normal;
  font-size: 10px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  opacity: 0.75;
}

.mc-sentences {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.mc-sentences li {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  background: #f7f8fa;
  font-size: 13px;
  line-height: 1.5;
}

.mc-sentences li::before {
  content: attr(data-rank);
  flex: none;
  min-width: 18px;
  padding: 0 4px;
  border-radius: 999px;
  background: #eef0f3;
  color: #8a919f;
  font-size: 11px;
  font-weight: 700;
  text-align: center;
}

.mc-sentence-text {
  flex: 1;
  min-width: 0;
  word-break: break-all;
}

.mc-sentences li em {
  flex: none;
  font-style: normal;
  font-weight: 700;
  color: #b45309;
  font-variant-numeric: tabular-nums;
}

.dark .mc-report-title {
  color: #e5e7eb;
}

.dark .mc-chip {
  background: rgba(59, 130, 246, 0.14);
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.3);
}

.dark .mc-chip-phrase {
  background: rgba(139, 92, 246, 0.14);
  color: #c4b5fd;
  border-color: rgba(139, 92, 246, 0.3);
}

.dark .mc-sentences li {
  background: #25272b;
}

.dark .mc-sentences li::before {
  background: #34373c;
  color: #9aa1ab;
}
</style>
