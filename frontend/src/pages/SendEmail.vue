<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>发送邮件</h2>
        <p class="page-desc">在线发送邮件，支持单发、批量发送、定时发送</p>
      </div>
      <select v-model="sendMode" class="form-select" @change="resetForm">
        <option value="single">单封发送</option>
        <option value="batch">批量发送</option>
        <option value="scheduled">定时发送</option>
      </select>
    </div>
    
    <div class="send-container">
      <!-- 主表单 -->
      <div class="card send-form-card">
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
            <p class="field-hint">每行一个邮箱，或从收件人分组导入</p>
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
            <textarea v-model="form.body" rows="10" required></textarea>
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
          
          <!-- 附件上传 -->
          <div class="field">
            <label>附件</label>
            <div class="attachment-area" @click="$refs.fileInput.click()" @dragover.prevent @drop.prevent="handleDrop">
              <input ref="fileInput" type="file" multiple @change="handleFileSelect" style="display:none" />
              <div class="attachment-placeholder">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12" stroke="#94a3b8" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
                <span>点击或拖拽文件到此处上传</span>
              </div>
            </div>
            <div v-if="attachments.length > 0" class="attachment-list">
              <div v-for="(att, idx) in attachments" :key="idx" class="attachment-item">
                <span class="attachment-name">{{ att.filename }}</span>
                <span class="attachment-size">{{ formatSize(att.size) }}</span>
                <button type="button" class="attachment-remove" @click="removeAttachment(idx)">×</button>
              </div>
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
              {{ loading ? '发送中...' : buttonText }}
            </button>
          </div>
        </form>
      </div>
      
      <!-- 模板选择器 -->
      <div class="card template-card">
        <div class="card-header">
          <h3>快速选择模板</h3>
          <router-link to="/templates" class="text-link">管理模板</router-link>
        </div>
        <div v-if="templates.length > 0" class="template-list">
          <button v-for="t in templates" :key="t.id" class="template-btn" @click="applyTemplate(t)">
            <span class="template-name">{{ t.name }}</span>
            <span class="template-type">{{ t.is_html ? 'HTML' : '文本' }}</span>
          </button>
        </div>
        <div v-else class="template-empty">
          <p>暂无模板</p>
          <router-link to="/templates" class="text-link">创建模板</router-link>
        </div>
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
  name: 'SendEmail',
  setup() {
    const sendMode = ref('single')
    const loading = ref(false)
    const result = ref(null)
    const batchEmails = ref('')
    const scheduledTime = ref('')
    const attachments = ref([])
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
    
    const templates = computed(() => store.templates)
    
    const buttonText = computed(() => {
      if (loading.value) return '发送中...'
      if (sendMode.value === 'single') return '立即发送'
      if (sendMode.value === 'batch') return '批量发送'
      return '定时发送'
    })
    
    // 文件处理
    const handleFileSelect = (e) => {
      const files = Array.from(e.target.files)
      files.forEach(file => addAttachment(file))
      e.target.value = ''
    }
    
    const handleDrop = (e) => {
      const files = Array.from(e.dataTransfer.files)
      files.forEach(file => addAttachment(file))
    }
    
    const addAttachment = (file) => {
      const reader = new FileReader()
      reader.onload = () => {
        attachments.value.push({
          filename: file.name,
          content: reader.result.split(',')[1], // base64
          type: file.type || 'application/octet-stream',
          size: file.size
        })
      }
      reader.readAsDataURL(file)
    }
    
    const removeAttachment = (idx) => {
      attachments.value.splice(idx, 1)
    }
    
    const formatSize = (bytes) => {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }
    
    const applyTemplate = (t) => {
      form.value.subject = t.subject || ''
      form.value.body = t.body || ''
      form.value.is_html = t.is_html || false
      actions.showToast('已应用模板: ' + t.name)
    }
    
    const sendEmail = async () => {
      loading.value = true
      result.value = null
      try {
        const data = {
          ...form.value,
          attachments: attachments.value
        }
        
        if (form.value.cc) {
          data.cc = form.value.cc.split(',').map(e => e.trim()).filter(Boolean)
        }
        if (form.value.bcc) {
          data.bcc = form.value.bcc.split(',').map(e => e.trim()).filter(Boolean)
        }
        
        if (sendMode.value === 'single') {
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
      attachments.value = []
      result.value = null
    }
    
    onMounted(() => {
      if (templates.value.length === 0) {
        actions.loadTemplates()
      }
    })
    
    return {
      sendMode,
      loading,
      result,
      batchEmails,
      scheduledTime,
      attachments,
      form,
      templates,
      buttonText,
      handleFileSelect,
      handleDrop,
      removeAttachment,
      formatSize,
      applyTemplate,
      sendEmail,
      resetForm
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.send-container {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 24px;
}

.send-form-card {
  min-width: 0;
}

.template-card {
  position: sticky;
  top: 88px;
  height: fit-content;
}

.send-form {
  padding: 24px;
}

.field-options {
  display: flex;
  gap: 16px;
  margin-top: 8px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
}

/* 附件上传 */
.attachment-area {
  border: 2px dashed #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.attachment-area:hover {
  border-color: #3b82f6;
  background: #f8fafc;
}

.attachment-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #94a3b8;
  font-size: 14px;
}

.attachment-list {
  margin-top: 12px;
}

.attachment-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  margin-bottom: 8px;
}

.attachment-name {
  flex: 1;
  font-size: 14px;
  color: #334155;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-size {
  font-size: 12px;
  color: #94a3b8;
}

.attachment-remove {
  background: #fee2e2;
  border: none;
  color: #dc2626;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
}

/* 模板列表 */
.template-list {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.template-btn {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.template-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.template-name {
  font-size: 14px;
  font-weight: 500;
  color: #334155;
}

.template-type {
  font-size: 11px;
  padding: 2px 6px;
  background: #e2e8f0;
  border-radius: 4px;
  color: #64748b;
}

.template-empty {
  padding: 24px;
  text-align: center;
  color: #94a3b8;
}

.template-empty p {
  margin: 0 0 8px;
}

.text-link {
  font-size: 13px;
  color: #3b82f6;
  text-decoration: none;
}

.text-link:hover {
  text-decoration: underline;
}

/* 响应式 */
@media (max-width: 1024px) {
  .send-container {
    grid-template-columns: 1fr;
  }
  
  .template-card {
    position: static;
  }
}
</style>