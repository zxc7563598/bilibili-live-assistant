<!-- Copyright © 2023 Ronnie Zhang (大脸怪). MIT License. -->

<template>
  <AppCard class="flex items-center px-12" border-b="1px solid light_border dark:dark_border">
    <MenuCollapse />

    <AppTab class="w-0 flex-1 px-12" />

    <span class="mx-6 opacity-20">|</span>

    <div class="flex flex-shrink-0 items-center px-12 text-18">
      <ToggleTheme />

      <Fullscreen />

      <ThemeSetting class="mr-16" />

      <UserAvatar />
    </div>
  </AppCard>
</template>

<script setup>
import { ThemeSetting, ToggleTheme } from '@/components'
import { AppTab, Fullscreen, MenuCollapse, UserAvatar } from '@/layouts/components'
</script>
