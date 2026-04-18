<template>
  <div
    class="layout"
    :class="{
      'is-collapsed': appStore.sidebarCollapsed,
      'is-mobile-open': mobileOpen
    }"
  >
    <!-- 移动端遮罩 -->
    <div
      v-if="isMobile"
      class="layout__backdrop"
      :class="{ 'is-visible': mobileOpen }"
      @click="mobileOpen = false"
    />

    <!-- 侧边栏 -->
    <aside class="layout__sidebar">
      <div class="layout__brand" :class="{ 'is-compact': isCompact }">
        <picture v-if="!isCompact" class="layout__brand-banner">
          <source :srcset="logoWebp" type="image/webp" />
          <img :src="logoPng" alt="SMTP Lite" />
        </picture>
        <div v-else class="layout__logo">
          <el-icon :size="18"><Message /></el-icon>
        </div>
      </div>

      <nav class="layout__nav">
        <template v-for="group in navGroups" :key="group.key">
          <div v-if="!isCompact" class="layout__nav-group">
            {{ t(group.label) }}
          </div>
          <router-link
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            class="layout__nav-item"
            active-class="is-active"
            @click="onNavClick"
          >
            <el-icon :size="16"><component :is="item.icon" /></el-icon>
            <span v-if="!isCompact" class="layout__nav-label">{{ t(item.label) }}</span>
          </router-link>
        </template>
      </nav>

      <div class="layout__sidebar-foot">
        <button
          v-if="!isMobile"
          class="layout__collapse-btn"
          @click="appStore.toggleSidebar"
          :title="appStore.sidebarCollapsed ? t('common.expand') : t('common.collapse')"
        >
          <el-icon :size="16">
            <Fold v-if="!appStore.sidebarCollapsed" />
            <Expand v-else />
          </el-icon>
        </button>
      </div>
    </aside>

    <!-- 主体 -->
    <div class="layout__main">
      <header class="layout__topbar">
        <div class="layout__topbar-left">
          <button
            v-if="isMobile"
            class="layout__iconbtn layout__iconbtn--square"
            @click="mobileOpen = !mobileOpen"
            aria-label="menu"
          >
            <el-icon :size="18"><Menu /></el-icon>
          </button>
          <h1 class="layout__page-title">{{ pageTitle }}</h1>
        </div>
        <div class="layout__topbar-right">
          <el-dropdown trigger="click" @command="onLocaleChange">
            <button class="layout__iconbtn">
              <el-icon :size="16"><Operation /></el-icon>
              <span class="layout__locale-text">{{ localeLabel }}</span>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="zh-CN">简体中文</el-dropdown-item>
                <el-dropdown-item command="en-US">English</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="click" @command="onUserCommand">
            <button class="layout__user">
              <div class="layout__avatar">{{ avatarLetter }}</div>
              <span class="layout__username">{{ authStore.username || 'admin' }}</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  {{ t('nav.settings') }}
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  {{ t('common.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 全局强制更新横幅 -->
      <div
        v-if="versionStore.shouldShowForceBanner"
        class="layout__force-banner"
        role="alert"
      >
        <div class="layout__force-banner-inner">
          <el-icon :size="18" class="layout__force-banner-icon"><Warning /></el-icon>
          <div class="layout__force-banner-text">
            <strong>{{ t('forceBanner.title', { version: versionStore.latestVersion }) }}</strong>
            <span>{{ t('forceBanner.desc') }}</span>
          </div>
          <router-link to="/settings" class="layout__force-banner-cta">
            {{ t('forceBanner.cta') }}
          </router-link>
          <button
            class="layout__force-banner-close"
            :title="t('common.close')"
            @click="versionStore.dismissForceBanner"
          >
            <el-icon :size="14"><Close /></el-icon>
          </button>
        </div>
      </div>

      <main class="layout__content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, markRaw, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  DataAnalysis,
  Message,
  Promotion,
  Key,
  Files,
  User,
  Document,
  Link,
  CircleClose,
  Reading,
  Setting,
  SwitchButton,
  ArrowDown,
  Fold,
  Expand,
  Operation,
  Menu,
  Warning,
  Close
} from '@element-plus/icons-vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useVersionStore } from '@/stores/version'
import { setI18nLocale } from '@/i18n'
import logoWebp from '@/assets/logo.webp'
import logoPng from '@/assets/logo.png'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const versionStore = useVersionStore()

const mobileOpen = ref(false)
const isMobile = ref(false)

const mq = window.matchMedia('(max-width: 992px)')
const syncMobile = (e: MediaQueryListEvent | MediaQueryList) => {
  isMobile.value = 'matches' in e ? e.matches : (e as MediaQueryList).matches
  if (!isMobile.value) mobileOpen.value = false
}
syncMobile(mq)

onMounted(() => {
  mq.addEventListener?.('change', syncMobile as EventListener)
  // 启动时静默检测更新（失败不打扰用户）
  versionStore.checkUpdate(true)
})
onBeforeUnmount(() => mq.removeEventListener?.('change', syncMobile as EventListener))

// 移动端点击菜单后自动关闭抽屉
function onNavClick() {
  if (isMobile.value) mobileOpen.value = false
}

// 在 PC 模式下：sidebar collapsed 决定是否紧凑；移动端抽屉打开时展开完整内容
const isCompact = computed(() => (isMobile.value ? false : appStore.sidebarCollapsed))

const navGroups = [
  {
    key: 'core',
    label: 'nav.group.core',
    items: [
      { to: '/stats', label: 'nav.stats', icon: markRaw(DataAnalysis) },
      { to: '/send', label: 'nav.send', icon: markRaw(Promotion) },
      { to: '/smtp', label: 'nav.smtp', icon: markRaw(Message) }
    ]
  },
  {
    key: 'resources',
    label: 'nav.group.resources',
    items: [
      { to: '/keys', label: 'nav.keys', icon: markRaw(Key) },
      { to: '/templates', label: 'nav.templates', icon: markRaw(Files) },
      { to: '/recipients', label: 'nav.recipients', icon: markRaw(User) },
      { to: '/logs', label: 'nav.logs', icon: markRaw(Document) }
    ]
  },
  {
    key: 'advanced',
    label: 'nav.group.advanced',
    items: [
      { to: '/webhooks', label: 'nav.webhooks', icon: markRaw(Link) },
      { to: '/blacklist', label: 'nav.blacklist', icon: markRaw(CircleClose) },
      { to: '/docs', label: 'nav.docs', icon: markRaw(Reading) }
    ]
  },
  {
    key: 'system',
    label: 'nav.group.system',
    items: [{ to: '/settings', label: 'nav.settings', icon: markRaw(Setting) }]
  }
]

const pageTitle = computed(() => {
  const key = route.meta.title as string | undefined
  return key ? t(key) : ''
})

const localeLabel = computed(() => (locale.value === 'zh-CN' ? '中文' : 'EN'))

const avatarLetter = computed(() =>
  (authStore.username || 'A').charAt(0).toUpperCase()
)

function onLocaleChange(cmd: 'zh-CN' | 'en-US') {
  appStore.setLocale(cmd)
  setI18nLocale(cmd)
}

function onUserCommand(cmd: string) {
  if (cmd === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (cmd === 'settings') {
    router.push('/settings')
  }
}
</script>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: 232px 1fr;
  min-height: 100vh;
  background: var(--color-bg-page);
  transition: grid-template-columns var(--transition);
}

/* ---------- 强制更新全局横幅 ---------- */
.layout__force-banner {
  background: linear-gradient(135deg, #b91c1c, #dc2626);
  color: #fff;
}

.layout__force-banner-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: 1200px;
  margin: 0 auto;
  padding: 10px 20px;
}

.layout__force-banner-icon { flex-shrink: 0; color: #fff; }

.layout__force-banner-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12.5px;
  line-height: 1.4;
}

.layout__force-banner-text strong {
  font-weight: 600;
  letter-spacing: 0.01em;
}

.layout__force-banner-text span {
  opacity: 0.9;
  font-size: 11.5px;
}

.layout__force-banner-cta {
  flex-shrink: 0;
  padding: 6px 14px;
  background: #fff;
  color: #b91c1c;
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
  text-decoration: none;
  transition: background var(--transition-fast);
}

.layout__force-banner-cta:hover { background: #fef2f2; }

.layout__force-banner-close {
  flex-shrink: 0;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast);
}

.layout__force-banner-close:hover { background: rgba(255, 255, 255, 0.28); }

@media (max-width: 576px) {
  .layout__force-banner-inner { flex-wrap: wrap; padding: 10px 14px; }
  .layout__force-banner-text { order: 3; flex-basis: 100%; }
}

.layout.is-collapsed {
  grid-template-columns: 64px 1fr;
}

/* ---------- 侧边栏 ---------- */
.layout__sidebar {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: hidden;
  z-index: 20;
}

.layout__brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 72px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--color-border-light);
}

.layout__brand.is-compact {
  padding: 10px 0;
}

.layout__brand-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 180px;
  max-height: 52px;
}

.layout__brand-banner img {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
  display: block;
  user-select: none;
  -webkit-user-drag: none;
}

.layout__logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff;
  flex-shrink: 0;
}

.layout__brand-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.layout__brand-name {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.layout__brand-tagline {
  font-size: 11px;
  color: var(--color-text-secondary);
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.layout__nav {
  flex: 1;
  overflow-y: auto;
  padding: 12px 10px 8px;
  display: flex;
  flex-direction: column;
}

.layout__nav-group {
  padding: 14px 10px 6px;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-placeholder);
}

.layout__nav-group:first-child { padding-top: 4px; }

.layout__nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  margin-bottom: 2px;
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-text-regular);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
}

.layout__nav-item:hover {
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
}

.layout__nav-item.is-active {
  background: var(--color-primary);
  color: #fff;
  font-weight: 500;
  box-shadow: var(--shadow-primary);
}

.layout__nav-label {
  overflow: hidden;
  text-overflow: ellipsis;
}

.layout.is-collapsed .layout__nav-item {
  justify-content: center;
  padding: 10px 0;
}

.layout__sidebar-foot {
  padding: 10px;
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: flex-end;
  min-height: 50px;
}

.layout.is-collapsed .layout__sidebar-foot {
  justify-content: center;
}

.layout__collapse-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: #fff;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: border-color var(--transition-fast), color var(--transition-fast);
}

.layout__collapse-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

/* ---------- 主体 ---------- */
.layout__main {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.layout__topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 24px;
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: saturate(180%) blur(10px);
  border-bottom: 1px solid var(--color-border);
  min-height: 56px;
}

.layout__topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.layout__page-title {
  font-family: var(--font-display);
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.layout__topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.layout__iconbtn,
.layout__user {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 12px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: #fff;
  color: var(--color-text-regular);
  font-size: 13px;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.layout__iconbtn--square { padding: 0; width: 34px; justify-content: center; }

.layout__iconbtn:hover,
.layout__user:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-lighter);
  color: var(--color-primary-dark);
}

.layout__locale-text { font-size: 12px; font-weight: 500; }

.layout__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #60a5fa);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-display);
}

.layout__username {
  font-size: 13px;
  font-weight: 500;
}

.layout__content {
  flex: 1;
  padding: 20px 24px 32px;
  min-height: 0;
}

/* ---------- 移动端遮罩 ---------- */
.layout__backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  z-index: 19;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--transition);
}

.layout__backdrop.is-visible {
  opacity: 1;
  pointer-events: auto;
}

/* ---------- 过渡 ---------- */
.fade-enter-active,
.fade-leave-active { transition: opacity 140ms ease; }
.fade-enter-from,
.fade-leave-to { opacity: 0; }

/* ---------- 响应式 ---------- */

/* 992px 以下：侧边栏变成抽屉 */
@media (max-width: 992px) {
  .layout { grid-template-columns: 1fr; }
  .layout.is-collapsed { grid-template-columns: 1fr; }

  .layout__sidebar {
    position: fixed;
    top: 0;
    left: 0;
    width: 260px;
    height: 100vh;
    transform: translateX(-100%);
    transition: transform var(--transition);
    box-shadow: var(--shadow-lg);
  }

  .layout.is-mobile-open .layout__sidebar {
    transform: translateX(0);
  }

  .layout__sidebar-foot { display: none; }
  .layout__nav-group { display: block; }
  .layout__nav-label { display: inline; }
  .layout__nav-item { justify-content: flex-start; padding: 8px 10px; }
}

@media (max-width: 768px) {
  .layout__content { padding: 16px; }
  .layout__topbar { padding: 10px 14px; }
  .layout__username { display: none; }
  .layout__page-title { font-size: 15px; }
}
</style>
