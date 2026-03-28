<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>收件人分组</h2>
        <p class="page-desc">管理收件人分组，便于批量发送</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="exportRecipients">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          导出
        </button>
        <button class="btn-primary" @click="showGroupModal = true">+ 新建分组</button>
      </div>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead><tr><th>分组名称</th><th>描述</th><th>收件人数</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="g in groups" :key="g.id" :class="{ active: currentGroup === g.id }">
            <td class="cell-main">{{ g.name }}</td>
            <td>{{ g.description || '-' }}</td>
            <td><span class="badge">{{ g.count || 0 }}</span></td>
            <td>
              <div class="action-btns">
                <button class="btn-action" @click="viewRecipients(g.id)">查看</button>
                <button class="btn-action danger" @click="deleteGroup(g.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="groups.length === 0">
            <td colspan="4" class="empty-cell"><div class="empty-state">暂无分组</div></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 收件人列表 -->
    <div v-if="currentGroup" class="card" style="margin-top:16px">
      <div class="card-header">
        <span>收件人列表</span>
        <div>
          <button class="btn-secondary" style="margin-right:8px" @click="showImportModal = true">批量导入</button>
          <button class="btn-primary" @click="showRecipientModal = true">+ 添加</button>
        </div>
      </div>
      <table class="data-table">
        <thead><tr><th>邮箱</th><th>名称</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="r in recipients" :key="r.id">
            <td>{{ r.email }}</td>
            <td>{{ r.name || '-' }}</td>
            <td><span :class="['badge', r.status]">{{ r.status === 'active' ? '正常' : '黑名单' }}</span></td>
            <td><button class="btn-action danger" @click="deleteRecipient(r.id)">删除</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 分组弹窗 -->
    <div v-if="showGroupModal" class="modal-overlay" @click.self="showGroupModal = false">
      <div class="modal">
        <div class="modal-header"><h3>新建分组</h3><button class="modal-close" @click="showGroupModal = false">×</button></div>
        <form @submit.prevent="createGroup">
          <div class="field"><label>名称 *</label><input v-model="groupForm.name" required /></div>
          <div class="field"><label>描述</label><input v-model="groupForm.description" /></div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="showGroupModal = false">取消</button><button type="submit" class="btn-primary">创建</button></div>
        </form>
      </div>
    </div>
    
    <!-- 收件人弹窗 -->
    <div v-if="showRecipientModal" class="modal-overlay" @click.self="showRecipientModal = false">
      <div class="modal">
        <div class="modal-header"><h3>添加收件人</h3><button class="modal-close" @click="showRecipientModal = false">×</button></div>
        <form @submit.prevent="createRecipient">
          <div class="field"><label>邮箱 *</label><input v-model="recipientForm.email" type="email" required /></div>
          <div class="field"><label>名称</label><input v-model="recipientForm.name" /></div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="showRecipientModal = false">取消</button><button type="submit" class="btn-primary">添加</button></div>
        </form>
      </div>
    </div>
    
    <!-- 批量导入弹窗 -->
    <div v-if="showImportModal" class="modal-overlay" @click.self="showImportModal = false">
      <div class="modal">
        <div class="modal-header"><h3>批量导入</h3><button class="modal-close" @click="showImportModal = false">×</button></div>
        <form @submit.prevent="batchImport">
          <div class="field"><label>邮箱列表（每行一个）</label><textarea v-model="batchEmails" rows="6" placeholder="user1@example.com&#10;user2@example.com"></textarea></div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="showImportModal = false">取消</button><button type="submit" class="btn-primary">导入</button></div>
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
  name: 'Recipients',
  setup() {
    const showGroupModal = ref(false)
    const showRecipientModal = ref(false)
    const showImportModal = ref(false)
    const currentGroup = ref('')
    const recipients = ref([])
    const groupForm = ref({ name: '', description: '' })
    const recipientForm = ref({ email: '', name: '' })
    const batchEmails = ref('')
    const groups = computed(() => store.recipientGroups)
    
    const createGroup = async () => {
      try {
        await axios.post(`${API}/recipient-groups`, groupForm.value, { headers: actions.getHeaders() })
        showGroupModal.value = false
        groupForm.value = { name: '', description: '' }
        actions.loadRecipientGroups()
        actions.showToast('创建成功')
      } catch (e) { actions.showToast('创建失败', 'error') }
    }
    
    const deleteGroup = async (id) => {
      if (!confirm('确定删除？分组内收件人也会删除。')) return
      try {
        await axios.delete(`${API}/recipient-groups/${id}`, { headers: actions.getHeaders() })
        actions.loadRecipientGroups()
        if (currentGroup.value === id) { currentGroup.value = ''; recipients.value = [] }
        actions.showToast('已删除')
      } catch (e) {}
    }
    
    const viewRecipients = async (id) => {
      currentGroup.value = id
      const res = await axios.get(`${API}/recipients?group_id=${id}`, { headers: actions.getHeaders() })
      recipients.value = res.data || []
    }
    
    const createRecipient = async () => {
      try {
        await axios.post(`${API}/recipients`, { ...recipientForm.value, group_id: currentGroup.value }, { headers: actions.getHeaders() })
        showRecipientModal.value = false
        recipientForm.value = { email: '', name: '' }
        viewRecipients(currentGroup.value)
        actions.showToast('添加成功')
      } catch (e) { actions.showToast('添加失败', 'error') }
    }
    
    const deleteRecipient = async (id) => {
      try {
        await axios.delete(`${API}/recipients/${id}`, { headers: actions.getHeaders() })
        viewRecipients(currentGroup.value)
      } catch (e) {}
    }
    
    const batchImport = async () => {
      try {
        await axios.post(`${API}/recipients/batch`, { group_id: currentGroup.value, emails: batchEmails.value }, { headers: actions.getHeaders() })
        showImportModal.value = false
        batchEmails.value = ''
        viewRecipients(currentGroup.value)
        actions.showToast('导入成功')
      } catch (e) { actions.showToast('导入失败', 'error') }
    }
    
    const exportRecipients = async () => {
      try {
        const response = await fetch('/api/v1/export/recipients', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        const blob = await response.blob()
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `recipients_${new Date().toISOString().slice(0,10)}.csv`
        a.click()
        window.URL.revokeObjectURL(url)
        actions.showToast('导出成功')
      } catch (e) {
        actions.showToast('导出失败', 'error')
      }
    }
    
    onMounted(() => actions.loadRecipientGroups())
    
    return { showGroupModal, showRecipientModal, showImportModal, currentGroup, recipients, groupForm, recipientForm, batchEmails, groups, createGroup, deleteGroup, viewRecipients, createRecipient, deleteRecipient, batchImport, exportRecipients }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.header-actions {
  display: flex;
  gap: 8px;
}

.card-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #f1f5f9; }
tr.active { background: #eff6ff; }
.badge.active { background: #dcfce7; color: #166534; }
.badge.blacklisted { background: #fee2e2; color: #991b1b; }
</style>