<template>
  <div class="usage">
    <div class="usage__bar" :title="title">
      <div
        class="usage__fill"
        :class="toneClass"
        :style="{ width: `${pct}%` }"
      ></div>
    </div>
    <div class="usage__text">
      <span class="usage__used">{{ used ?? 0 }}</span>
      <span class="usage__sep">/</span>
      <span class="usage__limit">{{ limitLabel }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  used?: number | null
  limit?: number | null
}>()

const pct = computed(() => {
  const u = Number(props.used ?? 0)
  const l = Number(props.limit ?? 0)
  if (!l) return 0
  return Math.min(Math.round((u / l) * 100), 100)
})

const toneClass = computed(() => {
  const p = pct.value
  if (p >= 90) return 'usage__fill--danger'
  if (p >= 70) return 'usage__fill--warning'
  return 'usage__fill--primary'
})

const limitLabel = computed(() => {
  if (!props.limit) return '∞'
  return props.limit.toLocaleString()
})

const title = computed(() => `${props.used ?? 0} / ${limitLabel.value} · ${pct.value}%`)
</script>

<style scoped>
.usage {
  display: flex;
  align-items: center;
  gap: 10px;
}

.usage__bar {
  flex: 1;
  min-width: 60px;
  height: 6px;
  background: var(--color-bg-hover);
  border-radius: 3px;
  overflow: hidden;
}

.usage__fill {
  height: 100%;
  border-radius: 3px;
  transition: width var(--transition);
}

.usage__fill--primary { background: linear-gradient(90deg, #3b82f6, #60a5fa); }
.usage__fill--warning { background: linear-gradient(90deg, #d97706, #f59e0b); }
.usage__fill--danger  { background: linear-gradient(90deg, #dc2626, #f87171); }

.usage__text {
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.usage__sep { margin: 0 2px; color: var(--color-text-placeholder); }
.usage__limit { color: var(--color-text-regular); }
</style>
