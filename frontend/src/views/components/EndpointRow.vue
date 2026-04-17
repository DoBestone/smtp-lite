<template>
  <div class="ep">
    <span class="ep__method" :class="`ep__method--${methodLower}`">{{ method }}</span>
    <code class="ep__path">{{ path }}</code>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ method: string; path: string }>()
const methodLower = computed(() => props.method.toLowerCase())
</script>

<style scoped>
.ep {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px 6px 6px;
  background: var(--color-bg-subtle);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 12px;
}

.ep__method {
  padding: 2px 8px;
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 0.04em;
}

.ep__method--get    { background: #16a34a; }
.ep__method--post   { background: #3b82f6; }
.ep__method--put    { background: #d97706; }
.ep__method--delete { background: #dc2626; }
.ep__method--patch  { background: #7c3aed; }

.ep__path {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--color-text-primary);
  word-break: break-all;
}
</style>
