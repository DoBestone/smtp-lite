<template>
  <div class="logs-page">
    <div class="logs-page__toolbar">
      <p class="logs-page__sub">{{ t('logs.subtitle') }}</p>
      <div class="logs-page__actions">
        <el-button :icon="Download" @click="onExport" :loading="exporting">
          {{ t('logs.exportCsv') }}
        </el-button>
        <el-button :icon="Refresh" @click="load" :loading="loading" />
      </div>
    </div>

    <el-card shadow="never" class="logs-page__card">
      <el-table
        v-loading="loading"
        :data="logs"
        class="logs-table"
        row-key="id"
      >
        <el-table-column :label="t('logs.time')" width="150">
          <template #default="{ row }">
            <span class="mono logs-table__time">{{ formatDateTime(row.created_at) }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('logs.toEmail')" min-width="200">
          <template #default="{ row }">
            <span class="mono logs-table__email">{{ row.to_email }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('logs.subject')" min-width="240">
          <template #default="{ row }">
            <span class="logs-table__subject" :title="row.subject">{{ row.subject || '—' }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <StatusPill :kind="statusKind(row.status)">{{ statusLabel(row.status) }}</StatusPill>
          </template>
        </el-table-column>

        <el-table-column :label="t('logs.tracking')" width="160">
          <template #default="{ row }">
            <div class="logs-table__track">
              <span
                class="logs-track"
                :class="row.opened ? 'is-on' : 'is-off'"
                :title="row.opened_at ? formatDateTime(row.opened_at) : t('logs.notOpened')"
              >
                <el-icon><View /></el-icon>
                {{ row.opened ? t('logs.opened') : t('logs.notOpened') }}
              </span>
              <span
                class="logs-track"
                :class="row.clicked ? 'is-on' : 'is-off'"
                :title="row.clicked_at ? formatDateTime(row.clicked_at) : t('logs.notClicked')"
              >
                <el-icon><Pointer /></el-icon>
                {{ row.clicked ? t('logs.clicked') : t('logs.notClicked') }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('logs.error')" min-width="200">
          <template #default="{ row }">
            <span v-if="row.error_message" class="logs-table__error" :title="row.error_message">
              <el-icon><WarningFilled /></el-icon>
              {{ row.error_message }}
            </span>
            <span v-else class="logs-table__muted">—</span>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="Document" :text="'暂无发送记录'" />
        </template>
      </el-table>

      <div v-if="total > 0" class="logs-page__pager">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next"
          background
          @current-change="load"
          @size-change="onSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Document,
  Download,
  Pointer,
  Refresh,
  View,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { logsApi } from '@/api'
import type { SendLog, SendLogStatus } from '@/api/types'
import StatusPill from './components/StatusPill.vue'
import EmptyHint from './components/EmptyHint.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const logs = ref<SendLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(50)
const loading = ref(false)
const exporting = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await logsApi.list(page.value, pageSize.value)
    logs.value = res.logs ?? []
    total.value = res.total ?? 0
  } finally {
    loading.value = false
  }
}

function onSizeChange() {
  page.value = 1
  load()
}

onMounted(load)

function statusKind(
  status: SendLogStatus
): 'success' | 'danger' | 'info' | 'muted' {
  if (status === 'sent' || status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'pending') return 'info'
  return 'muted'
}

function statusLabel(status: SendLogStatus): string {
  if (status === 'sent' || status === 'success') return t('logs.statusSent')
  if (status === 'failed') return t('logs.statusFailed')
  if (status === 'pending') return t('logs.statusPending')
  return status
}

async function onExport() {
  exporting.value = true
  try {
    const blob = await logsApi.exportCsv()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `send_logs_${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success(t('common.success'))
  } finally {
    exporting.value = false
  }
}
</script>

<style scoped>
.logs-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.logs-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.logs-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }

.logs-page__actions { display: flex; gap: 8px; }

.logs-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.logs-page__card :deep(.el-card__body) { padding: 0; }

.logs-page__pager {
  padding: 12px 16px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: flex-end;
}

/* ---------- 表格 ---------- */
.logs-table { font-size: 12.5px; }

.logs-table__time { font-size: 11.5px; color: var(--color-text-secondary); }
.logs-table__email { color: var(--color-text-primary); }

.logs-table__subject {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-regular);
}

.logs-table__track {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.logs-track {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.01em;
}

.logs-track.is-on { color: var(--color-primary); }
.logs-track.is-off { color: var(--color-text-placeholder); }
.logs-track :deep(.el-icon) { font-size: 11px; }

.logs-table__error {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11.5px;
  color: var(--color-danger);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.logs-table__muted { color: var(--color-text-placeholder); }
</style>
