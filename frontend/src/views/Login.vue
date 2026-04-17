<template>
  <div class="login">
    <!-- 左侧：装饰 Hero · 仅桌面显示 -->
    <aside class="login__hero">
      <div class="login__mesh" aria-hidden="true"></div>
      <div class="login__noise" aria-hidden="true"></div>

      <div class="login__hero-inner">
        <div class="login__brand">
          <div class="login__logo">
            <el-icon :size="20"><Message /></el-icon>
          </div>
          <div class="login__brand-text">
            <div class="login__brand-name">SMTP Lite</div>
            <div class="login__brand-tag">v2.x · 聚合发信控制台</div>
          </div>
        </div>

        <div class="login__copy">
          <h1 class="login__copy-title">
            一个后台<br />
            管住所有<span class="login__accent">发信账号</span>
          </h1>
          <p class="login__copy-desc">
            自动轮询 · API Key 鉴权 · 模板与收件人组 · Webhook 与打开点击追踪
          </p>
        </div>

        <ul class="login__features">
          <li>
            <span class="login__bullet"><el-icon><Connection /></el-icon></span>
            <div>
              <div class="login__feature-title">多账号聚合</div>
              <div class="login__feature-desc">按日配额自动轮询，单账号故障不影响整体</div>
            </div>
          </li>
          <li>
            <span class="login__bullet"><el-icon><Lock /></el-icon></span>
            <div>
              <div class="login__feature-title">API Key 精细管理</div>
              <div class="login__feature-desc">可独立重置与吊销，最近一次使用时间可追</div>
            </div>
          </li>
          <li>
            <span class="login__bullet"><el-icon><DataLine /></el-icon></span>
            <div>
              <div class="login__feature-title">开信点击全链路</div>
              <div class="login__feature-desc">像素追踪 + 链接重写，Webhook 事件实时回推</div>
            </div>
          </li>
        </ul>

        <footer class="login__hero-foot">
          <span class="login__dot"></span>
          <span>系统运行正常</span>
          <span class="login__sep">·</span>
          <span class="login__mono">{{ nowLabel }}</span>
        </footer>
      </div>
    </aside>

    <!-- 右侧：登录卡 -->
    <section class="login__form-col">
      <div class="login__form-wrap">
        <div class="login__form-head">
          <div class="login__brand login__brand--mobile">
            <div class="login__logo">
              <el-icon :size="18"><Message /></el-icon>
            </div>
            <span class="login__brand-name">SMTP Lite</span>
          </div>

          <h2 class="login__title">{{ t('login.title') }}</h2>
          <p class="login__subtitle">{{ t('login.subtitle') }}</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="login__form"
          @submit.prevent="submit"
        >
          <el-form-item :label="t('login.username')" prop="username">
            <el-input
              v-model="form.username"
              size="large"
              :placeholder="t('login.username')"
              autocomplete="username"
            >
              <template #prefix><el-icon><User /></el-icon></template>
            </el-input>
          </el-form-item>

          <el-form-item :label="t('login.password')" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              size="large"
              show-password
              :placeholder="t('login.password')"
              autocomplete="current-password"
              @keyup.enter="submit"
            >
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>

          <el-button
            type="primary"
            class="login__submit"
            :loading="loading"
            native-type="submit"
            @click="submit"
          >
            {{ t('login.submit') }}
            <el-icon v-if="!loading" class="login__submit-arrow"><Right /></el-icon>
          </el-button>
        </el-form>

        <div class="login__hint">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ t('login.defaultHint') }}</span>
        </div>

        <div class="login__lang">
          <button
            v-for="opt in langs"
            :key="opt.code"
            class="login__lang-btn"
            :class="{ 'is-active': locale === opt.code }"
            @click="switchLocale(opt.code)"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  Connection,
  DataLine,
  InfoFilled,
  Lock,
  Message,
  Right,
  User
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { setI18nLocale } from '@/i18n'

const { t, locale } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const langs = [
  { code: 'zh-CN' as const, label: '简体中文' },
  { code: 'en-US' as const, label: 'English' }
]

function switchLocale(code: 'zh-CN' | 'en-US') {
  appStore.setLocale(code)
  setI18nLocale(code)
}

const nowLabel = ref('')
let timer: number | null = null

function updateNow() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  nowLabel.value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(() => {
  updateNow()
  timer = window.setInterval(updateNow, 30_000)
})
onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/stats'
    router.replace(redirect)
  } catch {
    // 拦截器已处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login {
  display: grid;
  grid-template-columns: 1.05fr 1fr;
  min-height: 100vh;
  background: #fff;
}

/* ============================================================
   Hero 侧
   ============================================================ */

.login__hero {
  position: relative;
  overflow: hidden;
  background: #0b1a36;
  color: #fff;
  padding: 48px 56px;
  display: flex;
}

.login__mesh {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(60% 55% at 18% 22%, rgba(59, 130, 246, 0.55) 0%, transparent 60%),
    radial-gradient(50% 45% at 82% 18%, rgba(14, 165, 233, 0.35) 0%, transparent 65%),
    radial-gradient(60% 50% at 70% 90%, rgba(37, 99, 235, 0.45) 0%, transparent 60%),
    linear-gradient(145deg, #0b1a36 0%, #0d2654 55%, #0b1a36 100%);
}

.login__noise {
  position: absolute;
  inset: 0;
  opacity: 0.08;
  mix-blend-mode: overlay;
  pointer-events: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='220' height='220'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
}

.login__hero-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 36px;
  max-width: 480px;
  margin: auto 0;
  width: 100%;
  animation: rise 480ms cubic-bezier(.2,.7,.2,1) both;
}

@keyframes rise {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: none; }
}

.login__brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login__brand--mobile { display: none; }

.login__logo {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  color: #fff;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.35);
}

.login__brand-text { line-height: 1.3; }
.login__brand-name {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.login__brand-tag {
  font-family: var(--font-mono);
  font-size: 11px;
  color: rgba(255, 255, 255, 0.55);
  letter-spacing: 0.02em;
}

.login__copy-title {
  font-family: var(--font-display);
  font-size: clamp(30px, 3.4vw, 44px);
  font-weight: 700;
  line-height: 1.15;
  letter-spacing: -0.02em;
  color: #fff;
}

.login__accent {
  background: linear-gradient(135deg, #60a5fa, #7dd3fc);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.login__copy-desc {
  margin-top: 14px;
  max-width: 420px;
  font-size: 14px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.7);
}

.login__features {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.login__features li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(6px);
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.login__features li:hover {
  border-color: rgba(96, 165, 250, 0.45);
  background: rgba(59, 130, 246, 0.08);
}

.login__bullet {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(59, 130, 246, 0.16);
  color: #7dd3fc;
  border: 1px solid rgba(96, 165, 250, 0.3);
}

.login__feature-title {
  font-family: var(--font-display);
  font-size: 13.5px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 3px;
}

.login__feature-desc {
  font-size: 12px;
  line-height: 1.55;
  color: rgba(255, 255, 255, 0.6);
}

.login__hero-foot {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11.5px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: auto;
  padding-top: 18px;
}

.login__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.18);
}

.login__sep { color: rgba(255, 255, 255, 0.3); }
.login__mono { font-family: var(--font-mono); letter-spacing: 0.02em; }

/* ============================================================
   Form 侧
   ============================================================ */

.login__form-col {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 32px;
  background:
    radial-gradient(circle at 100% 0%, rgba(59, 130, 246, 0.05), transparent 40%),
    #fff;
}

.login__form-wrap {
  width: 100%;
  max-width: 380px;
  animation: rise 480ms cubic-bezier(.2,.7,.2,1) 80ms both;
}

.login__form-head { margin-bottom: 24px; }

.login__title {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 600;
  letter-spacing: -0.015em;
  color: var(--color-text-primary);
  margin-bottom: 6px;
}

.login__subtitle {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.login__form :deep(.el-form-item__label) {
  padding-bottom: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.login__form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.login__form :deep(.el-input__wrapper) {
  padding: 2px 14px;
  border-radius: 10px;
  transition: box-shadow var(--transition-fast);
}

.login__submit {
  width: 100%;
  height: 46px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.01em;
  margin-top: 6px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border: none;
  box-shadow: 0 6px 18px rgba(59, 130, 246, 0.28);
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.login__submit:hover {
  box-shadow: 0 8px 22px rgba(59, 130, 246, 0.38);
}

.login__submit:active { transform: translateY(1px); }

.login__submit-arrow { margin-left: 2px; }

.login__hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 18px;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
  font-size: 11.5px;
  line-height: 1.5;
}

.login__hint :deep(.el-icon) { margin-top: 2px; flex-shrink: 0; }

.login__lang {
  display: flex;
  gap: 6px;
  justify-content: center;
  margin-top: 24px;
  padding-top: 18px;
  border-top: 1px dashed var(--color-border);
}

.login__lang-btn {
  padding: 4px 10px;
  background: transparent;
  border: none;
  font-size: 11.5px;
  color: var(--color-text-placeholder);
  cursor: pointer;
  border-radius: 4px;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.login__lang-btn:hover {
  color: var(--color-primary);
  background: var(--color-primary-lighter);
}

.login__lang-btn.is-active {
  color: var(--color-primary);
  font-weight: 600;
}

/* ============================================================
   响应式
   ============================================================ */

@media (max-width: 992px) {
  .login { grid-template-columns: 1fr; }
  .login__hero { display: none; }
  .login__brand--mobile { display: flex; margin-bottom: 18px; }
}

@media (max-width: 576px) {
  .login__form-col { padding: 28px 20px; }
  .login__title { font-size: 20px; }
}
</style>
