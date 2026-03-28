<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>API Key</h2>
        <p class="page-desc">用于第三方服务调用发信接口的密钥管理</p>
      </div>
      <button class="btn-primary" @click="showModal = true">+ 创建 Key</button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>Key 前缀</th>
            <th>最后使用</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="key in keys" :key="key.id">
            <td class="cell-main">{{ key.name }}</td>
            <td><code class="code-tag">{{ key.key_prefix }}••••••••</code></td>
            <td>{{ key.last_used_at ? formatDate(key.last_used_at) : '从未使用' }}</td>
            <td>{{ formatDate(key.created_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn-action" @click="resetKey(key.id)">重置</button>
                <button class="btn-action danger" @click="deleteKey(key.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="keys.length === 0">
            <td colspan="5" class="empty-cell"><div class="empty-state">暂无 API Key</div></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 创建弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>创建 API Key</h3>
          <button class="modal-close" @click="showModal = false">×</button>
        </div>
        <form @submit.prevent="createKey">
          <div class="field">
            <label>Key 名称 *</label>
            <input v-model="newKeyName" placeholder="如：my-app、production" required />
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showModal = false">取消</button>
            <button type="submit" class="btn-primary">创建</button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- 新Key显示 -->
    <div v-if="newKey" class="modal-overlay" @click.self="newKey = null">
      <div class="modal">
        <div class="modal-header">
          <h3>API Key 已创建</h3>
          <button class="modal-close" @click="newKey = null">×</button>
        </div>
        <div class="key-display">
          <p class="key-warning">⚠️ 请保存此 Key，它只会显示一次！</p>
          <code class="key-value">{{ newKey }}</code>
          <button class="btn-primary" @click="copyKey" style="margin-top:16px">复制</button>
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
  name: 'ApiKeys',
  setup() {
    const showModal = ref(false)
    const newKeyName = ref('')
    const newKey = ref('')
    const keys = computed(() => store.apiKeys)
    
    const createKey = async () => {
      try {
        const res = await axios.post(`${API}/api-keys`, { name: newKeyName.value }, { headers: actions.getHeaders() })
        newKey.value = res.data.key
        showModal.value = false
        newKeyName.value = ''
        actions.loadApiKeys()
      } catch (e) {
        actions.showToast('创建失败', 'error')
      }
    }
    
    const resetKey = async (id) => {
      if (!confirm('确定重置此 Key？旧 Key 将失效。')) return
      try {
        const res = await axios.post(`${API}/api-keys/${id}/reset`, {}, { headers: actions.getHeaders() })
        newKey.value = res.data.key
        actions.loadApiKeys()
      } catch (e) {
        actions.showToast('重置失败', 'error')
      }
    }
    
    const deleteKey = async (id) => {
      if (!confirm('确定删除此 Key？')) return
      try {
        await axios.delete(`${API}/api-keys/${id}`, { headers: actions.getHeaders() })
        actions.loadApiKeys()
        actions.showToast('已删除')
      } catch (e) {
        actions.showToast('删除失败', 'error')
      }
    }
    
    const copyKey = () => {
      navigator.clipboard.writeText(newKey.value)
      actions.showToast('已复制')
    }
    
    const formatDate = (date) => {
      if (!date) return '-'
      return new Date(date).toLocaleString('zh-CN')
    }
    
    onMounted(() => actions.loadApiKeys())
    
    return { showModal, newKeyName, newKey, keys, createKey, resetKey, deleteKey, copyKey, formatDate }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.code-tag {
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
}

.key-display {
  padding: 20px;
  text-align: center;
}

.key-warning {
  color: #f59e0b;
  margin-bottom: 16px;
}

.key-value {
  display: block;
  background: #f1f5f9;
  padding: 12px;
  border-radius: 8px;
  font-family: monospace;
  word-break: break-all;
}
</style>