<!-- Copyright © 2023 Ronnie Zhang (大脸怪). MIT License. -->
<template>
  <AppPage show-footer>
    <div class="flex gap-12">
      <n-card class="flex-1" :bordered="false">
        <div class="flex items-center">
          <n-avatar round :size="64" :src="userStore.avatar" class="flex-shrink-0" />
          <div class="ml-16">
            <div class="text-20 font-semibold">
              你好，{{ userStore.nickName ?? userStore.username }}
            </div>
            <div class="mt-4 text-14 opacity-60">
              当前角色：{{ userStore.currentRole?.name }}
            </div>
          </div>
        </div>

        <div class="mt-16 text-14 leading-relaxed opacity-70">
          欢迎回到 GoAdminKit 控制台。这里是你的系统中枢，你可以在这里快速进入各项功能，
          管理系统、配置权限、构建属于你自己的后台世界。
        </div>
      </n-card>

      <!-- 右侧：项目信息 -->
      <n-card class="w-320" :bordered="false">
        <div class="text-16 font-semibold">
          GoAdminKit
        </div>
        <div class="mt-8 text-13 opacity-60">
          开箱即用的全栈后台系统
        </div>

        <div class="mt-16 text-12 leading-relaxed opacity-60">
          基于 Go + Vue3 构建，将前后端整合为单一二进制部署，
          简化运维成本，专注业务本身。
        </div>

        <div class="mt-16 flex flex-wrap gap-8">
          <n-tag size="small">
            Go
          </n-tag>
          <n-tag size="small">
            Vue3
          </n-tag>
          <n-tag size="small">
            RBAC
          </n-tag>
          <n-tag size="small">
            JWT
          </n-tag>
        </div>
      </n-card>
    </div>
    <div class="grid grid-cols-4 mt-12 gap-12">
      <n-card
        v-for="item in links" :key="item.title" hoverable class="cursor-pointer transition hover:shadow-md"
        @click="open(item.link)"
      >
        <div class="text-14 font-medium">
          {{ item.title }}
        </div>
        <div class="mt-6 text-12 opacity-60">
          {{ item.desc }}
        </div>
      </n-card>
    </div>
    <div class="mt-12 flex gap-12">
      <!-- 项目介绍 -->
      <n-card class="flex-1" title="📦 项目介绍" segmented>
        <div class="text-14 leading-relaxed opacity-80">
          GoAdminKit 是一个为“快速落地后台系统”而设计的全栈解决方案。
          <br><br>
          它将 Go 后端与 Vue3 前端整合在同一个仓库中，并在构建阶段将前端资源嵌入到后端，
          最终只需交付一个二进制文件即可运行完整后台系统。
          <br><br>
          相比传统前后端分离部署方案，大幅降低部署复杂度与运维成本，
          非常适合中小型项目与内部系统。
        </div>
        <div class="grid grid-cols-2 mt-16 gap-8 text-13">
          <div>✔ RBAC 权限控制</div>
          <div>✔ 分层架构</div>
          <div>✔ JWT 鉴权</div>
          <div>✔ 统一错误码设计</div>
          <div>✔ 完整后台管理界面</div>
          <div>✔ 高扩展性</div>
        </div>
      </n-card>
      <!-- 技术栈 -->
      <n-card class="w-400" title="🛠 技术栈" segmented>
        <VChart :option="skillOption" autoresize />
      </n-card>
    </div>
    <n-card class="mt-12" title="📈 随便放个折线图" segmented>
      <div class="h-320">
        <VChart :option="trendOption" autoresize />
      </div>
    </n-card>
  </AppPage>
</template>

<script setup>
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { useUserStore } from '@/store'

const userStore = useUserStore()

echarts.use([
  TooltipComponent,
  LegendComponent,
  GridComponent,
  PieChart,
  LineChart,
  CanvasRenderer,
])

function open(url) {
  window.open(url)
}

const links = [
  {
    title: '项目文档',
    desc: '查看完整开发文档',
    link: 'https://hejunjie.life/gak',
  },
  {
    title: 'GitHub',
    desc: '源码仓库',
    link: 'https://github.com/zxc7563598/GoAdminKit',
  },
  {
    title: 'AI解析',
    desc: '结构与代码解读',
    link: 'https://zread.ai/zxc7563598/GoAdminKit',
  },
  {
    title: '个人网站',
    desc: '何俊杰',
    link: 'https://hejunjie.life',
  },
]

const skillOption = {
  tooltip: {
    trigger: 'item',
    formatter({ name, value }) {
      return `${name} ${value}%`
    },
  },
  legend: {
    left: 'center',
  },
  series: [
    {
      bottom: '12%',
      type: 'pie',
      radius: ['35%', '90%'],
      avoidLabelOverlap: true,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2,
      },
      label: {
        show: false,
        position: 'center',
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 36,
          fontWeight: 'bold',
        },
      },
      labelLine: {
        show: false,
      },
      data: [
        { value: 40, name: 'Vue' },
        { value: 30, name: 'Go' },
        { value: 20, name: 'JavaScript' },
        { value: 10, name: 'Other' },
      ],
    },
  ],
}

const trendOption = {
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L'],
  },
  yAxis: { type: 'value' },
  series: [
    {
      type: 'line',
      data: [120, 200, 150, 280, 300, 450, 120, 200, 150, 280, 671, 883],
      smooth: true,
    },
  ],
}
</script>
