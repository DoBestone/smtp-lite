<template>
  <div class="bl-page">
    <div class="bl-page__toolbar">
      <p class="bl-page__sub">{{ t('blacklist.subtitle') }}</p>
      <div class="bl-page__actions">
        <el-button :icon="Refresh" @click="load" :loading="loading" />
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('blacklist.addEmail') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="bl-page__card">
      <el-table v-loading="loading" :data="list" class="bl-table" row-key="id">
        <el-table-column :label="t('blacklist.email')" min-width="240">
          <template #default="{ row }">
            <span class="mono bl-table__email">{{ row.email }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('blacklist.reason')" min-width="240">
          <template #default="{ row }">
            <span v-if="row.reason" class="bl-table__reason">{{ row.reason }}</span>
            <span v-else class="bl-table__muted">—</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.createdAt')" width="180">
          <template #default="{ row }">
            <span class="mono bl-table__meta">{{ formatDateTime(row.created_at) }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operations')" width="100" align="right" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" @click="onDelete(row)">
              {{ t('common.remove') }}
            </el-button>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="CircleClose" :text="'暂无黑名单记录'" />
        </template>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="t('blacklist.addEmail')"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('blacklist.email')" required>
          <el-input v-model="form.email" type="email" placeholder="user@example.com">
            <template #prefix><el-icon><Message /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item :label="t('blacklist.reason')">
          <el-input
            v-model="form.reason"
            :placeholder="t('blacklist.reasonPlaceholder')"
            maxlength="200"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          {{ t('common.add') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { CircleClose, Message, Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { blacklistApi } from '@/api'
import type { BlacklistEntry } from '@/api/types'
import EmptyHint from './components/EmptyHint.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const list = ref<BlacklistEntry[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    list.value = await blacklistApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- 弹窗 ----------
const dialogVisible = ref(false)
const submitting = ref(false)
const form = reactive({ email: '', reason: '' })

function openCreate() {
  form.email = ''
  form.reason = ''
  dialogVisible.value = true
}

async function submit() {
  if (!form.email.trim()) {
    ElMessage.error('请输入邮箱')
    return
  }
  submitting.value = true
  try {
    await blacklistApi.create({
      email: form.email.trim(),
      reason: form.reason.trim() || undefined
    })
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function onDelete(row: BlacklistEntry) {
  const confirmed = await ElMessageBox.confirm(t('blacklist.deleteConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.remove'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await blacklistApi.remove(row.id)
  ElMessage.success(t('common.success'))
  await load()
}
</script>

<style scoped>
.bl-page { display: flex; flex-direction: column; gap: 16px; }

.bl-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.bl-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }
.bl-page__actions { display: flex; gap: 8px; }

.bl-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.bl-page__card :deep(.el-card__body) { padding: 0; }

.bl-table { font-size: 13px; }
.bl-table__email { color: var(--color-text-primary); }
.bl-table__reason { color: var(--color-text-regular); }
.bl-table__muted { color: var(--color-text-placeholder); }
.bl-table__meta { font-size: 11.5px; color: var(--color-text-secondary); }
</style>
