<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>邮件模板</h2>
        <p class="page-desc">保存常用邮件模板，发送时快速选用</p>
      </div>
      <button class="btn-primary" @click="openModal()">+ 新建模板</button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead>
          <tr><th>名称</th><th>主题</th><th>类型</th><th>创建时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="t in templates" :key="t.id">
            <td class="cell-main">{{ t.name }}</td>
            <td>{{ t.subject || '-' }}</td>
            <td><span class="badge">{{ t.is_html ? 'HTML' : '纯文本' }}</span></td>
            <td>{{ formatDate(t.created_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn-action" @click="openModal(t)">编辑</button>
                <button class="btn-action danger" @click="deleteTemplate(t.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="templates.length === 0">
            <td colspan="5" class="empty-cell"><div class="empty-state">暂无模板</div></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal" style="max-width:600px">
        <div class="modal-header">
          <h3>{{ editing ? '编辑模板' : '新建模板' }}</h3>
          <button class="modal-close" @click="showModal = false">×</button>
        </div>
        <form @submit.prevent="saveTemplate">
          <div class="field"><label>名称 *</label><input v-model="form.name" required /></div>
          <div class="field"><label>主题</label><input v-model="form.subject" /></div>
          <div class="field"><label>内容 *</label><textarea v-model="form.body" rows="6" required></textarea></div>
          <div class="field">
            <label class="checkbox"><input type="checkbox" v-model="form.is_html" /> HTML 格式</label>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn-secondary" @click="showModal = false">取消</button>
            <button type="submit" class="btn-primary">保存</button>
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
  name: 'Templates',
  setup() {
    const showModal = ref(false)
    const editing = ref(null)
    const form = ref({ name: '', subject: '', body: '', is_html: true })
    const templates = computed(() => store.templates)
    
    const openModal = (t = null) => {
      editing.value = t
      form.value = t ? { name: t.name, subject: t.subject, body: t.body, is_html: t.is_html } : { name: '', subject: '', body: '', is_html: true }
      showModal.value = true
    }
    
    const saveTemplate = async () => {
      try {
        if (editing.value) {
          await axios.put(`${API}/templates/${editing.value.id}`, form.value, { headers: actions.getHeaders() })
        } else {
          await axios.post(`${API}/templates`, form.value, { headers: actions.getHeaders() })
        }
        showModal.value = false
        actions.loadTemplates()
        actions.showToast('保存成功')
      } catch (e) {
        actions.showToast('保存失败', 'error')
      }
    }
    
    const deleteTemplate = async (id) => {
      if (!confirm('确定删除？')) return
      try {
        await axios.delete(`${API}/templates/${id}`, { headers: actions.getHeaders() })
        actions.loadTemplates()
        actions.showToast('已删除')
      } catch (e) {}
    }
    
    const formatDate = (d) => d ? new Date(d).toLocaleString('zh-CN') : '-'
    
    onMounted(() => actions.loadTemplates())
    
    return { showModal, editing, form, templates, openModal, saveTemplate, deleteTemplate, formatDate }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
.checkbox { display: flex; align-items: center; gap: 8px; cursor: pointer; }
</style>