<template>
  <div class="keys-page">
    <div class="keys-page__toolbar">
      <p class="keys-page__sub">{{ t('keys.subtitle') }}</p>
      <div class="keys-page__actions">
        <el-button :icon="Refresh" @click="load" :loading="loading" />
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('keys.addKey') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="keys-page__card">
      <el-table v-loading="loading" :data="keys" class="keys-table" row-key="id">
        <el-table-column :label="t('keys.name')" prop="name" min-width="200">
          <template #default="{ row }">
            <span class="keys-table__name">{{ row.name || '未命名' }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('keys.keyPrefix')" width="180">
          <template #default="{ row }">
            <code class="keys-table__prefix">{{ row.key_prefix }}····</code>
          </template>
        </el-table-column>

        <el-table-column :label="t('keys.lastUsed')" width="200">
          <template #default="{ row }">
            <span v-if="row.last_used_at" class="mono keys-table__meta">
              {{ formatDateTime(row.last_used_at) }}
            </span>
            <span v-else class="keys-table__meta keys-table__meta--muted">
              {{ t('keys.neverUsed') }}
            </span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.createdAt')" width="180">
          <template #default="{ row }">
            <span class="mono keys-table__meta">{{ formatDateTime(row.created_at) }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operations')" width="180" align="right" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="onReset(row)">重置</el-button>
            <el-button link type="danger" @click="onDelete(row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="Key" :text="'暂无 API Key，点击右上角生成第一个'" />
        </template>
      </el-table>
    </el-card>

    <!-- 新建弹窗 -->
    <el-dialog
      v-model="createVisible"
      :title="t('keys.addKey')"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top" @submit.prevent="submitCreate">
        <el-form-item :label="t('keys.name')">
          <el-input
            v-model="newKeyName"
            :placeholder="t('keys.placeholder')"
            maxlength="100"
            show-word-limit
            @keyup.enter="submitCreate"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCreate">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Reveal 弹窗：一次性显示完整 Key -->
    <el-dialog
      v-model="revealVisible"
      :title="t('keys.revealTitle')"
      width="520px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="reveal">
        <div class="reveal__warn">
          <el-icon><WarningFilled /></el-icon>
          <span>{{ t('keys.revealHint') }}</span>
        </div>
        <div class="reveal__box">
          <code class="reveal__value">{{ revealedKey }}</code>
          <el-button
            size="small"
            type="primary"
            plain
            :icon="DocumentCopy"
            @click="copyKey"
          >
            {{ t('keys.copy') }}
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="revealVisible = false">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  DocumentCopy,
  Key,
  Plus,
  Refresh,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiKeysApi } from '@/api'
import type { ApiKey } from '@/api/types'
import EmptyHint from './components/EmptyHint.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const keys = ref<ApiKey[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    keys.value = await apiKeysApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- Create ----------
const createVisible = ref(false)
const newKeyName = ref('')
const submitting = ref(false)

function openCreate() {
  newKeyName.value = ''
  createVisible.value = true
}

async function submitCreate() {
  if (!newKeyName.value.trim()) {
    ElMessage.error('请填写 Key 名称')
    return
  }
  submitting.value = true
  try {
    const created = await apiKeysApi.create(newKeyName.value.trim())
    createVisible.value = false
    if (created.key) reveal(created.key)
    await load()
  } finally {
    submitting.value = false
  }
}

// ---------- Reveal ----------
const revealVisible = ref(false)
const revealedKey = ref('')

function reveal(k: string) {
  revealedKey.value = k
  revealVisible.value = true
}

async function copyKey() {
  try {
    await navigator.clipboard.writeText(revealedKey.value)
    ElMessage.success(t('keys.copied'))
  } catch {
    ElMessage.error('复制失败')
  }
}

// ---------- Reset / Delete ----------
async function onReset(row: ApiKey) {
  const confirmed = await ElMessageBox.confirm(t('keys.resetConfirm'), {
    type: 'warning',
    confirmButtonText: '重置',
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  const updated = await apiKeysApi.reset(row.id)
  if (updated.key) reveal(updated.key)
  await load()
}

async function onDelete(row: ApiKey) {
  const confirmed = await ElMessageBox.confirm(t('keys.deleteConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await apiKeysApi.remove(row.id)
  ElMessage.success(t('common.success'))
  await load()
}
</script>

<style scoped>
.keys-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.keys-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.keys-page__sub {
  font-size: 12.5px;
  color: var(--color-text-secondary);
}

.keys-page__actions {
  display: flex;
  gap: 8px;
}

.keys-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.keys-page__card :deep(.el-card__body) { padding: 0; }

.keys-table { font-size: 13px; }
.keys-table__name { font-weight: 500; color: var(--color-text-primary); }

.keys-table__prefix {
  display: inline-block;
  padding: 2px 10px;
  background: var(--color-bg-hover);
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--color-text-regular);
  letter-spacing: 0.04em;
}

.keys-table__meta { font-size: 11.5px; color: var(--color-text-secondary); }
.keys-table__meta--muted { color: var(--color-text-placeholder); font-family: var(--font-body); }

/* Reveal */
.reveal {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.reveal__warn {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  background: #fef3c7;
  color: #b45309;
  border: 1px solid #fde68a;
  border-radius: 8px;
  font-size: 12.5px;
}

.reveal__warn :deep(.el-icon) { margin-top: 2px; flex-shrink: 0; }

.reveal__box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #0f172a;
  border-radius: 10px;
}

.reveal__value {
  flex: 1;
  min-width: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  color: #7dd3fc;
  overflow: auto;
  white-space: nowrap;
  scrollbar-width: thin;
}
</style>
