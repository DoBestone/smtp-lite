<template>
  <div class="cb">
    <div class="cb__bar">
      <div class="cb__lang" v-if="lang">{{ lang }}</div>
      <button class="cb__copy" @click="copy">
        <el-icon :size="13"><DocumentCopy v-if="!copied" /><CircleCheck v-else /></el-icon>
        <span>{{ copied ? '已复制' : '复制' }}</span>
      </button>
    </div>
    <pre class="cb__pre" v-text="code"></pre>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CircleCheck, DocumentCopy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{ code: string; lang?: string }>()
const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
.cb {
  position: relative;
  background: #0f172a;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #1e293b;
}

.cb__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px 6px 14px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}

.cb__lang {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: #94a3b8;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.cb__copy {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid #334155;
  border-radius: 5px;
  color: #cbd5e1;
  font-size: 11px;
  cursor: pointer;
  transition: background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
}

.cb__copy:hover {
  background: rgba(59, 130, 246, 0.12);
  border-color: #3b82f6;
  color: #7dd3fc;
}

.cb__pre {
  margin: 0;
  padding: 14px 16px;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.65;
  color: #e2e8f0;
  overflow: auto;
  white-space: pre;
  max-height: 520px;
}

.cb__pre::-webkit-scrollbar-thumb { background: #334155; }
</style>
