<template>
  <div class="uc">
    <div class="uc__card">
      <div class="uc__icon">
        <el-icon :size="22"><Tools /></el-icon>
      </div>
      <h2 class="uc__title">{{ label }}</h2>
      <p class="uc__desc">
        该页面将在 <strong>Phase {{ phase }}</strong> 完成重构 · 旧版仍可通过直接调 API 使用
      </p>
      <div class="uc__meta">
        <span class="uc__tag">{{ route.path }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Tools } from '@element-plus/icons-vue'

const props = defineProps<{ phase: number }>()
const route = useRoute()
const { t } = useI18n()

const label = computed(() => {
  const key = route.meta.title as string | undefined
  return key ? t(key) : '页面'
})
</script>

<style scoped>
.uc {
  display: flex;
  justify-content: center;
  padding: 32px 16px;
}

.uc__card {
  max-width: 480px;
  width: 100%;
  padding: 32px 28px;
  background: #fff;
  border: 1px dashed var(--color-border);
  border-radius: 14px;
  text-align: center;
  box-shadow: var(--shadow-xs);
}

.uc__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  margin: 0 auto 14px;
  border-radius: 12px;
  background: var(--color-primary-lighter);
  color: var(--color-primary);
}

.uc__title {
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 6px;
}

.uc__desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin-bottom: 14px;
}

.uc__meta { margin-top: 14px; }

.uc__tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 4px;
  background: var(--color-bg-hover);
  color: var(--color-text-secondary);
  font-size: 11px;
  font-family: var(--font-mono);
}
</style>
