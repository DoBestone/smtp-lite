<template>
  <span class="pill" :class="[`pill--${kind}`]">
    <span class="pill__dot"></span>
    <span class="pill__text"><slot /></span>
  </span>
</template>

<script setup lang="ts">
defineProps<{
  kind: 'success' | 'warning' | 'danger' | 'info' | 'primary' | 'muted'
}>()
</script>

<style scoped>
.pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 22px;
  padding: 0 10px;
  border-radius: 4px;
  font-size: 11.5px;
  font-weight: 500;
  letter-spacing: 0.01em;
  white-space: nowrap;
}

.pill__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.pill--success { background: #dcfce7; color: #15803d; }
.pill--success .pill__dot { background: #16a34a; }

.pill--warning { background: #fef3c7; color: #b45309; }
.pill--warning .pill__dot { background: #d97706; }

.pill--danger { background: #fee2e2; color: #b91c1c; }
.pill--danger .pill__dot { background: #dc2626; }

.pill--info { background: #cffafe; color: #0e7490; }
.pill--info .pill__dot { background: #0891b2; }

.pill--primary { background: var(--color-primary-lighter); color: var(--color-primary-dark); }
.pill--primary .pill__dot { background: var(--color-primary); }

.pill--muted { background: var(--color-bg-hover); color: var(--color-text-secondary); }
.pill--muted .pill__dot { background: var(--color-text-placeholder); }
</style>
