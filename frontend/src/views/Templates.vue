<template>
  <div class="tpl-page">
    <div class="tpl-page__toolbar">
      <p class="tpl-page__sub">{{ t('templates.subtitle') }}</p>
      <div class="tpl-page__actions">
        <el-button :icon="Refresh" @click="load" :loading="loading" />
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ t('templates.addTemplate') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="tpl-page__card">
      <el-table v-loading="loading" :data="templates" class="tpl-table" row-key="id">
        <el-table-column :label="t('templates.name')" min-width="200">
          <template #default="{ row }">
            <div class="tpl-table__main">
              <span class="tpl-table__name">{{ row.name }}</span>
              <span v-if="row.description" class="tpl-table__desc">{{ row.description }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('templates.subject')" min-width="240">
          <template #default="{ row }">
            <span class="tpl-table__subject">{{ row.subject || '—' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="格式" width="84">
          <template #default="{ row }">
            <span class="type-tag" :class="{ 'type-tag--html': row.is_html }">
              {{ row.is_html ? 'HTML' : 'TEXT' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.updatedAt')" width="170">
          <template #default="{ row }">
            <span class="mono tpl-table__meta">{{ formatDateTime(row.updated_at) }}</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operations')" width="180" align="right" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" @click="onDelete(row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>

        <template #empty>
          <EmptyHint :icon="Files" :text="'暂无邮件模板'" />
        </template>
      </el-table>
    </el-card>

    <!-- 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editing ? '编辑模板' : t('templates.addTemplate')"
      :width="840"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="tpl-form"
      >
        <div class="tpl-form__row">
          <el-form-item :label="t('templates.name')" prop="name" class="tpl-form__name">
            <el-input v-model="form.name" placeholder="如：欢迎邮件 / 密码重置" maxlength="100" />
          </el-form-item>
          <el-form-item label="类型" class="tpl-form__type">
            <el-radio-group v-model="form.is_html">
              <el-radio :value="false">纯文本</el-radio>
              <el-radio :value="true">HTML</el-radio>
            </el-radio-group>
          </el-form-item>
        </div>

        <el-form-item :label="t('templates.subject')" prop="subject">
          <el-input v-model="form.subject" placeholder="邮件主题" maxlength="200" />
        </el-form-item>

        <el-form-item :label="t('templates.description')">
          <el-input v-model="form.description" placeholder="内部备注，不会显示给收件人" maxlength="200" />
        </el-form-item>

        <el-form-item :label="t('templates.body')" prop="body">
          <div class="tpl-form__editor">
            <el-input
              v-model="form.body"
              type="textarea"
              :rows="12"
              resize="none"
              class="tpl-form__editor-input"
              placeholder="支持变量占位符，如 {{ .Name }}"
            />
            <div class="tpl-form__preview">
              <div class="tpl-form__preview-head">{{ t('templates.preview') }}</div>
              <div class="tpl-form__preview-body">
                <div v-if="form.is_html && form.body" class="tpl-form__preview-html" v-html="form.body"></div>
                <pre v-else-if="!form.is_html && form.body" class="tpl-form__preview-text">{{ form.body }}</pre>
                <div v-else class="tpl-form__preview-empty">{{ t('templates.previewEmpty') }}</div>
              </div>
            </div>
          </div>
        </el-form-item>
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
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Files, Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { templatesApi } from '@/api'
import type { Template, TemplatePayload } from '@/api/types'
import EmptyHint from './components/EmptyHint.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const templates = ref<Template[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    templates.value = await templatesApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---------- 表单 ----------
const dialogVisible = ref(false)
const editing = ref<Template | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const initialForm = (): TemplatePayload => ({
  name: '',
  subject: '',
  body: '',
  is_html: false,
  description: ''
})

const form = reactive<TemplatePayload>(initialForm())

const rules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  subject: [{ required: true, message: '请输入主题', trigger: 'blur' }],
  body: [{ required: true, message: '请输入正文', trigger: 'blur' }]
}

function openCreate() {
  editing.value = null
  Object.assign(form, initialForm())
  dialogVisible.value = true
}

function openEdit(row: Template) {
  editing.value = row
  Object.assign(form, {
    name: row.name,
    subject: row.subject,
    body: row.body,
    is_html: row.is_html,
    description: row.description ?? ''
  })
  dialogVisible.value = true
}

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editing.value) {
      await templatesApi.update(editing.value.id, { ...form })
    } else {
      await templatesApi.create({ ...form })
    }
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function onDelete(row: Template) {
  const confirmed = await ElMessageBox.confirm(t('templates.deleteConfirm'), {
    type: 'warning',
    confirmButtonText: t('common.delete'),
    cancelButtonText: t('common.cancel')
  }).catch(() => false)
  if (!confirmed) return
  await templatesApi.remove(row.id)
  ElMessage.success(t('common.success'))
  await load()
}
</script>

<style scoped>
.tpl-page { display: flex; flex-direction: column; gap: 16px; }

.tpl-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
}

.tpl-page__sub { font-size: 12.5px; color: var(--color-text-secondary); }
.tpl-page__actions { display: flex; gap: 8px; }

.tpl-page__card {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}
.tpl-page__card :deep(.el-card__body) { padding: 0; }

.tpl-table { font-size: 13px; }
.tpl-table__main { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.tpl-table__name { font-weight: 500; color: var(--color-text-primary); }
.tpl-table__desc {
  font-size: 11.5px;
  color: var(--color-text-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tpl-table__subject {
  color: var(--color-text-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  max-width: 100%;
}
.tpl-table__meta { font-size: 11.5px; color: var(--color-text-secondary); }

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--color-bg-hover);
  color: var(--color-text-secondary);
  font-family: var(--font-mono);
  font-size: 10.5px;
  letter-spacing: 0.04em;
}
.type-tag--html { background: #fef3c7; color: #b45309; }

/* 表单 */
.tpl-form :deep(.el-form-item) { margin-bottom: 12px; }
.tpl-form :deep(.el-form-item__label) {
  padding-bottom: 5px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.tpl-form__row { display: flex; gap: 16px; }
.tpl-form__name { flex: 1; }
.tpl-form__type { min-width: 180px; }

.tpl-form__editor {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;
}

.tpl-form__editor-input :deep(.el-textarea__inner) {
  font-family: var(--font-mono);
  font-size: 12.5px;
  line-height: 1.6;
}

.tpl-form__preview {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

.tpl-form__preview-head {
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-secondary);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  background: var(--color-bg-subtle);
  border-bottom: 1px solid var(--color-border);
}

.tpl-form__preview-body {
  flex: 1;
  padding: 14px 16px;
  min-height: 280px;
  max-height: 320px;
  overflow: auto;
}

.tpl-form__preview-html {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-primary);
}

.tpl-form__preview-text {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-regular);
  white-space: pre-wrap;
  word-break: break-word;
}

.tpl-form__preview-empty {
  font-size: 12px;
  color: var(--color-text-placeholder);
  font-style: italic;
}

@media (max-width: 768px) {
  .tpl-form__row { flex-direction: column; gap: 0; }
  .tpl-form__editor { grid-template-columns: 1fr; }
  .tpl-form__preview-body { min-height: 160px; }
}
</style>
