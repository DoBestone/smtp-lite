<template>
  <div
    class="stat-card"
    :class="[`stat-card--${tone}`]"
    :style="{ animationDelay: `${delay}ms` }"
  >
    <div class="stat-card__head">
      <span class="stat-card__title">{{ title }}</span>
      <div class="stat-card__icon">
        <el-icon :size="18"><component :is="icon" /></el-icon>
      </div>
    </div>
    <div class="stat-card__value-wrap">
      <span class="stat-card__value">{{ displayValue }}</span>
      <span v-if="suffix" class="stat-card__suffix">{{ suffix }}</span>
    </div>
    <div v-if="sub" class="stat-card__sub">{{ sub }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'

const props = withDefaults(
  defineProps<{
    title: string
    value?: number | null
    icon: Component
    suffix?: string
    sub?: string
    decimals?: number
    delay?: number
    tone?: 'primary' | 'success' | 'danger' | 'warning' | 'info' | 'accent'
  }>(),
  { tone: 'primary', delay: 0, decimals: 0 }
)

const displayValue = computed(() => {
  const v = props.value
  if (v == null || Number.isNaN(v)) return '—'
  if (props.decimals && props.decimals > 0) return Number(v).toFixed(props.decimals)
  return Math.round(Number(v)).toLocaleString()
})
</script>

<style scoped>
.stat-card {
  position: relative;
  padding: 14px 16px 16px;
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: var(--shadow-xs);
  overflow: hidden;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  opacity: 0;
  transform: translateY(6px);
  animation: rise .42s cubic-bezier(.2,.7,.2,1) forwards;
}

.stat-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-primary);
}

@keyframes rise {
  to { opacity: 1; transform: none; }
}

.stat-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}

.stat-card__title {
  font-size: 11.5px;
  font-weight: 500;
  color: var(--color-text-secondary);
  letter-spacing: 0.02em;
  line-height: 1.4;
}

.stat-card__icon {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-card__value-wrap {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.stat-card__value {
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.015em;
  line-height: 1.1;
}

.stat-card__suffix {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.stat-card__sub {
  margin-top: 6px;
  font-size: 11.5px;
  font-family: var(--font-mono);
  color: var(--color-text-secondary);
}

/* Tone —— 仅图标用语义色，不过度占屏 */
.stat-card--primary .stat-card__icon {
  background: var(--color-primary-lighter);
  color: var(--color-primary);
}
.stat-card--success .stat-card__icon {
  background: #dcfce7;
  color: #16a34a;
}
.stat-card--danger .stat-card__icon {
  background: #fee2e2;
  color: #dc2626;
}
.stat-card--warning .stat-card__icon {
  background: #fef3c7;
  color: #d97706;
}
.stat-card--info .stat-card__icon {
  background: #cffafe;
  color: #0891b2;
}
.stat-card--accent .stat-card__icon {
  background: #f3e8ff;
  color: #7c3aed;
}
</style>
