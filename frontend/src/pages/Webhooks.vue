<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>Webhook 回调</h2>
        <p class="page-desc">配置事件回调，邮件发送/打开时自动通知</p>
      </div>
      <button class="btn-primary" @click="showModal = true">+ 新建 Webhook</button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead><tr><th>名称</th><th>URL</th><th>事件</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="w in webhooks" :key="w.id">
            <td class="cell-main">{{ w.name }}</td>
            <td class="text-truncate">{{ w.url }}</td>
            <td><span class="badge">{{ getEventCount(w.events) }} 个</span></td>
            <td><span :class="['badge', w.enabled ? 'active' : 'disabled']">{{ w.enabled ? '启用' : '禁用' }}</span></td>
            <td>
              <div class="action-btns">
                <button class="btn-action" @click="toggleWebhook(w.id)">{{ w.enabled ? '禁用' : '启用' }}</button>
                <button class="btn-action" @click="testWebhook(w.id)">测试</button>
                <button class="btn-action danger" @click="deleteWebhook(w.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="webhooks.length === 0">
            <td colspan="5" class="empty-cell"><div class="empty-state">暂无 Webhook</div></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <div class="modal-header"><h3>新建 Webhook</h3><button class="modal-close" @click="showModal = false">×</button></div>
        <form @submit.prevent="createWebhook">
          <div class="field"><label>名称 *</label><input v-model="form.name" required /></div>
          <div class="field"><label>URL *</label><input v-model="form.url" type="url" required /></div>
          <div class="field"><label>Secret</label><input v-model="form.secret" placeholder="可选" /></div>
          <div class="field"><label>订阅事件</label>
            <div class="event-checks">
              <label v-for="e in eventOptions" :key="e.key" class="checkbox">
                <input type="checkbox" :value="e.key" v-model="form.events" /> {{ e.label }}
              </label>
            </div>
          </div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="showModal = false">取消</button><button type="submit" class="btn-primary">创建</button></div>
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
  name: 'Webhooks',
  setup() {
    const showModal = ref(false)
    const form = ref({ name: '', url: '', secret: '', events: [] })
    const eventOptions = [
      { key: 'send_success', label: '发送成功' },
      { key: 'send_failed', label: '发送失败' },
      { key: 'opened', label: '邮件打开' },
      { key: 'clicked', label: '链接点击' }
    ]
    const webhooks = computed(() => store.webhooks)
    
    const getEventCount = (events) => {
      if (!events) return 0
      try { return JSON.parse(events).length } catch { return 0 }
    }
    
    const createWebhook = async () => {
      try {
        await axios.post(`${API}/webhooks`, form.value, { headers: actions.getHeaders() })
        showModal.value = false
        form.value = { name: '', url: '', secret: '', events: [] }
        actions.loadWebhooks()
        actions.showToast('创建成功')
      } catch (e) { actions.showToast('创建失败', 'error') }
    }
    
    const toggleWebhook = async (id) => {
      try {
        await axios.post(`${API}/webhooks/${id}/toggle`, {}, { headers: actions.getHeaders() })
        actions.loadWebhooks()
      } catch (e) {}
    }
    
    const testWebhook = async (id) => {
      try {
        await axios.post(`${API}/webhooks/${id}/test`, {}, { headers: actions.getHeaders() })
        actions.showToast('测试事件已发送')
      } catch (e) { actions.showToast('发送失败', 'error') }
    }
    
    const deleteWebhook = async (id) => {
      if (!confirm('确定删除？')) return
      try {
        await axios.delete(`${API}/webhooks/${id}`, { headers: actions.getHeaders() })
        actions.loadWebhooks()
        actions.showToast('已删除')
      } catch (e) {}
    }
    
    onMounted(() => actions.loadWebhooks())
    
    return { showModal, form, eventOptions, webhooks, getEventCount, createWebhook, toggleWebhook, testWebhook, deleteWebhook }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
.text-truncate { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.badge.disabled { background: #f1f5f9; color: #64748b; }
.event-checks { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 8px; }
.checkbox { display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 14px; }
</style>