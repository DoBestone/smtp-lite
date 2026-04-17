<template>
  <div class="settings-page">
    <p class="settings-page__sub">{{ t('settings.subtitle') }}</p>

    <!-- 版本信息 · 并排卡片 -->
    <section class="settings-section">
      <div class="settings-section__head">
        <h3>{{ t('settings.versionTitle') }}</h3>
        <el-button
          size="small"
          :icon="Refresh"
          :loading="checking"
          @click="checkUpdate"
        >
          {{ checking ? t('settings.checking') : t('settings.checkUpdate') }}
        </el-button>
      </div>

      <div class="ver-grid">
        <div class="ver-cell">
          <div class="ver-cell__label">{{ t('settings.currentVersion') }}</div>
          <div class="ver-cell__value mono">{{ currentVersion || '—' }}</div>
        </div>
        <div class="ver-cell">
          <div class="ver-cell__label">{{ t('settings.latestVersion') }}</div>
          <div class="ver-cell__value mono">
            {{ latestVersion || '—' }}
            <StatusPill
              v-if="updateStatus === 'available'"
              kind="warning"
              class="ver-cell__tag"
            >{{ t('settings.hasUpdate') }}</StatusPill>
            <StatusPill
              v-else-if="updateStatus === 'latest'"
              kind="success"
              class="ver-cell__tag"
            >{{ t('settings.isLatest') }}</StatusPill>
          </div>
        </div>
      </div>

      <!-- 更新提示条 -->
      <div
        v-if="updateStatus === 'available'"
        class="update-banner"
      >
        <div class="update-banner__icon">
          <el-icon :size="18"><Upload /></el-icon>
        </div>
        <div class="update-banner__text">
          <div class="update-banner__title">
            {{ tFmt('settings.updateAvailable', { version: latestVersion }) }}
          </div>
          <div class="update-banner__desc">{{ t('settings.updateAvailableDesc') }}</div>
        </div>
        <el-button
          type="primary"
          :loading="updating"
          @click="doUpdate"
        >
          {{ updating ? t('settings.updating') : t('settings.doUpdate') }}
        </el-button>
      </div>

      <!-- 更新日志预览 -->
      <div v-if="changelog" class="changelog">
        <div class="changelog__head">
          <el-icon><Document /></el-icon>
          <span>{{ t('settings.changelogTitle') }}</span>
        </div>
        <pre class="changelog__body">{{ changelog }}</pre>
      </div>
    </section>

    <!-- 账户 -->
    <section class="settings-section">
      <div class="settings-section__head">
        <h3>{{ t('settings.accountTitle') }}</h3>
        <span class="settings-section__meta mono">{{ authStore.username || 'admin' }}</span>
      </div>

      <div class="settings-list">
        <button class="settings-row" @click="usernameVisible = true">
          <div class="settings-row__icon"><el-icon><User /></el-icon></div>
          <div class="settings-row__main">
            <div class="settings-row__title">{{ t('settings.changeUsername') }}</div>
            <div class="settings-row__desc">{{ t('settings.changeUsernameDesc') }}</div>
          </div>
          <el-icon class="settings-row__arrow"><ArrowRight /></el-icon>
        </button>

        <button class="settings-row" @click="passwordVisible = true">
          <div class="settings-row__icon"><el-icon><Lock /></el-icon></div>
          <div class="settings-row__main">
            <div class="settings-row__title">{{ t('settings.changePassword') }}</div>
            <div class="settings-row__desc">{{ t('settings.changePasswordDesc') }}</div>
          </div>
          <el-icon class="settings-row__arrow"><ArrowRight /></el-icon>
        </button>

        <div class="settings-row settings-row--inline">
          <div class="settings-row__icon"><el-icon><Operation /></el-icon></div>
          <div class="settings-row__main">
            <div class="settings-row__title">{{ t('settings.language') }}</div>
            <div class="settings-row__desc">{{ t('settings.languageDesc') }}</div>
          </div>
          <el-select
            :model-value="locale"
            @change="onLocaleChange"
            size="small"
            style="width: 140px"
          >
            <el-option label="简体中文" value="zh-CN" />
            <el-option label="English" value="en-US" />
          </el-select>
        </div>
      </div>
    </section>

    <!-- 修改用户名弹窗 -->
    <el-dialog
      v-model="usernameVisible"
      :title="t('settings.changeUsername')"
      width="420px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('settings.usernameNew')">
          <el-input v-model="usernameForm.newUsername" autocomplete="username" />
        </el-form-item>
        <el-form-item :label="t('settings.passwordCurrent')">
          <el-input
            v-model="usernameForm.password"
            type="password"
            show-password
            autocomplete="current-password"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="usernameVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submittingUsername" @click="submitUsername">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改密码弹窗 -->
    <el-dialog
      v-model="passwordVisible"
      :title="t('settings.changePassword')"
      width="420px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('settings.passwordCurrent')">
          <el-input
            v-model="passwordForm.old"
            type="password"
            show-password
            autocomplete="current-password"
          />
        </el-form-item>
        <el-form-item :label="t('settings.passwordNew')">
          <el-input
            v-model="passwordForm.next"
            type="password"
            show-password
            autocomplete="new-password"
          />
        </el-form-item>
        <el-form-item :label="t('settings.passwordConfirm')">
          <el-input
            v-model="passwordForm.confirm"
            type="password"
            show-password
            autocomplete="new-password"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submittingPassword" @click="submitPassword">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ArrowRight,
  Document,
  Lock,
  Operation,
  Refresh,
  Upload,
  User
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { authApi, systemApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { setI18nLocale } from '@/i18n'
import StatusPill from './components/StatusPill.vue'

const { t, locale } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ---------- 版本 ----------
const currentVersion = ref('')
const latestVersion = ref('')
const changelog = ref('')
const checking = ref(false)
const updating = ref(false)

const updateStatus = computed<'available' | 'latest' | ''>(() => {
  if (!currentVersion.value || !latestVersion.value) return ''
  return currentVersion.value === latestVersion.value ? 'latest' : 'available'
})

async function loadVersion() {
  try {
    const v = await systemApi.version()
    currentVersion.value = v.version
  } catch {
    /* ignore */
  }
}

async function loadChangelog() {
  try {
    const c = await systemApi.changelog()
    changelog.value = c.changelog?.trim() ?? ''
  } catch {
    /* ignore */
  }
}

async function checkUpdate() {
  checking.value = true
  try {
    const r = await systemApi.updateCheck()
    latestVersion.value = r.latest_version ?? ''
    if (r.current_version) currentVersion.value = r.current_version
  } finally {
    checking.value = false
  }
}

async function doUpdate() {
  const confirmed = await ElMessageBox.confirm(
    `确定升级到 ${latestVersion.value}？升级过程中服务会短暂不可用。`,
    { type: 'warning', confirmButtonText: t('settings.doUpdate'), cancelButtonText: t('common.cancel') }
  ).catch(() => false)
  if (!confirmed) return

  updating.value = true
  try {
    await systemApi.updatePrepare()
    await systemApi.update({ confirm: true })
    ElMessage.success('更新任务已启动，请稍候刷新')
  } finally {
    updating.value = false
  }
}

// ---------- 用户名弹窗 ----------
const usernameVisible = ref(false)
const submittingUsername = ref(false)
const usernameForm = reactive({ newUsername: '', password: '' })

async function submitUsername() {
  if (!usernameForm.newUsername.trim() || !usernameForm.password) {
    ElMessage.error('请填写完整')
    return
  }
  submittingUsername.value = true
  try {
    await authApi.changeUsername(usernameForm.newUsername.trim(), usernameForm.password)
    ElMessage.success(t('settings.changed'))
    usernameVisible.value = false
    authStore.logout()
    router.push('/login')
  } finally {
    submittingUsername.value = false
  }
}

// ---------- 密码弹窗 ----------
const passwordVisible = ref(false)
const submittingPassword = ref(false)
const passwordForm = reactive({ old: '', next: '', confirm: '' })

async function submitPassword() {
  if (!passwordForm.old || !passwordForm.next) {
    ElMessage.error('请填写完整')
    return
  }
  if (passwordForm.next !== passwordForm.confirm) {
    ElMessage.error(t('settings.passwordMismatch'))
    return
  }
  submittingPassword.value = true
  try {
    await authApi.changePassword(passwordForm.old, passwordForm.next)
    ElMessage.success(t('settings.changed'))
    passwordVisible.value = false
    authStore.logout()
    router.push('/login')
  } finally {
    submittingPassword.value = false
  }
}

// ---------- 语言 ----------
function onLocaleChange(v: unknown) {
  const code = v as 'zh-CN' | 'en-US'
  appStore.setLocale(code)
  setI18nLocale(code)
}

// ---------- i18n 插值辅助 ----------
function tFmt(key: string, values: Record<string, string>): string {
  return t(key, values)
}

onMounted(() => {
  loadVersion()
  loadChangelog()
  checkUpdate()
})
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 760px;
}

.settings-page__sub {
  font-size: 12.5px;
  color: var(--color-text-secondary);
}

/* ---------- Section ---------- */
.settings-section {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 18px 20px;
  box-shadow: var(--shadow-xs);
}

.settings-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--color-border);
  margin-bottom: 14px;
}

.settings-section__head h3 {
  font-family: var(--font-display);
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.settings-section__meta {
  font-size: 11.5px;
  color: var(--color-text-secondary);
}

/* ---------- 版本 ---------- */
.ver-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-bottom: 14px;
}

.ver-cell {
  padding: 12px 14px;
  background: var(--color-bg-subtle);
  border-radius: 10px;
}

.ver-cell__label {
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-secondary);
  letter-spacing: 0.02em;
  margin-bottom: 4px;
}

.ver-cell__value {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--font-mono);
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.ver-cell__tag { flex-shrink: 0; }

/* 更新提示条 */
.update-banner {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  border: 1px solid #bfdbfe;
  margin-bottom: 14px;
}

.update-banner__icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary);
  color: #fff;
  box-shadow: var(--shadow-primary);
}

.update-banner__text { flex: 1; }

.update-banner__title {
  font-family: var(--font-display);
  font-size: 13.5px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 2px;
}

.update-banner__desc {
  font-size: 12px;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 更新日志 */
.changelog {
  margin-top: 14px;
  padding: 12px 14px;
  background: var(--color-bg-subtle);
  border-radius: 10px;
  border: 1px solid var(--color-border);
}

.changelog__head {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11.5px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: 8px;
  letter-spacing: 0.02em;
}

.changelog__body {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--color-text-regular);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.65;
  max-height: 240px;
  overflow: auto;
}

/* ---------- 账户列表 ---------- */
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.settings-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.settings-row:not(.settings-row--inline):hover {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
}

.settings-row--inline { cursor: default; }

.settings-row__icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary-lighter);
  color: var(--color-primary);
  flex-shrink: 0;
}

.settings-row__main { flex: 1; min-width: 0; }

.settings-row__title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-primary);
  line-height: 1.35;
}

.settings-row__desc {
  font-size: 11.5px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.settings-row__arrow {
  color: var(--color-text-placeholder);
  font-size: 13px;
  transition: transform var(--transition-fast), color var(--transition-fast);
}

.settings-row:hover .settings-row__arrow {
  color: var(--color-primary);
  transform: translateX(2px);
}

@media (max-width: 576px) {
  .ver-grid { grid-template-columns: 1fr; }
  .update-banner { flex-direction: column; align-items: flex-start; }
}
</style>
