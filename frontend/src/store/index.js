import { reactive } from 'vue'
import axios from 'axios'

const API = '/api/v1'

// 语言包
const translations = {
  'zh-CN': {
    nav: {
      smtp: 'SMTP 账号',
      send: '发送邮件',
      keys: 'API Key',
      templates: '邮件模板',
      recipients: '收件人',
      logs: '发送日志',
      stats: '统计数据',
      webhooks: 'Webhook',
      blacklist: '黑名单',
      docs: 'API 文档',
      settings: '系统设置'
    },
    common: {
      add: '添加',
      edit: '编辑',
      delete: '删除',
      save: '保存',
      cancel: '取消',
      confirm: '确认',
      success: '成功',
      failed: '失败',
      loading: '加载中...'
    }
  },
  'en-US': {
    nav: {
      smtp: 'SMTP Accounts',
      send: 'Send Email',
      keys: 'API Keys',
      templates: 'Templates',
      recipients: 'Recipients',
      logs: 'Send Logs',
      stats: 'Statistics',
      webhooks: 'Webhooks',
      blacklist: 'Blacklist',
      docs: 'API Docs',
      settings: 'Settings'
    },
    common: {
      add: 'Add',
      edit: 'Edit',
      delete: 'Delete',
      save: 'Save',
      cancel: 'Cancel',
      confirm: 'Confirm',
      success: 'Success',
      failed: 'Failed',
      loading: 'Loading...'
    }
  }
}

// 创建一个简单的全局状态
export const store = reactive({
  // 用户状态
  isLoggedIn: !!localStorage.getItem('token'),
  username: '',
  locale: localStorage.getItem('locale') || 'zh-CN',
  
  // 数据
  smtpAccounts: [],
  apiKeys: [],
  templates: [],
  recipientGroups: [],
  webhooks: [],
  blacklist: [],
  logs: [],
  stats: {},
  queueStats: {},
  
  // UI 状态
  loading: false,
  toast: { show: false, msg: '', type: 'success' }
})

// 方法
export const actions = {
  getHeaders() {
    return { Authorization: `Bearer ${localStorage.getItem('token')}` }
  },
  
  showToast(msg, type = 'success') {
    store.toast = { show: true, msg, type }
    setTimeout(() => { store.toast.show = false }, 3000)
  },
  
  // 多语言
  t(key) {
    const parts = key.split('.')
    let value = translations[store.locale]
    for (const part of parts) {
      value = value?.[part]
    }
    return value || key
  },
  
  setLocale(locale) {
    store.locale = locale
    localStorage.setItem('locale', locale)
  },
  
  async login(username, password) {
    const res = await axios.post(`${API}/auth/login`, { username, password })
    localStorage.setItem('token', res.data.token)
    store.isLoggedIn = true
    store.username = username
    return res.data
  },
  
  logout() {
    localStorage.removeItem('token')
    store.isLoggedIn = false
    store.username = ''
  },
  
  async loadSmtpAccounts() {
    const res = await axios.get(`${API}/smtp-accounts`, { headers: this.getHeaders() })
    store.smtpAccounts = res.data || []
  },
  
  async loadApiKeys() {
    const res = await axios.get(`${API}/api-keys`, { headers: this.getHeaders() })
    store.apiKeys = res.data || []
  },
  
  async loadTemplates() {
    const res = await axios.get(`${API}/templates`, { headers: this.getHeaders() })
    store.templates = res.data || []
  },
  
  async loadRecipientGroups() {
    const res = await axios.get(`${API}/recipient-groups`, { headers: this.getHeaders() })
    store.recipientGroups = res.data || []
  },
  
  async loadWebhooks() {
    const res = await axios.get(`${API}/webhooks`, { headers: this.getHeaders() })
    store.webhooks = res.data || []
  },
  
  async loadBlacklist() {
    const res = await axios.get(`${API}/blacklist`, { headers: this.getHeaders() })
    store.blacklist = res.data || []
  },
  
  async loadLogs(page = 1, pageSize = 50) {
    const res = await axios.get(`${API}/send/logs?page=${page}&page_size=${pageSize}`, { headers: this.getHeaders() })
    store.logs = res.data.logs || []
    return res.data
  },
  
  async loadStats() {
    const res = await axios.get(`${API}/stats`, { headers: this.getHeaders() })
    store.stats = res.data || {}
  },
  
  async loadQueueStats() {
    const res = await axios.get(`${API}/queue/stats`, { headers: this.getHeaders() })
    store.queueStats = res.data || {}
  },
  
  async loadAll() {
    await Promise.all([
      this.loadSmtpAccounts(),
      this.loadApiKeys(),
      this.loadTemplates(),
      this.loadRecipientGroups(),
      this.loadWebhooks(),
      this.loadBlacklist(),
      this.loadStats(),
      this.loadQueueStats()
    ])
  }
}