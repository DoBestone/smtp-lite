<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>发送邮件</h2>
        <p class="page-desc">在线发送邮件，支持单发、批量发送、定时发送</p>
      </div>
      <select v-model="sendMode" class="form-select">
        <option value="single">单封发送</option>
        <option value="batch">批量发送</option>
        <option value="scheduled">定时发送</option>
      </select>
    </div>
    
    <div class="card">
      <form @submit.prevent="sendEmail" class="send-form">
        <!-- 收件人 -->
        <div v-if="sendMode !== 'batch'" class="field">
          <label>收件人 *</label>
          <input v-model="form.to" type="email" placeholder="recipient@example.com" required />
        </div>
        
        <!-- 批量发送 -->
        <div v-if="sendMode === 'batch'" class="field">
          <label>收件人列表 *</label>
          <textarea v-model="batchEmails" rows="5" placeholder="每行一个邮箱地址&#10;user1@example.com&#10;user2@example.com" required></textarea>
        </div>
        
        <!-- 定时发送时间 -->
        <div v-if="sendMode === 'scheduled'" class="field">
          <label>发送时间 *</label>
          <input v-model="scheduledTime" type="datetime-local" required />
        </div>
        
        <div class="field">
          <label>邮件主题 *</label>
          <input v-model="form.subject" required />
        </div>
        
        <div class="field">
          <label>邮件内容 *</label>
          <textarea v-model="form.body" rows="8" required></textarea>
          <div class="field-options">
            <label class="checkbox">
              <input type="checkbox" v-model="form.is_html" />
              HTML 格式
            </label>
            <label v-if="form.is_html" class="checkbox">
              <input type="checkbox" v-model="form.track_enabled" />
              启用追踪
            </label>
          </div>
        </div>
        
        <div class="field-row">
          <div class="field">
            <label>发件人名称</label>
            <input v-model="form.from_name" placeholder="可选" />
          </div>
          <div v-if="sendMode === 'single'" class="field">
            <label>抄送</label>
            <input v-model="form.cc" placeholder="多人用逗号分隔" />
          </div>
        </div>
        
        <div v-if="sendMode === 'single'" class="field">
          <label>密送</label>
          <input v-model="form.bcc" placeholder="多人用逗号分隔" />
        </div>
        
        <!-- 结果提示 -->
        <div v-if="result" :class="['alert', result.success ? 'success' : 'error']">
          {{ result.message }}
        </div>
        
        <div class="form-actions">
          <button type="button" class="btn-secondary" @click="resetForm">清空</button>
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '发送中...' : (sendMode === 'single' ? '立即发送' : sendMode === 'batch' ? '批量发送' : '定时发送') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { actions } from '@/store'
import axios from 'axios'

const API = '/api/v1'

export default {
  name: 'SendEmail',
  setup() {
    const sendMode = ref('single')
    const loading = ref(false)
    const result = ref(null)
    const batchEmails = ref('')
    const scheduledTime = ref('')
    const form = ref({
      to: '',
      subject: '',
      body: '',
      is_html: false,
      from_name: '',
      cc: '',
      bcc: '',
      track_enabled: false
    })
    
    const sendEmail = async () => {
      loading.value = true
      result.value = null
      try {
        if (sendMode.value === 'single') {
          const data = { ...form.value }
          if (form.value.cc) data.cc = form.value.cc.split(',').map(e => e.trim()).filter(Boolean)
          if (form.value.bcc) data.bcc = form.value.bcc.split(',').map(e => e.trim()).filter(Boolean)
          const res = await axios.post(`${API}/send`, data, { headers: actions.getHeaders() })
          result.value = res.data
        } else if (sendMode.value === 'batch') {
          const emails = batchEmails.value.split('\n').map(e => e.trim()).filter(Boolean)
          await axios.post(`${API}/send/batch`, {
            name: `批量发送 ${new Date().toLocaleString()}`,
            emails,
            subject: form.value.subject,
            body: form.value.body,
            is_html: form.value.is_html,
            from_name: form.value.from_name
          }, { headers: actions.getHeaders() })
          result.value = { success: true, message: `已加入队列，共 ${emails.length} 封` }
        } else if (sendMode.value === 'scheduled') {
          await axios.post(`${API}/send/scheduled`, {
            to: form.value.to,
            subject: form.value.subject,
            body: form.value.body,
            is_html: form.value.is_html,
            from_name: form.value.from_name,
            scheduled_at: scheduledTime.value
          }, { headers: actions.getHeaders() })
          result.value = { success: true, message: '已加入定时队列' }
        }
        
        if (result.value?.success) {
          actions.showToast(result.value.message)
          actions.loadStats()
        }
      } catch (e) {
        result.value = { success: false, message: e.response?.data?.error || '发送失败' }
      } finally {
        loading.value = false
      }
    }
    
    const resetForm = () => {
      form.value = { to: '', subject: '', body: '', is_html: false, from_name: '', cc: '', bcc: '', track_enabled: false }
      batchEmails.value = ''
      scheduledTime.value = ''
      result.value = null
    }
    
    return { sendMode, loading, result, batchEmails, scheduledTime, form, sendEmail, resetForm }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.send-form {
  padding: 24px;
}

.form-select {
  padding: 10px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: white;
}

.field-options {
  display: flex;
  gap: 16px;
  margin-top: 8px;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #64748b;
  cursor: pointer;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.alert {
  padding: 12px 16px;
  border-radius: 8px;
  margin-top: 16px;
  font-size: 14px;
}

.alert.success {
  background: #dcfce7;
  color: #166534;
}

.alert.error {
  background: #fee2e2;
  color: #991b1b;
}
</style>