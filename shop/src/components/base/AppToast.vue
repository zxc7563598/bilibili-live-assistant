<script setup>
import { computed } from 'vue'
import { close, toasts } from '@/utils/toast'

// 顶部：新消息在最上方向下堆；底部：新消息在最下方向上堆
const groups = computed(() => [
  {
    name: 'toast-top',
    items: toasts.filter(t => t.position === 'top'),
    pad: 'safe-top pt-4',
    stack: 'flex flex-col-reverse items-center gap-2',
  },
  {
    name: 'toast-bottom',
    items: toasts.filter(t => t.position === 'bottom'),
    pad: 'safe-bottom pb-4',
    stack: 'flex flex-col items-center gap-2',
  },
])

// 类型 → 图标 + 图标底色（Tailwind，日/夜自动适配）
const types = {
  success: { icon: 'success', cls: 'bg-success/12 text-success' },
  fail: { icon: 'fail', cls: 'bg-danger/12 text-danger' },
  warning: { icon: 'warning', cls: 'bg-warning/12 text-warning' },
  info: { icon: 'info', cls: 'bg-info/12 text-info' },
}
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed inset-0 z-[100] flex flex-col justify-between">
      <div v-for="g in groups" :key="g.name" class="px-4" :class="g.pad">
        <TransitionGroup :name="g.name" tag="div" :class="g.stack">
          <div v-for="t in g.items" :key="t.id" role="status" class="pointer-events-auto inline-flex max-w-[85vw] items-center gap-2.5 rounded-full border border-line-strong bg-surface py-2.5 pl-3 pr-4 shadow-(--shadow-pop)">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full" :class="types[t.type].cls">
              <AppIcon :name="types[t.type].icon" :size="14" />
            </span>
            <span class="min-w-0 text-sm font-medium text-fg">{{ t.message }}</span>
            <button v-if="t.closable || t.duration === 0" type="button" aria-label="关闭提示" class="-mr-1.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-fg-3 transition-colors hover:text-fg" @click="close(t.id)">
              <AppIcon name="close" :size="14" />
            </button>
          </div>
        </TransitionGroup>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* 顶部：从屏幕上方滑入/滑出 */
.toast-top-enter-active,
.toast-top-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}
.toast-top-enter-from,
.toast-top-leave-to {
  opacity: 0;
  transform: translateY(-100%);
}
.toast-top-move {
  transition: transform 0.3s ease;
}

/* 底部：从屏幕下方滑入/滑出 */
.toast-bottom-enter-active,
.toast-bottom-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}
.toast-bottom-enter-from,
.toast-bottom-leave-to {
  opacity: 0;
  transform: translateY(100%);
}
.toast-bottom-move {
  transition: transform 0.3s ease;
}
</style>
