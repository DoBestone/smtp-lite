<template>
  <div class="app">
    <!-- 登录页 -->
    <div v-if="!isLoggedIn" class="login-container">
      <div class="login-box">
        <h1>📬 SMTP Lite</h1>
        <p class="subtitle">个人邮箱聚合系统</p>
        <form @submit.prevent="login">
          <input v-model="username" placeholder="用户名" type="text" required />
          <input v-model="password" placeholder="密码" type="password" required />
          <button type="submit" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>
        <p v-if="error" class="error">{{ error }}</p>
      </div>
    </div>

    <!-- 主界面 -->
    <div v-else class="main-container">
      <header>
        <h1>📬 SMTP Lite</h1>
        <button @click="logout" class="btn-logout">退出</button>
      </header>

      <nav>
        <button :class="{ active: tab === 'smtp' }" @click="tab = 'smtp'">SMTP 账号</button>
        <button :class="{ active: tab === 'keys' }" @click="tab = 'keys'">API Key</button>
        <button :class="{ active: tab === 'logs' }" @click="tab = 'logs'">发送日志</button>
        <button :class="{ active: tab === 'stats' }" @click="tab = 'stats'">统计</button>
      </nav>

      <main>
        <!-- SMTP 账号管理 -->
        <div v-if="tab === 'smtp'" class="panel">
          <div class="panel-header">
            <h2>SMTP 账号</h2>
            <button @click="showAddSmtp = true" class="btn-primary">+ 添加账号</button>
          </div>
          
          <table>
            <thead>
              <tr>
                <th>邮箱</th>
                <th>SMTP 服务器</th>
                <th>端口</th>
                <th>日限额</th>
                <th>已用</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="acc in smtpAccounts" :key="acc.id">
                <td>{{ acc.email }}</td>
                <td>{{ acc.smtp_host }}</td>
                <td>{{ acc.smtp_port }}</td>
                <td>{{ acc.daily_limit || '无限制' }}</td>
                <td>{{ acc.daily_used }}</td>
                <td>
                  <span :class="['status', acc.status]">{{ acc.status }}</span>
                </td>
                <td class="actions">
                  <button @click="testSmtp(acc.id)" class="btn-sm">测试</button>
                  <button @click="toggleSmtp(acc.id)" class="btn-sm">
                    {{ acc.status === 'active' ? '禁用' : '启用' }}
                  </button>
                  <button @click="deleteSmtp(acc.id)" class="btn-sm btn-danger">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-if="smtpAccounts.length === 0" class="empty">暂无 SMTP 账号</p>
        </div>

        <!-- API Key 管理 -->
        <div v-if="tab === 'keys'" class="panel">
          <div class="panel-header">
            <h2>API Key</h2>
            <button @click="createApiKey" class="btn-primary">+ 创建 Key</button>
          </div>
          
          <table>
            <thead>
              <tr>
                <th>名称</th>
                <th>前缀</th>
                <th>最后使用</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="key in apiKeys" :key="key.id">
                <td>{{ key.name }}</td>
                <td><code>{{ key.key_prefix }}...</code></td>
                <td>{{ key.last_used_at || '从未使用' }}</td>
                <td>{{ formatDate(key.created_at) }}</td>
                <td>
                  <button @click="deleteApiKey(key.id)" class="btn-sm btn-danger">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-if="apiKeys.length === 0" class="empty">暂无 API Key</p>
        </div>

        <!-- 发送日志 -->
        <div v-if="tab === 'logs'" class="panel">
          <div class="panel-header">
            <h2>发送日志</h2>
            <button @click="loadLogs" class="btn-sm">刷新</button>
          </div>
          
          <table>
            <thead>
              <tr>
                <th>收件人</th>
                <th>主题</th>
                <th>状态</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in logs" :key="log.id">
                <td>{{ log.to_email }}</td>
                <td>{{ log.subject || '-' }}</td>
                <td>
                  <span :class="['status', log.status]">{{ log.status }}</span>
                </td>
                <td>{{ formatDate(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <p v-if="logs.length === 0" class="empty">暂无发送记录</p>
        </div>

        <!-- 统计 -->
        <div v-if="tab === 'stats'" class="panel">
          <h2>统计信息</h2>
          <div class="stats-grid">
            <div class="stat-card">
              <div class="stat-value">{{ stats.total_sent || 0 }}</div>
              <div class="stat-label">总发送量</div>
            </div>
            <div class="stat-card success">
              <div class="stat-value">{{ stats.success || 0 }}</div>
              <div class="stat-label">成功</div>
            </div>
            <div class="stat-card failed">
              <div class="stat-value">{{ stats.failed || 0 }}</div>
              <div class="stat-label">失败</div>
            </div>
            <div class="stat-card">
              <div class="stat-value">{{ stats.today_sent || 0 }}</div>
              <div class="stat-label">今日发送</div>
            </div>
            <div class="stat-card">
              <div class="stat-value">{{ (stats.success_rate || 0).toFixed(1) }}%</div>
              <div class="stat-label">成功率</div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- 添加 SMTP 账号弹窗 -->
    <div v-if="showAddSmtp" class="modal-overlay" @click.self="showAddSmtp = false">
      <div class="modal">
        <h3>添加 SMTP 账号</h3>
        <form @submit.prevent="addSmtpAccount">
          <input v-model="newSmtp.email" placeholder="邮箱地址" type="email" required />
          <input v-model="newSmtp.password" placeholder="密码/授权码" type="password" required />
          <input v-model="newSmtp.smtp_host" placeholder="SMTP 服务器 (如 smtp.gmail.com)" required />
          <input v-model.number="newSmtp.smtp_port" placeholder="端口 (默认 587)" type="number" />
          <input v-model.number="newSmtp.daily_limit" placeholder="每日限额 (留空不限制)" type="number" />
          <div class="modal-actions">
            <button type="button" @click="showAddSmtp = false">取消</button>
            <button type="submit" class="btn-primary">添加</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 新建 API Key 弹窗 -->
    <div v-if="newKeyInfo" class="modal-overlay" @click.self="newKeyInfo = null">
      <div class="modal">
        <h3>API Key 已创建</h3>
        <p class="warning">⚠️ 请保存此 Key，它只会显示一次！</p>
        <code class="key-display">{{ newKeyInfo.key }}</code>
        <div class="modal-actions">
          <button @click="copyKey" class="btn-primary">复制</button>
          <button @click="newKeyInfo = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

const API = '/api/v1'

export default {
  data() {
    return {
      isLoggedIn: false,
      username: '',
      password: '',
      loading: false,
      error: '',
      tab: 'smtp',
      smtpAccounts: [],
      apiKeys: [],
      logs: [],
      stats: {},
      showAddSmtp: false,
      newSmtp: {
        email: '',
        password: '',
        smtp_host: '',
        smtp_port: 587,
        daily_limit: 500
      },
      newKeyInfo: null
    }
  },
  mounted() {
    const token = localStorage.getItem('token')
    if (token) {
      this.isLoggedIn = true
      this.loadData()
    }
  },
  methods: {
    async login() {
      this.loading = true
      this.error = ''
      try {
        const res = await axios.post(`${API}/auth/login`, {
          username: this.username,
          password: this.password
        })
        localStorage.setItem('token', res.data.token)
        this.isLoggedIn = true
        this.loadData()
      } catch (e) {
        this.error = e.response?.data?.error || '登录失败'
      } finally {
        this.loading = false
      }
    },
    logout() {
      localStorage.removeItem('token')
      this.isLoggedIn = false
    },
    getHeaders() {
      return { Authorization: `Bearer ${localStorage.getItem('token')}` }
    },
    async loadData() {
      this.loadSmtpAccounts()
      this.loadApiKeys()
      this.loadStats()
    },
    async loadSmtpAccounts() {
      try {
        const res = await axios.get(`${API}/smtp-accounts`, { headers: this.getHeaders() })
        this.smtpAccounts = res.data
      } catch (e) {
        console.error('加载 SMTP 账号失败', e)
      }
    },
    async loadApiKeys() {
      try {
        const res = await axios.get(`${API}/api-keys`, { headers: this.getHeaders() })
        this.apiKeys = res.data
      } catch (e) {
        console.error('加载 API Key 失败', e)
      }
    },
    async loadLogs() {
      try {
        const res = await axios.get(`${API}/logs`, { headers: this.getHeaders() })
        this.logs = res.data.logs || []
      } catch (e) {
        console.error('加载日志失败', e)
      }
    },
    async loadStats() {
      try {
        const res = await axios.get(`${API}/stats`, { headers: this.getHeaders() })
        this.stats = res.data
      } catch (e) {
        console.error('加载统计失败', e)
      }
    },
    async addSmtpAccount() {
      try {
        await axios.post(`${API}/smtp-accounts`, this.newSmtp, { headers: this.getHeaders() })
        this.showAddSmtp = false
        this.newSmtp = { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 }
        this.loadSmtpAccounts()
      } catch (e) {
        alert(e.response?.data?.error || '添加失败')
      }
    },
    async testSmtp(id) {
      try {
        const res = await axios.post(`${API}/smtp-accounts/${id}/test`, {}, { headers: this.getHeaders() })
        alert(res.data.success ? '连接成功！' : '连接失败: ' + res.data.error)
      } catch (e) {
        alert('测试失败: ' + (e.response?.data?.error || e.message))
      }
    },
    async toggleSmtp(id) {
      try {
        await axios.post(`${API}/smtp-accounts/${id}/toggle`, {}, { headers: this.getHeaders() })
        this.loadSmtpAccounts()
      } catch (e) {
        alert('操作失败')
      }
    },
    async deleteSmtp(id) {
      if (!confirm('确定删除此账号？')) return
      try {
        await axios.delete(`${API}/smtp-accounts/${id}`, { headers: this.getHeaders() })
        this.loadSmtpAccounts()
      } catch (e) {
        alert('删除失败')
      }
    },
    async createApiKey() {
      const name = prompt('请输入 Key 名称：')
      if (!name) return
      try {
        const res = await axios.post(`${API}/api-keys`, { name }, { headers: this.getHeaders() })
        this.newKeyInfo = res.data
        this.loadApiKeys()
      } catch (e) {
        alert('创建失败')
      }
    },
    async deleteApiKey(id) {
      if (!confirm('确定删除此 Key？')) return
      try {
        await axios.delete(`${API}/api-keys/${id}`, { headers: this.getHeaders() })
        this.loadApiKeys()
      } catch (e) {
        alert('删除失败')
      }
    },
    copyKey() {
      navigator.clipboard.writeText(this.newKeyInfo.key)
      alert('已复制到剪贴板')
    },
    formatDate(date) {
      if (!date) return '-'
      return new Date(date).toLocaleString('zh-CN')
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f7fa;
  color: #333;
}

.app {
  min-height: 100vh;
}

/* 登录页 */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  background: white;
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.2);
  text-align: center;
  width: 100%;
  max-width: 400px;
}

.login-box h1 {
  margin-bottom: 8px;
}

.subtitle {
  color: #666;
  margin-bottom: 24px;
}

.login-box input {
  width: 100%;
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
}

.login-box button {
  width: 100%;
  padding: 12px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
}

.login-box button:disabled {
  background: #ccc;
}

.error {
  color: #e74c3c;
  margin-top: 12px;
}

/* 主界面 */
.main-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

header h1 {
  font-size: 24px;
}

.btn-logout {
  padding: 8px 16px;
  background: #eee;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

nav {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}

nav button {
  padding: 10px 20px;
  border: none;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

nav button.active {
  background: #667eea;
  color: white;
}

.panel {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

h2 {
  font-size: 18px;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

th {
  font-weight: 600;
  color: #666;
  font-size: 13px;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  text-transform: uppercase;
}

.status.active, .status.success {
  background: #d4edda;
  color: #155724;
}

.status.disabled, .status.failed {
  background: #f8d7da;
  color: #721c24;
}

.actions {
  display: flex;
  gap: 4px;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.btn-danger {
  color: #dc3545;
  border-color: #dc3545;
}

.btn-primary {
  background: #667eea;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
}

.empty {
  text-align: center;
  color: #999;
  padding: 40px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
  margin-top: 20px;
}

.stat-card {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
}

.stat-label {
  color: #666;
  margin-top: 4px;
}

.stat-card.success .stat-value { color: #28a745; }
.stat-card.failed .stat-value { color: #dc3545; }

/* 弹窗 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  background: white;
  padding: 24px;
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
}

.modal h3 {
  margin-bottom: 16px;
}

.modal input {
  width: 100%;
  padding: 10px;
  margin-bottom: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.warning {
  color: #856404;
  background: #fff3cd;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.key-display {
  display: block;
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-size: 14px;
  word-break: break-all;
}
</style>