<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>系统设置</h2>
        <p class="page-desc">版本信息、系统更新与配置管理</p>
      </div>
    </div>
    
    <!-- 版本信息 -->
    <div class="card mb-16">
      <div class="card-header">
        <h3>版本信息</h3>
      </div>
      <div class="version-info">
        <div class="version-item">
          <span class="version-label">当前版本</span>
          <span class="version-value">{{ currentVersion || '加载中...' }}</span>
        </div>
        <div class="version-item">
          <span class="version-label">最新版本</span>
          <span v-if="latestVersion" class="version-value">
            {{ latestVersion }}
            <span v-if="updateStatus === 'available'" class="update-badge new">有新版本</span>
            <span v-else-if="updateStatus === 'latest'" class="update-badge latest">已是最新</span>
          </span>
          <span v-else class="version-value text-muted">-</span>
        </div>
      </div>
    </div>
    
    <!-- 账户设置 -->
    <div class="card mb-16">
      <div class="card-header">
        <h3>账户设置</h3>
      </div>
      <div class="settings-list">
        <div class="settings-item" @click="showChangePwd = true">
          <div class="settings-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <rect x="3" y="11" width="18" height="11" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M7 11V7a5 5 0 0110 0v4" stroke="currentColor" stroke-width="1.5"/>
            </svg>
          </div>
          <div class="settings-content">
            <div class="settings-title">修改密码</div>
            <div class="settings-desc">更改登录密码</div>
          </div>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" class="settings-arrow">
            <path d="M9 18l6-6-6-6" stroke="#94a3b8" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        
        <div class="settings-item">
          <div class="settings-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <path d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" stroke="currentColor" stroke-width="1.5"/>
            </svg>
          </div>
          <div class="settings-content">
            <div class="settings-title">语言设置</div>
            <div class="settings-desc">界面显示语言</div>
          </div>
          <select v-model="locale" class="locale-select" @change="changeLocale">
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
          </select>
        </div>
      </div>
    </div>
    
    <!-- 检测更新 -->
    <div class="card mb-16">
      <div class="card-header">
        <h3>系统更新</h3>
        <button class="btn-secondary" @click="checkUpdate" :disabled="checking">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2"/>
          </svg>
          {{ checking ? '检测中...' : '检测更新' }}
        </button>
      </div>
      
      <div v-if="updateStatus === 'available'" class="update-available">
        <div class="update-icon">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12" stroke="#22c55e" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="update-content">
          <h4>发现新版本 {{ latestVersion }}</h4>
          <p>建议更新以获取最新功能和安全修复</p>
        </div>
        <button class="btn-primary" @click="doUpdate" :disabled="updating">
          {{ updating ? '更新中...' : '立即更新' }}
        </button>
      </div>
      
      <div v-else-if="updateStatus === 'latest'" class="update-latest">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M20 6L9 17l-5-5" stroke="#22c55e" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>当前已是最新版本</span>
      </div>
      
      <div v-else-if="updateStatus === 'error'" class="update-error">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="#ef4444" stroke-width="2"/>
          <path d="M12 8v4M12 16h.01" stroke="#ef4444" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>检测失败，请稍后重试</span>
      </div>
    </div>
    
    <!-- 更新进度弹窗 -->
    <div v-if="updateProgress" class="modal-overlay">
      <div class="modal update-modal">
        <div v-if="updateProgress === 'updating'" class="update-progress">
          <div class="spinner-large"></div>
          <h3>正在更新...</h3>
          <p class="text-muted">请勿关闭页面</p>
          <div class="update-steps">
            <div :class="['step', { done: updateStep >= 1 }]">1. 拉取代码</div>
            <div :class="['step', { done: updateStep >= 2 }]">2. 编译程序</div>
            <div :class="['step', { done: updateStep >= 3 }]">3. 重启服务</div>
          </div>
        </div>
        
        <div v-else-if="updateProgress === 'done'" class="update-result success">
          <div class="result-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" stroke="#22c55e" stroke-width="2"/>
              <path d="M8 12l3 3 5-5" stroke="#22c55e" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <h3>更新完成！</h3>
          <p>页面将在 3 秒后自动刷新</p>
        </div>
        
        <div v-else-if="updateProgress === 'error'" class="update-result error">
          <div class="result-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" stroke="#ef4444" stroke-width="2"/>
              <path d="M15 9l-6 6M9 9l6 6" stroke="#ef4444" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <h3>更新失败</h3>
          <button class="btn-secondary" @click="updateProgress = ''">关闭</button>
        </div>
      </div>
    </div>
    
    <!-- 系统状态 -->
    <div class="card">
      <div class="card-header">
        <h3>系统状态</h3>
        <button class="btn-secondary" @click="loadQueueStats">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          刷新
        </button>
      </div>
      <div class="system-stats">
        <div class="stat-row">
          <span class="stat-name">发送队列</span>
          <span class="stat-values">
            <span class="queue-stat pending">待发送 {{ queueStats.pending || 0 }}</span>
            <span class="queue-stat processing">处理中 {{ queueStats.processing || 0 }}</span>
            <span class="queue-stat sent">已发送 {{ queueStats.sent || 0 }}</span>
            <span class="queue-stat failed">失败 {{ queueStats.failed || 0 }}</span>
          </span>
        </div>
      </div>
    </div>
    
    <!-- 修改密码弹窗 -->
    <ChangePasswordModal v-if="showChangePwd" @close="showChangePwd = false" />
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { store, actions } from '@/store'
import axios from 'axios'
import ChangePasswordModal from '@/components/ChangePasswordModal.vue'

const API = '/api/v1'

export default {
  name: 'Settings',
  components: { ChangePasswordModal },
  setup() {
    const currentVersion = ref('')
    const latestVersion = ref('')
    const updateStatus = ref('') // '', 'checking', 'latest', 'available', 'error'
    const checking = ref(false)
    const updating = ref(false)
    const updateProgress = ref('') // '', 'updating', 'done', 'error'
    const updateStep = ref(0)
    const showChangePwd = ref(false)
    const locale = ref(store.locale)
    
    const queueStats = computed(() => store.queueStats)
    
    const changeLocale = () => {
      actions.setLocale(locale.value)
      actions.showToast('语言已切换 / Language changed')
    }
    
    const loadVersion = async () => {
      try {
        const res = await axios.get(`${API}/version`)
        currentVersion.value = res.data.version
      } catch (e) {}
    }
    
    const checkUpdate = async () => {
      checking.value = true
      updateStatus.value = ''
      try {
        const res = await axios.get('https://api.github.com/repos/DoBestone/smtp-lite/releases/latest')
        latestVersion.value = res.data.tag_name
        updateStatus.value = latestVersion.value === currentVersion.value ? 'latest' : 'available'
      } catch (e) {
        updateStatus.value = 'error'
      } finally {
        checking.value = false
      }
    }
    
    const doUpdate = async () => {
      updating.value = true
      updateProgress.value = 'updating'
      updateStep.value = 1
      
      try {
        await axios.post(`${API}/system/update`, {}, { headers: actions.getHeaders() })
        updateStep.value = 2
        
        // 轮询等待新版本
        const target = latestVersion.value
        const startTime = Date.now()
        const timeout = 120000 // 2分钟超时
        
        const poll = async () => {
          if (Date.now() - startTime > timeout) {
            updateProgress.value = 'error'
            return
          }
          
          try {
            const res = await axios.get(`${API}/version`)
            if (res.data.version === target) {
              updateStep.value = 3
              updateProgress.value = 'done'
              setTimeout(() => window.location.reload(), 3000)
              return
            }
          } catch (e) {}
          
          setTimeout(poll, 2000)
        }
        
        poll()
      } catch (e) {
        updateProgress.value = 'error'
      } finally {
        updating.value = false
      }
    }
    
    const loadQueueStats = () => actions.loadQueueStats()
    
    onMounted(() => {
      loadVersion()
      loadQueueStats()
    })
    
    return {
      currentVersion,
      latestVersion,
      updateStatus,
      checking,
      updating,
      updateProgress,
      updateStep,
      queueStats,
      showChangePwd,
      locale,
      checkUpdate,
      doUpdate,
      loadQueueStats,
      changeLocale
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.mb-16 { margin-bottom: 16px; }

.version-info {
  padding: 20px;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f1f5f9;
}

.version-item:last-child {
  border-bottom: none;
}

.version-label {
  color: #64748b;
  font-size: 14px;
}

.version-value {
  font-weight: 600;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.update-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.update-badge.new {
  background: #dcfce7;
  color: #166534;
}

.update-badge.latest {
  background: #f1f5f9;
  color: #64748b;
}

/* 设置列表 */
.settings-list {
  padding: 8px 0;
}

.settings-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  cursor: pointer;
  transition: background 0.15s;
}

.settings-item:hover {
  background: #f8fafc;
}

.settings-icon {
  width: 40px;
  height: 40px;
  background: #f1f5f9;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
}

.settings-item:hover .settings-icon {
  background: #e2e8f0;
  color: #475569;
}

.settings-content {
  flex: 1;
}

.settings-title {
  font-size: 15px;
  font-weight: 500;
  color: #1e293b;
}

.settings-desc {
  font-size: 13px;
  color: #94a3b8;
  margin-top: 2px;
}

.settings-arrow {
  flex-shrink: 0;
}

.locale-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: #fff;
  cursor: pointer;
  outline: none;
}

.locale-select:focus {
  border-color: #3b82f6;
}

/* 更新区域 */
.update-available {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  border-radius: 0 0 12px 12px;
}

.update-icon {
  flex-shrink: 0;
}

.update-content {
  flex: 1;
}

.update-content h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  color: #166534;
}

.update-content p {
  margin: 0;
  font-size: 13px;
  color: #16a34a;
}

.update-latest,
.update-error {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 0 0 12px 12px;
}

/* 更新进度弹窗 */
.update-modal {
  max-width: 400px;
  text-align: center;
}

.update-progress {
  padding: 40px 24px;
}

.spinner-large {
  width: 48px;
  height: 48px;
  border: 3px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.update-progress h3 {
  margin: 0 0 8px;
  color: #1e293b;
}

.update-steps {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.step {
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 12px;
  background: #f1f5f9;
  color: #94a3b8;
}

.step.done {
  background: #dbeafe;
  color: #1e40af;
}

.update-result {
  padding: 40px 24px;
}

.result-icon {
  margin-bottom: 16px;
}

.update-result h3 {
  margin: 0 0 8px;
}

.update-result.success h3 { color: #166534; }
.update-result.error h3 { color: #991b1b; }

/* 系统状态 */
.system-stats {
  padding: 16px 20px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.stat-name {
  color: #64748b;
  font-size: 14px;
}

.stat-values {
  display: flex;
  gap: 16px;
}

.queue-stat {
  font-size: 13px;
  font-weight: 500;
}

.queue-stat.pending { color: #f59e0b; }
.queue-stat.processing { color: #3b82f6; }
.queue-stat.sent { color: #22c55e; }
.queue-stat.failed { color: #ef4444; }
</style>