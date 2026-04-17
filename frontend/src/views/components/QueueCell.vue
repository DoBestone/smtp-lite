<template>
  <div class="q-cell" :class="[`q-cell--${tone}`]">
    <div class="q-cell__label">
      <span class="q-cell__dot" :class="{ 'is-animate': animate }"></span>
      {{ label }}
    </div>
    <div class="q-cell__value">{{ displayValue }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    value?: number | null
    tone?: 'primary' | 'success' | 'danger' | 'info'
    animate?: boolean
  }>(),
  { tone: 'info', animate: false }
)

const displayValue = computed(() => {
  const v = props.value
  if (v == null) return '0'
  return Number(v).toLocaleString()
})
</script>

<style scoped>
.q-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-subtle);
  transition: border-color var(--transition-fast);
}

.q-cell:hover { border-color: var(--color-primary); }

.q-cell__label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 11.5px;
  font-weight: 500;
  color: var(--color-text-secondary);
  letter-spacing: 0.02em;
}

.q-cell__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-text-placeholder);
}

.q-cell--info .q-cell__dot { background: #0ea5e9; }
.q-cell--primary .q-cell__dot { background: var(--color-primary); }
.q-cell--success .q-cell__dot { background: var(--color-success); }
.q-cell--danger .q-cell__dot { background: var(--color-danger); }

.q-cell__dot.is-animate {
  box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.45);
  animation: q-pulse 1.6s ease-out infinite;
}

@keyframes q-pulse {
  0%   { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.45); }
  70%  { box-shadow: 0 0 0 7px rgba(59, 130, 246, 0); }
  100% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0); }
}

.q-cell__value {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
  line-height: 1.1;
}
</style>
