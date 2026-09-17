<script setup>
import AppDialog from './components/base/AppDialog.vue'
import AppToast from './components/base/AppToast.vue'
import PWABadge from './components/PWABadge.vue'
</script>

<template>
  <div class="app">
    <router-view v-slot="{ Component }">
      <Transition name="page" mode="out-in">
        <component :is="Component" />
      </Transition>
    </router-view>
  </div>
  <PWABadge />
  <AppToast />
  <AppDialog />
</template>

<style scoped></style>
