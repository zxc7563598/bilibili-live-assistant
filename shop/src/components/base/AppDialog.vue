<script setup>
import { cancelClick, confirmClick, state } from '@/utils/dialog'

// 类型 → 图标 + 图标底色
const types = {
  success: { icon: 'success', cls: 'bg-success/15 text-success' },
  fail: { icon: 'fail', cls: 'bg-danger/15 text-danger' },
  warning: { icon: 'warning', cls: 'bg-warning/15 text-warning' },
  info: { icon: 'info', cls: 'bg-info/15 text-info' },
  confirm: { icon: 'warning', cls: 'bg-warning/15 text-warning' },
}
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="state.visible" class="fixed inset-0 z-80 flex items-center justify-center p-4" role="dialog" aria-modal="true">
        <div class="absolute inset-0 bg-black/40" @click="state.maskClosable && cancelClick()" />
        <div class="relative w-full max-w-sm rounded-3xl bg-surface p-6 shadow-pop">
          <div class="flex flex-col items-center text-center">
            <div class="flex h-12 w-12 items-center justify-center rounded-full" :class="types[state.type].cls">
              <AppIcon :name="types[state.type].icon" :size="24" />
            </div>
            <h2 v-if="state.title" class="mt-4 text-base font-semibold text-fg">
              {{ state.title }}
            </h2>
            <p v-if="state.message" class="mt-1.5 whitespace-pre-line text-sm text-fg-2">
              {{ state.message }}
            </p>
          </div>
          <div class="mt-5 flex gap-3">
            <AppButton v-if="state.showCancel" variant="ghost" size="md" class="flex-1" :disabled="state.loading" @click="cancelClick">
              {{ state.cancelText }}
            </AppButton>
            <AppButton v-if="state.showConfirm" :variant="state.confirmVariant" size="md" class="flex-1" :loading="state.loading" @click="confirmClick">
              {{ state.confirmText }}
            </AppButton>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 外层淡入淡出 + 面板缩放位移（同 AppBottomSheet 的双段过渡，无 @keyframes） */
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.2s ease;
}
.dialog-enter-active > div:last-child,
.dialog-leave-active > div:last-child {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
.dialog-enter-from > div:last-child,
.dialog-leave-to > div:last-child {
  transform: scale(0.92) translateY(8px);
}
</style>
