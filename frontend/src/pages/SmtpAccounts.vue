<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>SMTP 账号管理</h2>
        <p class="page-desc">管理用于发送邮件的 SMTP 账号，支持多账号轮询与故障转移</p>
      </div>
      <button class="btn-primary" @click="showModal = true">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        添加账号
      </button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>邮箱</th>
            <th>SMTP 服务器</th>
            <th>端口</th>
            <th>日限额</th>
            <th>已发送</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="acc in accounts" :key="acc.id">
            <td class="cell-main">{{ acc.email }}</td>
            <td>{{ acc.smtp_host }}</td>
            <td>{{ acc.smtp_port }}</td>
            <td>{{ acc.daily_limit || '无限制' }}</td>
            <td>{{ acc.daily_used }} / {{ acc.daily_limit || '∞' }}</td>
            <td>
              <span :class="['badge', acc.status]">{{ acc.status === 'active' ? '正常' : '禁用' }}</span>
            </td>
            <td>
              <div class="action-btns">
                <button class="btn-action" @click="testAccount(acc.id)" :disabled="testing === acc.id">
                  {{ testing === acc.id ? '测试中...' : '测试' }}
                </button>
                <button class="btn-action" @click="toggleAccount(acc.id)">
                  {{ acc.status === 'active' ? '禁用' : '启用' }}
                </button>
                <button class="btn-action danger" @click="deleteAccount(acc.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="accounts.length === 0">
            <td colspan="7" class="empty-cell">
              <div class="empty-state">暂无 SMTP 账号</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 添加账号弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>添加 SMTP 账号</h3>
          <button class="modal-close" @click="showModal = false">×</button>
        </div>
        <form @submit.prevent="createAccount">
          <div class="field">
            <label>邮箱地址 *</label>
            <input v-model="form.email" type="email" required />
          </div>
          <div class="field">
            <label>密码/授权码 *</label>
            <input v-model="form.password" type="password" required />
          </div>
          <div class="field-row">
            <div class="field">
              <label>SMTP 服务器 *</label>
              <input v-model="form.smtp_host" placeholder="smtp.gmail.com" required />
            </div>
            <div class="field">
              <label>端口</label>
              <input v-model.number="form.smtp_port" type="number" />
            </div>
          </div>
          <div class="field">
            <label>每日限额</label>
            <input v-model.number="form.daily_limit" type="number" placeholder="留空不限制" />
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showModal = false">取消</button>
            <button type="submit" class="btn-primary">添加</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { store, actions } from '@/store'
import axios from 'axios'

const API = '/api/v1'

export default {
  name: 'SmtpAccounts',
  setup() {
    const showModal = ref(false)
    const testing = ref(null)
    const form = ref({
      email: '',
      password: '',
      smtp_host: '',
      smtp_port: 587,
      daily_limit: 500
    })
    
    const accounts = computed(() => store.smtpAccounts)
    
    const createAccount = async () => {
      try {
        await axios.post(`${API}/smtp-accounts`, form.value, { headers: actions.getHeaders() })
        showModal.value = false
        form.value = { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 }
        actions.loadSmtpAccounts()
        actions.showToast('账号添加成功')
      } catch (e) {
        actions.showToast(e.response?.data?.error || '添加失败', 'error')
      }
    }
    
    const testAccount = async (id) => {
      testing.value = id
      try {
        const res = await axios.post(`${API}/smtp-accounts/${id}/test`, {}, { headers: actions.getHeaders() })
        actions.showToast(res.data.success ? '连接成功' : '连接失败: ' + res.data.error)
      } catch (e) {
        actions.showToast('测试失败: ' + e.response?.data?.error, 'error')
      } finally {
        testing.value = null
      }
    }
    
    const toggleAccount = async (id) => {
      try {
        await axios.post(`${API}/smtp-accounts/${id}/toggle`, {}, { headers: actions.getHeaders() })
        actions.loadSmtpAccounts()
      } catch (e) {
        actions.showToast('操作失败', 'error')
      }
    }
    
    const deleteAccount = async (id) => {
      if (!confirm('确定删除此账号？')) return
      try {
        await axios.delete(`${API}/smtp-accounts/${id}`, { headers: actions.getHeaders() })
        actions.loadSmtpAccounts()
        actions.showToast('账号已删除')
      } catch (e) {
        actions.showToast('删除失败', 'error')
      }
    }
    
    onMounted(() => {
      if (accounts.value.length === 0) {
        actions.loadSmtpAccounts()
      }
    })
    
    return {
      showModal,
      testing,
      form,
      accounts,
      createAccount,
      testAccount,
      toggleAccount,
      deleteAccount
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
</style>