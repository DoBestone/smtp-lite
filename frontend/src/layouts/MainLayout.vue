<template>
  <div class="app-layout">
    <!-- 移动端遮罩 -->
    <transition name="fade">
      <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>
    </transition>
    
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ open: sidebarOpen }">
      <!-- Logo -->
      <div class="sidebar-logo">
        <span class="logo-icon">📧</span>
        <div class="logo-text">
          <span class="logo-title">SMTP Lite</span>
          <span class="logo-subtitle">个人邮箱聚合</span>
        </div>
      </div>
      
      <!-- 导航菜单 -->
      <nav class="sidebar-nav">
        <div class="nav-section">
          <span class="nav-section-title">核心功能</span>
          <router-link to="/smtp" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/>
                <path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8"/>
              </svg>
            </span>
            <span class="nav-label">SMTP 账号</span>
          </router-link>
          
          <router-link to="/send" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </span>
            <span class="nav-label">发送邮件</span>
          </router-link>
          
          <router-link to="/logs" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">发送日志</span>
          </router-link>
          
          <router-link to="/stats" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M18 20V10M12 20V4M6 20v-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
            </span>
            <span class="nav-label">统计数据</span>
          </router-link>
        </div>
        
        <div class="nav-section">
          <span class="nav-section-title">资源管理</span>
          <router-link to="/keys" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">API Key</span>
          </router-link>
          
          <router-link to="/templates" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5"/>
                <polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">邮件模板</span>
          </router-link>
          
          <router-link to="/recipients" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" stroke="currentColor" stroke-width="1.5"/>
                <circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">收件人</span>
          </router-link>
        </div>
        
        <div class="nav-section">
          <span class="nav-section-title">高级功能</span>
          <router-link to="/webhooks" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">Webhook</span>
          </router-link>
          
          <router-link to="/blacklist" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
                <line x1="4.93" y1="4.93" x2="19.07" y2="19.07" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">黑名单</span>
          </router-link>
          
          <router-link to="/docs" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5"/>
                <polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">API 文档</span>
          </router-link>
        </div>
        
        <div class="nav-section nav-section-bottom">
          <router-link to="/settings" class="nav-item">
            <span class="nav-icon">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.5"/>
                <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </span>
            <span class="nav-label">系统设置</span>
          </router-link>
        </div>
      </nav>
      
      <!-- 底部统计 -->
      <div class="sidebar-footer">
        <div class="today-stats">
          <span class="stats-dot"></span>
          <span>今日发送 <strong>{{ stats.today_sent || 0 }}</strong> 封</span>
        </div>
      </div>
    </aside>
    
    <!-- 主内容区 -->
    <div class="main-wrapper">
      <!-- 顶部导航 -->
      <header class="topbar">
        <div class="topbar-left">
          <button class="menu-toggle" @click="sidebarOpen = !sidebarOpen">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
              <path d="M3 12h18M3 6h18M3 18h18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </button>
          <h1 class="page-title">{{ pageTitle }}</h1>
        </div>
        <div class="topbar-right">
          <div class="quick-stats">
            <span class="stat-item">
              <span class="stat-dot success"></span>
              成功率 {{ (stats.success_rate || 0).toFixed(1) }}%
            </span>
          </div>
          <div class="topbar-actions">
            <button class="action-btn logout" @click="logout">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>退出</span>
            </button>
          </div>
        </div>
      </header>
      
      <!-- 页面内容 -->
      <main class="page-content">
        <router-view />
      </main>
    </div>
    
    <!-- Toast -->
    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="toast.type">
        {{ toast.msg }}
      </div>
    </transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { store, actions } from '@/store'

const pageTitles = {
  '/smtp': 'SMTP 账号管理',
  '/send': '发送邮件',
  '/keys': 'API Key 管理',
  '/templates': '邮件模板',
  '/recipients': '收件人分组',
  '/logs': '发送日志',
  '/stats': '统计数据',
  '/webhooks': 'Webhook 回调',
  '/blacklist': '黑名单管理',
  '/docs': 'API 文档',
  '/settings': '系统设置'
}

export default {
  name: 'MainLayout',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const sidebarOpen = ref(false)
    
    const pageTitle = computed(() => pageTitles[route.path] || '')
    const stats = computed(() => store.stats)
    const toast = computed(() => store.toast)
    
    // 路由变化时关闭移动端侧边栏
    watch(() => route.path, () => {
      sidebarOpen.value = false
    })
    
    const logout = () => {
      actions.logout()
      router.push('/login')
    }
    
    onMounted(() => {
      actions.loadStats()
    })
    
    return { pageTitle, stats, toast, logout, sidebarOpen }
  }
}
</script>

<style scoped>
* { box-sizing: border-box; }

.app-layout {
  display: flex;
  min-height: 100vh;
  background: #f8fafc;
}

/* ========== 侧边栏 ========== */
.sidebar {
  width: 260px;
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  color: #fff;
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
  overflow: hidden;
}

.sidebar-logo {
  padding: 24px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

.logo-icon {
  font-size: 32px;
  line-height: 1;
}

.logo-text {
  display: flex;
  flex-direction: column;
}

.logo-title {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.logo-subtitle {
  font-size: 11px;
  color: rgba(255,255,255,0.5);
  margin-top: 2px;
}

/* 导航菜单 */
.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: 24px;
}

.nav-section-title {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 0 12px;
  margin-bottom: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: rgba(255,255,255,0.7);
  text-decoration: none;
  border-radius: 8px;
  margin-bottom: 4px;
  transition: all 0.2s;
  font-size: 14px;
}

.nav-item:hover {
  background: rgba(255,255,255,0.08);
  color: #fff;
}

.nav-item.router-link-active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.nav-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.9;
}

.nav-item.router-link-active .nav-icon {
  opacity: 1;
}

/* 底部导航区域 */
.nav-section-bottom {
  margin-top: auto;
  padding-top: 16px;
  border-top: 1px solid rgba(255,255,255,0.08);
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid rgba(255,255,255,0.08);
}

.today-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: rgba(255,255,255,0.6);
}

.stats-dot {
  width: 8px;
  height: 8px;
  background: #22c55e;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* ========== 主内容区 ========== */
.main-wrapper {
  flex: 1;
  margin-left: 260px;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

/* 顶部导航 */
.topbar {
  background: #fff;
  padding: 0 32px;
  height: 64px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 50;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.quick-stats {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #64748b;
}

.stat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.stat-dot.success { background: #22c55e; }

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
}

.action-btn.logout:hover {
  background: #fef2f2;
  border-color: #fecaca;
  color: #dc2626;
}

/* 页面内容 */
.page-content {
  flex: 1;
  padding: 24px 32px;
}

/* Toast */
.toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.toast.success {
  background: #dcfce7;
  color: #166534;
}

.toast.error {
  background: #fee2e2;
  color: #991b1b;
}

.toast-fade-enter-active, .toast-fade-leave-active {
  transition: all 0.3s ease;
}

.toast-fade-enter-from, .toast-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}

/* 汉堡菜单按钮 - 默认隐藏 */
.menu-toggle {
  display: none;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: none;
  border: none;
  border-radius: 8px;
  color: #475569;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.2s;
}

.menu-toggle:hover {
  background: #f1f5f9;
}

/* 遮罩层 - 默认隐藏 */
.sidebar-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 99;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

/* 响应式 */
@media (max-width: 1024px) {
  .sidebar {
    width: 72px;
  }
  
  .sidebar .logo-text,
  .sidebar .nav-section-title,
  .sidebar .nav-label,
  .sidebar-footer {
    display: none;
  }
  
  .sidebar-logo {
    justify-content: center;
    padding: 20px;
  }
  
  .nav-item {
    justify-content: center;
    padding: 12px;
  }
  
  .main-wrapper {
    margin-left: 72px;
  }
}

@media (max-width: 768px) {
  .menu-toggle {
    display: flex;
  }
  
  .sidebar-overlay {
    display: block;
  }
  
  .sidebar {
    width: 280px;
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: none;
  }
  
  .sidebar.open {
    transform: translateX(0);
    box-shadow: 4px 0 24px rgba(0, 0, 0, 0.2);
  }
  
  /* 侧边栏打开时恢复完整显示 */
  .sidebar.open .logo-text,
  .sidebar.open .nav-section-title,
  .sidebar.open .nav-label,
  .sidebar.open .sidebar-footer {
    display: flex;
  }
  .sidebar.open .nav-section-title {
    display: block;
  }
  
  .sidebar.open .sidebar-logo {
    justify-content: flex-start;
    padding: 24px 20px;
  }
  
  .sidebar.open .nav-item {
    justify-content: flex-start;
    padding: 12px 16px;
  }
  
  .main-wrapper {
    margin-left: 0;
  }
  
  .page-content {
    padding: 16px;
  }
  
  .topbar {
    padding: 0 16px;
  }
  
  .topbar-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .page-title {
    font-size: 17px;
  }
  
  .quick-stats {
    display: none;
  }
  
  .action-btn span {
    display: none;
  }
  
  .action-btn {
    padding: 8px;
  }
}
</style>