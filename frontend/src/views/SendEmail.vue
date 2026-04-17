<template>
  <div class="send-page">
    <!-- 顶部工具条 -->
    <div class="send-page__toolbar">
      <p class="send-page__sub">{{ t('send.subtitle') }}</p>
      <el-segmented
        v-model="mode"
        :options="modeOptions"
        size="default"
        class="send-page__mode"
      />
    </div>

    <div class="send-page__grid">
      <!-- 主表单 -->
      <el-card shadow="never" class="send-page__main">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="send-form"
          @submit.prevent="submit"
        >
          <!-- 单发 · 收件人 -->
          <el-form-item v-if="mode !== 'batch'" :label="t('send.to')" prop="to">
            <el-input
              v-model="form.to"
              type="email"
              :placeholder="t('send.toPlaceholder')"
            >
              <template #prefix><el-icon><User /></el-icon></template>
            </el-input>
          </el-form-item>

          <!-- 批量 · 收件人列表 -->
          <el-form-item v-if="mode === 'batch'" :label="t('send.batchList')" prop="batchEmails">
            <el-input
              v-model="batchEmails"
              type="textarea"
              :rows="5"
              :placeholder="t('send.batchPlaceholder')"
            />
          </el-form-item>

          <!-- 分组导入 -->
          <div class="send-form__group-row">
            <div class="send-form__field">
              <label class="send-form__label">{{ t('send.importFromGroup') }}</label>
              <el-select
                v-model="groupId"
                placeholder="选择收件人分组"
                clearable
                filterable
                @change="onGroupChange"
                style="width: 100%"
              >
                <el-option
                  v-for="g in groups"
                  :key="g.id"
                  :label="`${g.name} (${g.count ?? 0})`"
                  :value="g.id"
                />
              </el-select>
            </div>
            <div v-if="mode !== 'batch'" class="send-form__field">
              <label class="send-form__label">分组成员</label>
              <el-select
                v-model="pickedEmail"
                placeholder="选择收件人"
                clearable
                filterable
                :disabled="!recipientOptions.length"
                @change="onPickRecipient"
                style="width: 100%"
              >
                <el-option
                  v-for="r in recipientOptions"
                  :key="r.id"
                  :label="r.name ? `${r.email} (${r.name})` : r.email"
                  :value="r.email"
                />
              </el-select>
            </div>
            <div v-else class="send-form__field send-form__field--action">
              <el-button
                type="primary"
                plain
                :disabled="!groupId || !recipientOptions.length"
                @click="importGroup"
              >
                导入 {{ recipientOptions.length || 0 }} 个成员
              </el-button>
            </div>
          </div>

          <!-- 定时发送 -->
          <el-form-item
            v-if="mode === 'scheduled'"
            :label="t('send.scheduleAt')"
            prop="scheduledAt"
          >
            <el-date-picker
              v-model="scheduledAt"
              type="datetime"
              :placeholder="t('send.scheduleAt')"
              format="YYYY-MM-DD HH:mm"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              style="width: 100%"
            />
          </el-form-item>

          <el-form-item :label="t('send.subject')" prop="subject">
            <el-input v-model="form.subject" :placeholder="t('send.subject')" />
          </el-form-item>

          <el-form-item :label="t('send.body')" prop="body">
            <el-input
              v-model="form.body"
              type="textarea"
              :rows="8"
              resize="vertical"
            />
            <div class="send-form__body-opts">
              <el-checkbox v-model="form.is_html">{{ t('send.htmlMode') }}</el-checkbox>
              <el-checkbox
                v-if="form.is_html"
                v-model="form.track_enabled"
              >
                {{ t('send.trackEnabled') }}
              </el-checkbox>
            </div>
          </el-form-item>

          <!-- 展开额外字段 -->
          <div class="send-form__more">
            <button
              type="button"
              class="send-form__more-btn"
              @click="showMore = !showMore"
            >
              <el-icon :class="{ 'is-rotated': showMore }"><ArrowRight /></el-icon>
              {{ showMore ? '收起额外字段' : '发件人 · 抄送 · 附件' }}
            </button>
          </div>

          <div v-show="showMore" class="send-form__advanced">
            <div class="send-form__row">
              <el-form-item :label="t('send.fromName')" class="send-form__half">
                <el-input v-model="form.from_name" placeholder="可选" />
              </el-form-item>
              <el-form-item
                v-if="mode === 'single' || mode === 'scheduled'"
                :label="t('send.cc')"
                class="send-form__half"
              >
                <el-input v-model="form.cc" placeholder="多人用逗号分隔" />
              </el-form-item>
            </div>

            <el-form-item
              v-if="mode === 'single' || mode === 'scheduled'"
              :label="t('send.bcc')"
            >
              <el-input v-model="form.bcc" placeholder="多人用逗号分隔" />
            </el-form-item>

            <el-form-item :label="t('send.attachments')">
              <AttachmentPicker v-model="attachments" />
            </el-form-item>
          </div>

          <!-- 结果提示 -->
          <transition name="fade">
            <div
              v-if="result"
              class="send-form__alert"
              :class="result.success ? 'is-success' : 'is-error'"
            >
              <el-icon>
                <CircleCheckFilled v-if="result.success" />
                <CircleCloseFilled v-else />
              </el-icon>
              <div class="send-form__alert-body">
                <div>{{ result.message }}</div>
                <ul v-if="result.details && result.details.length">
                  <li v-for="(d, i) in result.details" :key="i">{{ d }}</li>
                </ul>
              </div>
            </div>
          </transition>

          <!-- 操作 -->
          <div class="send-form__actions">
            <el-button @click="reset()">{{ t('common.reset') }}</el-button>
            <el-button
              type="primary"
              :loading="loading"
              native-type="submit"
              @click="submit"
            >
              <el-icon v-if="!loading"><Promotion /></el-icon>
              {{ submitText }}
            </el-button>
          </div>
        </el-form>
      </el-card>

      <!-- 侧栏：模板 -->
      <el-card shadow="never" class="send-page__side">
        <div class="side-panel__head">
          <h3>{{ t('send.templatesQuick') }}</h3>
          <router-link to="/templates" class="side-panel__link">
            {{ t('send.templatesManage') }}
          </router-link>
        </div>

        <div v-if="templates.length" class="side-panel__list">
          <button
            v-for="tpl in templates"
            :key="tpl.id"
            class="tpl"
            :class="{ 'is-picked': pickedTemplateId === tpl.id }"
            @click="applyTemplate(tpl)"
          >
            <div class="tpl__head">
              <span class="tpl__name">{{ tpl.name }}</span>
              <span class="tpl__type" :class="{ 'tpl__type--html': tpl.is_html }">
                {{ tpl.is_html ? 'HTML' : 'TEXT' }}
              </span>
            </div>
            <div class="tpl__subject">{{ tpl.subject || '— 无主题 —' }}</div>
            <div v-if="tpl.description" class="tpl__desc">{{ tpl.description }}</div>
          </button>
        </div>
        <EmptyHint v-else :icon="Files" :text="t('send.templatesEmpty')" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArrowRight,
  CircleCheckFilled,
  CircleCloseFilled,
  Files,
  Promotion,
  User
} from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { recipientsApi, sendApi, templatesApi } from '@/api'
import type {
  Recipient,
  RecipientGroup,
  SendResult,
  Template
} from '@/api/types'
import EmptyHint from './components/EmptyHint.vue'
import AttachmentPicker from './components/AttachmentPicker.vue'

type Mode = 'single' | 'batch' | 'scheduled'

const { t } = useI18n()

const mode = ref<Mode>('single')
const modeOptions = computed(() => [
  { label: t('send.modeSingle'), value: 'single' },
  { label: t('send.modeBatch'), value: 'batch' },
  { label: t('send.modeScheduled'), value: 'scheduled' }
])

const form = reactive({
  to: '',
  subject: '',
  body: '',
  is_html: false,
  track_enabled: false,
  from_name: '',
  cc: '',
  bcc: ''
})

const batchEmails = ref('')
const scheduledAt = ref<string>('')
const attachments = ref<Array<{ filename: string; content: string; size?: number }>>([])
const showMore = ref(false)

const formRef = ref<FormInstance>()
const loading = ref(false)
const result = ref<SendResult | null>(null)

// ---------- 收件人分组 ----------
const groups = ref<RecipientGroup[]>([])
const groupId = ref<string>('')
const recipientOptions = ref<Recipient[]>([])
const pickedEmail = ref<string>('')

async function onGroupChange() {
  recipientOptions.value = []
  pickedEmail.value = ''
  if (!groupId.value) return
  const list = await recipientsApi.listByGroup(groupId.value)
  recipientOptions.value = list.filter((r) => r.status === 'active')
}

function onPickRecipient() {
  if (pickedEmail.value) form.to = pickedEmail.value
}

function importGroup() {
  if (!recipientOptions.value.length) return
  const existing = new Set(
    batchEmails.value.split('\n').map((s) => s.trim()).filter(Boolean)
  )
  recipientOptions.value.forEach((r) => existing.add(r.email))
  batchEmails.value = Array.from(existing).join('\n')
  ElMessage.success(`已导入 ${recipientOptions.value.length} 个成员`)
}

// ---------- 模板 ----------
const templates = ref<Template[]>([])
const pickedTemplateId = ref<string>('')

function applyTemplate(tpl: Template) {
  form.subject = tpl.subject
  form.body = tpl.body
  form.is_html = tpl.is_html
  pickedTemplateId.value = tpl.id
  ElMessage.success(`已应用模板「${tpl.name}」`)
}

// ---------- 校验 ----------
const rules = computed<FormRules>(() => ({
  to: [
    {
      required: mode.value !== 'batch',
      message: '请输入收件人',
      trigger: 'blur'
    },
    {
      type: 'email' as const,
      message: '邮箱格式不正确',
      trigger: 'blur',
      // 批量模式不校验单字段
      validator: (_r: unknown, value: string, cb: (err?: Error) => void) => {
        if (mode.value === 'batch') return cb()
        if (!value) return cb()
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        return re.test(value) ? cb() : cb(new Error('邮箱格式不正确'))
      }
    }
  ],
  subject: [{ required: true, message: '请输入主题', trigger: 'blur' }],
  body: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}))

// ---------- 提交 ----------
const submitText = computed(() => {
  if (loading.value) return t('send.sending')
  if (mode.value === 'single') return t('send.submit')
  if (mode.value === 'batch') return t('send.submitBatch')
  return t('send.submitScheduled')
})

async function submit() {
  if (mode.value === 'batch') {
    const emails = batchEmails.value
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean)
    if (!emails.length) {
      ElMessage.error('请填写收件人列表')
      return
    }
    if (!form.subject || !form.body) {
      ElMessage.error('请填写主题与内容')
      return
    }
    loading.value = true
    result.value = null
    try {
      const res = await sendApi.sendBatch({
        emails,
        subject: form.subject,
        body: form.body,
        is_html: form.is_html,
        from_name: form.from_name || undefined,
        track_enabled: form.track_enabled,
        attachments: attachments.value.length ? attachments.value : undefined
      })
      result.value = res
      if (res.success) ElMessage.success(res.message)
    } finally {
      loading.value = false
    }
    return
  }

  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  if (mode.value === 'scheduled' && !scheduledAt.value) {
    ElMessage.error('请选择发送时间')
    return
  }

  loading.value = true
  result.value = null
  try {
    const basePayload = {
      to: form.to,
      subject: form.subject,
      body: form.body,
      is_html: form.is_html,
      track_enabled: form.track_enabled,
      from_name: form.from_name || undefined,
      cc: form.cc || undefined,
      bcc: form.bcc || undefined,
      attachments: attachments.value.length ? attachments.value : undefined
    }

    const res =
      mode.value === 'scheduled'
        ? await sendApi.sendScheduled({
            ...basePayload,
            scheduled_at: scheduledAt.value
          })
        : await sendApi.send(basePayload)

    result.value = res
    if (res.success) {
      ElMessage.success(res.message)
      if (mode.value !== 'scheduled') reset(true)
    }
  } finally {
    loading.value = false
  }
}

function reset(keepMeta = false) {
  form.to = ''
  form.subject = ''
  form.body = ''
  form.is_html = false
  form.track_enabled = false
  if (!keepMeta) {
    form.from_name = ''
    form.cc = ''
    form.bcc = ''
    attachments.value = []
    scheduledAt.value = ''
    batchEmails.value = ''
    groupId.value = ''
    recipientOptions.value = []
    pickedEmail.value = ''
    pickedTemplateId.value = ''
  }
  result.value = null
}

// 切换模式清除结果
watch(mode, () => {
  result.value = null
})

onMounted(async () => {
  try {
    const [gs, ts] = await Promise.all([
      recipientsApi.listGroups(),
      templatesApi.list()
    ])
    groups.value = gs
    templates.value = ts
  } catch {
    /* ignore */
  }
})
</script>

<style scoped>
.send-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.send-page__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
  flex-wrap: wrap;
}

.send-page__sub {
  font-size: 12.5px;
  color: var(--color-text-secondary);
}

.send-page__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 16px;
  align-items: start;
}

.send-page__main,
.send-page__side {
  border-radius: 12px;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-xs);
}

.send-page__main :deep(.el-card__body) { padding: 20px; }
.send-page__side :deep(.el-card__body) { padding: 18px; }

/* ---------- 表单 ---------- */
.send-form :deep(.el-form-item) { margin-bottom: 14px; }
.send-form :deep(.el-form-item__label) {
  padding-bottom: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.send-form__group-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
  padding: 12px 14px;
  background: var(--color-bg-subtle);
  border: 1px dashed var(--color-border);
  border-radius: 10px;
}

.send-form__field {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.send-form__field--action {
  justify-content: flex-end;
}

.send-form__label {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-bottom: 6px;
}

.send-form__body-opts {
  display: flex;
  gap: 16px;
  margin-top: 6px;
}

.send-form__body-opts :deep(.el-checkbox__label) {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.send-form__more {
  margin-bottom: 10px;
}

.send-form__more-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  padding: 0;
  font-size: 12px;
  color: var(--color-primary);
  cursor: pointer;
  font-weight: 500;
}

.send-form__more-btn :deep(.el-icon) {
  transition: transform var(--transition-fast);
}
.send-form__more-btn :deep(.el-icon.is-rotated) {
  transform: rotate(90deg);
}

.send-form__advanced {
  padding: 14px;
  background: var(--color-bg-subtle);
  border-radius: 10px;
  margin-bottom: 14px;
}

.send-form__row {
  display: flex;
  gap: 12px;
}
.send-form__half { flex: 1; min-width: 0; }

.send-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px dashed var(--color-border);
  margin-top: 6px;
}

/* 结果提示 */
.send-form__alert {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  font-size: 12.5px;
  margin-bottom: 14px;
  border: 1px solid transparent;
}

.send-form__alert.is-success {
  background: #dcfce7;
  color: #15803d;
  border-color: #bbf7d0;
}

.send-form__alert.is-error {
  background: #fee2e2;
  color: #b91c1c;
  border-color: #fecaca;
}

.send-form__alert :deep(.el-icon) { margin-top: 2px; flex-shrink: 0; }
.send-form__alert-body ul {
  margin: 6px 0 0;
  padding-left: 18px;
  font-size: 12px;
  opacity: 0.85;
}

/* ---------- 模板侧栏 ---------- */
.side-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px dashed var(--color-border);
}

.side-panel__head h3 {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.side-panel__link {
  font-size: 11.5px;
  color: var(--color-primary);
  font-weight: 500;
}

.side-panel__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 540px;
  overflow-y: auto;
}

.tpl {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.tpl:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
}

.tpl.is-picked {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
  box-shadow: var(--shadow-primary);
}

.tpl__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.tpl__name {
  font-family: var(--font-display);
  font-size: 12.5px;
  font-weight: 600;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.tpl__type {
  font-family: var(--font-mono);
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--color-bg-hover);
  color: var(--color-text-secondary);
  letter-spacing: 0.04em;
}

.tpl__type--html {
  background: #fef3c7;
  color: #b45309;
}

.tpl__subject {
  font-size: 11.5px;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tpl__desc {
  font-size: 11px;
  color: var(--color-text-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---------- 过渡 ---------- */
.fade-enter-active, .fade-leave-active { transition: opacity 160ms ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* ---------- 响应式 ---------- */
@media (max-width: 992px) {
  .send-page__grid { grid-template-columns: 1fr; }
  .send-page__side { order: 2; }
}

@media (max-width: 576px) {
  .send-form__group-row { grid-template-columns: 1fr; }
  .send-form__row { flex-direction: column; gap: 0; }
}
</style>
