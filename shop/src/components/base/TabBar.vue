<template>
  <nav class="glass-strong safe-bottom fixed inset-x-0 bottom-0 z-50">
    <div class="mx-auto flex w-full max-w-5xl items-stretch">
      <router-link v-for="t in tabs" :key="t.name" :to="t.path" class="flex flex-1 flex-col items-center gap-0.5 py-2.5 text-xs font-medium transition-colors" :class="route.name === t.name ? 'text-primary' : 'text-fg-3'">
        <AppIcon :name="t.icon" :size="24" />
        <span>{{ t.label }}</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup>
import { useRoute } from 'vue-router'

const tabs = [
  { path: '/', name: 'home', label: '首页', icon: 'home' },
  { path: '/profile', name: 'profile', label: '我的', icon: 'user' },
]

const route = useRoute()
</script>
