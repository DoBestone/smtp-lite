<template>
  <div class="wh-page">
    <div class="wh-page__toolbar">
      <p class="wh-page__sub">{{ t('webhooks.subtitle') }}</p>
      <div class="wh-page__actions">
        <el-button :icon="Refresh" @click="load" :loading="loading" />
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('webhooks.addWebhook') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="wh-page__card">
      <el-table v-loading="loading" :data="webhooks" class="wh-table" row-key="id">
        <el-table-column :label="t('webhooks.name')" min-width="180">
          <template #default="{ row }">
            <span class="wh-table__name">{{ row.name }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('webhooks.url')" min-width="260">
          <template #default="{ row }">
            <code class="wh-table__url" :title="row.url">{{ row.url }}</code>
          </template>
        </el-table-column>

        <el-table-column :label="t('webhooks.events')" width="220">
          <template #default="{ row }">
            <div class="wh-table__events">
              <span v-for="e in parseEvents(row.events)" :key="e" class="wh-evt">
                {{ eventLabel(e) }}
              </span>
              <span v-if="!parseEvents(row.events).length" class="wh-table__muted">—</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.status')" width="96">
          <template #default="{ row }">
            <StatusPill :kind="row.enabled ? 'success' : 'muted'">
              {{ row.enabled ? t('common.enable') : t('common.disable') }}
            </StatusPill>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operations')" width="220" align="right" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :loading="testingId === row.id" @click="onTest(row)">
              {{ t('common.test') }}
            </el-button>
            <el-button link @click="onToggle(row)">
              {{ row.enabled ? t('common.disable') : t('common.enable') }}
            </el-button>
            <el-button link type="danger" @click="onDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="Link" :text="'暂无 Webhook，点击右上角添加'" />
        </template>
      </el-table>
    </el-card>

    <!-- 弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="t('webhooks.addWebhook')"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top" class="wh-form">
        <el-form-item :label="t('webhooks.name')" required>
          <el-input v-model="form.name" placeholder="如：CRM 通知 / Slack 提醒" />
        </el-form-item>
        <el-form-item :label="t('webhooks.url')" required>
          <el-input v-model="form.url" placeholder="https://your-endpoint.example.com/hook">
            <template #prefix><el-icon><Link /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item :label="t('webhooks.secret')">
          <el-input
            v-model="form.secret"
            placeholder="可选 · 用于签名校验"
            show-password
          />
        </el-form-item>
        <el-form-item :label="t('webhooks.events')">
          <el-checkbox-group v-model="form.events" class="wh-form__events">
            <el-checkbox
              v-for="opt in eventOptions"
              :key="opt.key"
              :value="opt.key"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Link, Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { webhooksApi } from '@/api'
import type { Webhook } from '@/api/types'
import StatusPill from './components/StatusPill.vue'
import EmptyHint from './components/EmptyHint.vue'

const { t } = useI18n()

const webhooks = ref<Webhook[]>([])
const loading = ref(false)
const testingId = ref<string | null>(null)

const eventOptions = computed(() => [
  { key: 'send_success', label: t('webhooks.evtSendSuccess') },
  { key: 'send_failed', label: t('webhooks.evtSendFailed') },
  { key: 'opened', label: t('webhooks.evtOpened') },
  { key: 'clicked', label: t('webhooks.evtClicked') }
])

function eventLabel(key: string): string {
  return eventOptions.value.find((e) => e.key === key)?.label ?? key
}

function parseEvents(events: unknown): string[] {
  if (!events) return []
  if (Array.isArray(events)) return events as string[]
  try {
    const parsed = JSON.parse(String(events))
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

async function load() {
  loading.value = true
  try {
    webhooks.value = await webhooksApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- Dialog ----------
const dialogVisible = ref(false)
const submitting = ref(false)
const form = reactive({
  name: '',
  url: '',
  secret: '',
  events: [] as string[]
})

function openCreate() {
  form.name = ''
  form.url = ''
  form.secret = ''
  form.events = []
  dialogVisible.value = true
}

async function submit() {
  if (!form.name.trim() || !form.url.trim()) {
    ElMessage.error('请填写名称与 URL')
    return
  }
  submitting.value = true
  try {
    await webhooksApi.create({
      name: form.name,
      url: form.url,
      secret: form.secret || undefined,
      events: JSON.stringify(form.events)
    })
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function onTest(row: Webhook) {
  testingId.value = row.id
  try {
    await webhooksApi.test(row.id)
    ElMessage.success(t('webhooks.testSuccess'))
  } finally {
    testingId.value = null
  }
}

async function onToggle(row: Webhook) {
  await webhooksApi.toggle(row.id)
  await load()
}

async function onDelete(row: Webhook) {
  const confirmed = await ElMessageBox.confirm(t('webhooks.deleteConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await webhooksApi.remove(row.id)
  ElMessage.success(t('common.success'))
  await load()
}
</script>

<style scoped>
.wh-page { display: flex; flex-direction: column; gap: 16px; }

.wh-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.wh-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }
.wh-page__actions { display: flex; gap: 8px; }

.wh-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.wh-page__card :deep(.el-card__body) { padding: 0; }

.wh-table { font-size: 13px; }
.wh-table__name { font-weight: 500; color: var(--color-text-primary); }

.wh-table__url {
  display: inline-block;
  max-width: 100%;
  padding: 2px 8px;
  background: var(--color-bg-hover);
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--color-text-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.wh-table__events {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.wh-evt {
  display: inline-block;
  padding: 1px 7px;
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
  border-radius: 4px;
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.01em;
}

.wh-table__muted { color: var(--color-text-placeholder); font-size: 11.5px; }

/* Form */
.wh-form :deep(.el-form-item) { margin-bottom: 14px; }
.wh-form :deep(.el-form-item__label) {
  padding-bottom: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.wh-form__events {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.wh-form__events :deep(.el-checkbox) { margin-right: 0; height: 26px; }
.wh-form__events :deep(.el-checkbox__label) { font-size: 12.5px; }
</style>
