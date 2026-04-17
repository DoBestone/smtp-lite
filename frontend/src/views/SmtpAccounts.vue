<template>
  <div class="smtp-page">
    <!-- 顶部工具条 · 紧凑 · 裸放 -->
    <div class="smtp-page__toolbar">
      <p class="smtp-page__sub">{{ t('smtp.subtitle') }}</p>
      <div class="smtp-page__actions">
        <el-button :icon="Download" @click="exportCsv">{{ t('smtp.export') }}</el-button>
        <el-button :icon="Refresh" @click="load" :loading="loading" />
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('smtp.addAccount') }}
        </el-button>
      </div>
    </div>

    <!-- 数据表 -->
    <el-card shadow="never" class="smtp-page__card">
      <el-table
        v-loading="loading"
        :data="accounts"
        class="smtp-table"
        empty-text=""
        row-key="id"
      >
        <el-table-column :label="t('smtp.email')" min-width="200" prop="email">
          <template #default="{ row }">
            <div class="smtp-table__email">
              <div class="smtp-table__email-main">{{ row.email }}</div>
              <div v-if="row.last_error" class="smtp-table__email-err" :title="row.last_error">
                <el-icon><WarningFilled /></el-icon>
                {{ row.last_error }}
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('smtp.host')" min-width="180">
          <template #default="{ row }">
            <span class="mono">{{ row.smtp_host }}<span class="smtp-table__port">:{{ row.smtp_port }}</span></span>
          </template>
        </el-table-column>

        <el-table-column :label="t('smtp.dailyUsed')" width="180" align="left">
          <template #default="{ row }">
            <UsageBar :used="row.daily_used" :limit="row.daily_limit" />
          </template>
        </el-table-column>

        <el-table-column :label="t('smtp.priority')" width="80" align="center">
          <template #default="{ row }">
            <span class="mono num-cell">{{ row.priority ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.status')" width="96">
          <template #default="{ row }">
            <StatusPill :kind="row.status === 'active' ? 'success' : 'muted'">
              {{ row.status === 'active' ? t('smtp.statusActive') : t('smtp.statusDisabled') }}
            </StatusPill>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operations')" width="220" align="right" fixed="right">
          <template #default="{ row }">
            <div class="smtp-table__ops">
              <el-button
                link
                type="primary"
                :loading="testingId === row.id"
                @click="onTest(row)"
              >
                {{ t('smtp.testConn') }}
              </el-button>
              <el-button link @click="onToggle(row)">
                {{ row.status === 'active' ? t('common.disable') : t('common.enable') }}
              </el-button>
              <el-button link type="danger" @click="onDelete(row)">
                {{ t('common.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="Message" :text="'暂无 SMTP 账号，点击右上角添加'" />
        </template>
      </el-table>
    </el-card>

    <!-- 添加 / 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editing ? '编辑 SMTP 账号' : t('smtp.addAccount')"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="smtp-form"
      >
        <el-form-item :label="t('smtp.email')" prop="email">
          <el-input v-model="form.email" placeholder="your@domain.com">
            <template #prefix><el-icon><Message /></el-icon></template>
          </el-input>
        </el-form-item>

        <el-form-item :label="t('smtp.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editing ? '留空则不修改' : '邮箱密码或应用专用密码'"
          >
            <template #prefix><el-icon><Lock /></el-icon></template>
          </el-input>
        </el-form-item>

        <div class="smtp-form__row">
          <el-form-item :label="t('smtp.host')" prop="smtp_host" class="smtp-form__host">
            <el-input v-model="form.smtp_host" placeholder="smtp.gmail.com" />
          </el-form-item>
          <el-form-item :label="t('smtp.port')" prop="smtp_port" class="smtp-form__port">
            <el-input-number
              v-model="form.smtp_port"
              :min="1"
              :max="65535"
              controls-position="right"
              style="width: 100%"
            />
          </el-form-item>
        </div>

        <div class="smtp-form__row">
          <el-form-item :label="t('smtp.dailyLimit')" class="smtp-form__half">
            <el-input-number
              v-model="form.daily_limit"
              :min="0"
              :max="100000"
              controls-position="right"
              style="width: 100%"
              placeholder="0 表示不限制"
            />
          </el-form-item>
          <el-form-item :label="t('smtp.priority')" class="smtp-form__half">
            <el-input-number
              v-model="form.priority"
              :min="0"
              :max="100"
              controls-position="right"
              style="width: 100%"
            />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Download,
  Lock,
  Message,
  Plus,
  Refresh,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { smtpApi, systemApi } from '@/api'
import type { SmtpAccount, SmtpAccountPayload } from '@/api/types'
import StatusPill from './components/StatusPill.vue'
import UsageBar from './components/UsageBar.vue'
import EmptyHint from './components/EmptyHint.vue'

const { t } = useI18n()

const accounts = ref<SmtpAccount[]>([])
const loading = ref(false)
const testingId = ref<string | null>(null)

async function load() {
  loading.value = true
  try {
    accounts.value = await smtpApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- 弹窗 / 表单 ----------
const dialogVisible = ref(false)
const editing = ref<SmtpAccount | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const initialForm = (): SmtpAccountPayload & { priority: number } => ({
  email: '',
  password: '',
  smtp_host: '',
  smtp_port: 587,
  daily_limit: 500,
  priority: 0
})

const form = reactive(initialForm())

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { validator: (_, value, cb) => (!editing.value && !value ? cb(new Error('请输入密码')) : cb()), trigger: 'blur' }
  ],
  smtp_host: [{ required: true, message: '请输入 SMTP 服务器', trigger: 'blur' }]
}

function openCreate() {
  editing.value = null
  Object.assign(form, initialForm())
  dialogVisible.value = true
}

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editing.value) {
      const payload: Partial<SmtpAccountPayload> = { ...form }
      if (!payload.password) delete payload.password
      await smtpApi.update(editing.value.id, payload)
      ElMessage.success(t('common.success'))
    } else {
      await smtpApi.create({ ...form })
      ElMessage.success(t('common.success'))
    }
    dialogVisible.value = false
    await load()
  } catch {
    // 已由 interceptor 处理
  } finally {
    submitting.value = false
  }
}

// ---------- 操作 ----------
async function onTest(row: SmtpAccount) {
  testingId.value = row.id
  try {
    const r = await smtpApi.test(row.id)
    if (r.success) {
      ElMessage.success(t('smtp.testSuccess'))
    } else {
      ElMessage.error(`${t('smtp.testFailed')}${r.error ? ` · ${r.error}` : ''}`)
    }
  } finally {
    testingId.value = null
  }
}

async function onToggle(row: SmtpAccount) {
  try {
    await smtpApi.toggle(row.id)
    await load()
  } catch {
    /* ignore */
  }
}

async function onDelete(row: SmtpAccount) {
  const confirm = await ElMessageBox.confirm(t('smtp.confirmDelete'), {
    type: 'warning',
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirm) return

  await smtpApi.remove(row.id)
  ElMessage.success(t('smtp.deleted'))
  await load()
}

async function exportCsv() {
  const token = localStorage.getItem('token') ?? ''
  const resp = await fetch(systemApi.exportAccountsUrl, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!resp.ok) {
    ElMessage.error(t('common.failed'))
    return
  }
  const blob = await resp.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `smtp_accounts_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success(t('common.success'))
}
</script>

<style scoped>
.smtp-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.smtp-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.smtp-page__sub {
  font-size: 12.5px;
  color: var(--color-text-secondary);
}

.smtp-page__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.smtp-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}

.smtp-page__card :deep(.el-card__body) { padding: 0; }

.smtp-table { font-size: 13px; }

.smtp-table__email {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.4;
}

.smtp-table__email-main {
  font-weight: 500;
  color: var(--color-text-primary);
}

.smtp-table__email-err {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--color-danger);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.smtp-table__port { color: var(--color-text-placeholder); }

.smtp-table__ops {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.smtp-table__ops :deep(.el-button--text),
.smtp-table__ops :deep(.el-button.is-link) { padding: 4px 8px; font-size: 12.5px; }

.num-cell { font-variant-numeric: tabular-nums; }

/* 表单 */
.smtp-form :deep(.el-form-item__label) {
  padding-bottom: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
}
.smtp-form :deep(.el-form-item) { margin-bottom: 14px; }

.smtp-form__row {
  display: flex;
  gap: 12px;
}
.smtp-form__host { flex: 1; }
.smtp-form__port { width: 140px; }
.smtp-form__half { flex: 1; }

@media (max-width: 576px) {
  .smtp-page__toolbar { flex-direction: column; align-items: flex-start; }
  .smtp-page__actions { width: 100%; }
  .smtp-form__row { flex-direction: column; gap: 0; }
  .smtp-form__port { width: 100%; }
}
</style>
