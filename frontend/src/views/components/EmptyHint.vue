<template>
  <div class="empty">
    <div v-if="icon" class="empty__icon">
      <el-icon :size="22"><component :is="icon" /></el-icon>
    </div>
    <div class="empty__text">{{ text }}</div>
    <slot />
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon?: Component
  text: string
}>()
</script>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 32px 16px;
  color: var(--color-text-secondary);
}

.empty__icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--color-bg-hover);
  color: var(--color-text-placeholder);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.empty__text {
  font-size: 12.5px;
  color: var(--color-text-secondary);
}
</style>
