<script setup>
import AppToast from './components/base/AppToast.vue'
import PWABadge from './components/PWABadge.vue'
</script>

<template>
  <div class="app">
    <router-view />
  </div>
  <PWABadge />
  <AppToast />
</template>

<style scoped></style>
