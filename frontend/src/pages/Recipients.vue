<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>收件人分组</h2>
        <p class="page-desc">把分组维护、名单清理和发送入口放到一个工作流里，减少来回跳转。</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="exportRecipients">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          导出
        </button>
        <button class="btn-primary" @click="openCreateGroup">+ 新建分组</button>
      </div>
    </div>

    <div class="overview-grid">
      <div class="summary-card">
        <span class="summary-label">分组总数</span>
        <strong>{{ groups.length }}</strong>
        <small>管理全部收件人池</small>
      </div>
      <div class="summary-card">
        <span class="summary-label">收件人总量</span>
        <strong>{{ totalRecipientCount }}</strong>
        <small>所有分组累计人数</small>
      </div>
      <div class="summary-card">
        <span class="summary-label">当前可发送</span>
        <strong>{{ activeRecipientCount }}</strong>
        <small>{{ selectedGroup ? `${selectedGroup.name} 的正常成员` : '先选择一个分组' }}</small>
      </div>
    </div>

    <div class="recipients-layout">
      <section class="card group-panel">
        <div class="card-header group-panel-head">
          <div>
            <h3>分组目录</h3>
            <p>选择分组后，右侧直接维护成员并跳转群发。</p>
          </div>
          <span class="badge">{{ groups.length }} 组</span>
        </div>

        <div v-if="groups.length" class="group-list">
          <article
            v-for="g in groups"
            :key="g.id"
            class="group-card"
            :class="{ active: currentGroup === g.id }"
            tabindex="0"
            role="button"
            @click="viewRecipients(g.id)"
            @keydown.enter.prevent="viewRecipients(g.id)"
            @keydown.space.prevent="viewRecipients(g.id)"
          >
            <div class="group-card-top">
              <div class="group-copy">
                <h4>{{ g.name }}</h4>
                <p>{{ g.description || '未填写描述，可补充来源、标签或发送用途。' }}</p>
              </div>
              <span class="badge">{{ g.count || 0 }} 人</span>
            </div>
            <div class="group-meta">
              <span>{{ currentGroup === g.id ? '当前分组' : '点击查看成员' }}</span>
              <span>{{ g.description ? '已配置说明' : '待补充说明' }}</span>
            </div>
            <div class="group-actions">
              <button class="btn-action" @click.stop="viewRecipients(g.id)">查看</button>
              <button class="btn-action" @click.stop="openEditGroup(g)">编辑</button>
              <button class="btn-action" @click.stop="goToSend(g.id, 'batch')">群发</button>
              <button class="btn-action danger" @click.stop="deleteGroup(g.id)">删除</button>
            </div>
          </article>
        </div>

        <div v-else class="panel-empty">
          <h3>还没有分组</h3>
          <p>先创建一个分组，再导入成员后就能直接发起批量发送。</p>
          <button class="btn-primary" @click="openCreateGroup">创建第一个分组</button>
        </div>
      </section>

      <section class="card workspace-panel">
        <div class="workspace-hero" :class="{ inactive: !selectedGroup }">
          <div>
            <span class="eyebrow">收件人工作区</span>
            <h3>{{ selectedGroup ? selectedGroup.name : '选择一个分组开始管理' }}</h3>
            <p>
              {{ selectedGroup
                ? (selectedGroup.description || '当前分组还没有描述，可以继续导入收件人、修正名称或直接切换到发送页。')
                : '左侧选择一个分组后，这里会展示该分组的成员、状态和快捷操作。' }}
            </p>
          </div>
          <div class="workspace-kpis">
            <div>
              <span>总人数</span>
              <strong>{{ recipients.length }}</strong>
            </div>
            <div>
              <span>正常</span>
              <strong>{{ activeRecipientCount }}</strong>
            </div>
            <div>
              <span>黑名单</span>
              <strong>{{ blockedRecipientCount }}</strong>
            </div>
          </div>
        </div>

        <div class="workspace-toolbar">
          <div class="workspace-copy">
            <span class="status-pill" :class="{ active: !!selectedGroup }">{{ selectedGroup ? '已选择分组' : '未选择分组' }}</span>
            <p>{{ selectedGroup ? `当前正在管理 ${selectedGroup.name} 的收件人名单。` : '请选择左侧分组后再继续。' }}</p>
          </div>
          <div class="workspace-actions">
            <button class="btn-secondary" :disabled="!currentGroup" @click="showImportModal = true">批量导入</button>
            <button class="btn-secondary" :disabled="!currentGroup" @click="goToSend(currentGroup, 'batch')">去群发</button>
            <button class="btn-primary" :disabled="!currentGroup" @click="showRecipientModal = true">+ 添加收件人</button>
          </div>
        </div>

        <div v-if="currentGroup" class="workspace-table">
          <table class="data-table">
            <thead><tr><th>收件人</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              <tr v-for="r in recipients" :key="r.id">
                <td>
                  <div class="recipient-main">{{ r.email }}</div>
                  <div class="recipient-sub">{{ r.name || '未填写名称' }}</div>
                </td>
                <td><span :class="['badge', r.status === 'active' ? 'success' : 'failed']">{{ r.status === 'active' ? '正常' : '黑名单' }}</span></td>
                <td>
                  <div class="action-btns">
                    <button class="btn-action" @click="goToSend(currentGroup, 'single', r.email)">发送给此人</button>
                    <button class="btn-action danger" @click="deleteRecipient(r.id)">删除</button>
                  </div>
                </td>
              </tr>
              <tr v-if="recipients.length === 0">
                <td colspan="3" class="empty-cell"><div class="empty-state">当前分组还没有收件人</div></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else class="panel-empty workspace-empty">
          <h3>右侧工作区待激活</h3>
          <p>先选择一个分组，右侧才会显示成员列表和导入、发送等操作。</p>
        </div>
      </section>
    </div>

    <div v-if="showGroupModal" class="modal-overlay" @click.self="closeGroupModal">
      <div class="modal">
        <div class="modal-header"><h3>{{ editingGroup ? '编辑分组' : '新建分组' }}</h3><button class="modal-close" @click="closeGroupModal">×</button></div>
        <form @submit.prevent="saveGroup">
          <div class="field"><label>名称 *</label><input v-model="groupForm.name" placeholder="例如：订阅会员、测试用户" required /></div>
          <div class="field"><label>描述</label><input v-model="groupForm.description" placeholder="用于说明分组来源、场景或筛选条件" /></div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="closeGroupModal">取消</button><button type="submit" class="btn-primary">{{ editingGroup ? '保存修改' : '创建分组' }}</button></div>
        </form>
      </div>
    </div>

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
import { useRouter } from 'vue-router'
import { store, actions } from '@/store'
import axios from 'axios'

const API = '/api/v1'

export default {
  name: 'Recipients',
  setup() {
    const router = useRouter()
    const showGroupModal = ref(false)
    const showRecipientModal = ref(false)
    const showImportModal = ref(false)
    const editingGroup = ref(null)
    const currentGroup = ref('')
    const recipients = ref([])
    const groupForm = ref({ name: '', description: '' })
    const recipientForm = ref({ email: '', name: '' })
    const batchEmails = ref('')
    const groups = computed(() => store.recipientGroups)
    const selectedGroup = computed(() => groups.value.find(group => group.id === currentGroup.value) || null)
    const totalRecipientCount = computed(() => groups.value.reduce((sum, group) => sum + Number(group.count || 0), 0))
    const activeRecipientCount = computed(() => recipients.value.filter(item => item.status === 'active').length)
    const blockedRecipientCount = computed(() => recipients.value.filter(item => item.status !== 'active').length)
    
    const openCreateGroup = () => {
      editingGroup.value = null
      groupForm.value = { name: '', description: '' }
      showGroupModal.value = true
    }

    const openEditGroup = (group) => {
      editingGroup.value = group
      groupForm.value = { name: group.name || '', description: group.description || '' }
      showGroupModal.value = true
    }

    const closeGroupModal = () => {
      editingGroup.value = null
      groupForm.value = { name: '', description: '' }
      showGroupModal.value = false
    }

    const saveGroup = async () => {
      try {
        const payload = {
          name: groupForm.value.name.trim(),
          description: groupForm.value.description.trim()
        }
        if (!payload.name) {
          actions.showToast('分组名称不能为空', 'error')
          return
        }
        if (editingGroup.value) {
          await axios.put(`${API}/recipient-groups/${editingGroup.value.id}`, payload, { headers: actions.getHeaders() })
        } else {
          await axios.post(`${API}/recipient-groups`, payload, { headers: actions.getHeaders() })
        }
        await actions.loadRecipientGroups()
        closeGroupModal()
        actions.showToast(editingGroup.value ? '分组已更新' : '创建成功')
      } catch (e) {
        actions.showToast(e.response?.data?.error || (editingGroup.value ? '更新失败' : '创建失败'), 'error')
      }
    }
    
    const deleteGroup = async (id) => {
      if (!confirm('确定删除？分组内收件人也会删除。')) return
      try {
        await axios.delete(`${API}/recipient-groups/${id}`, { headers: actions.getHeaders() })
        await actions.loadRecipientGroups()
        if (currentGroup.value === id) { currentGroup.value = ''; recipients.value = [] }
        actions.showToast('已删除')
      } catch (e) {
        actions.showToast(e.response?.data?.error || '删除失败', 'error')
      }
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
        await Promise.all([viewRecipients(currentGroup.value), actions.loadRecipientGroups()])
        actions.showToast('添加成功')
      } catch (e) { actions.showToast(e.response?.data?.error || '添加失败', 'error') }
    }
    
    const deleteRecipient = async (id) => {
      try {
        await axios.delete(`${API}/recipients/${id}`, { headers: actions.getHeaders() })
        await Promise.all([viewRecipients(currentGroup.value), actions.loadRecipientGroups()])
        actions.showToast('已删除')
      } catch (e) {
        actions.showToast(e.response?.data?.error || '删除失败', 'error')
      }
    }
    
    const batchImport = async () => {
      try {
        const res = await axios.post(`${API}/recipients/batch`, { group_id: currentGroup.value, emails: batchEmails.value }, { headers: actions.getHeaders() })
        showImportModal.value = false
        batchEmails.value = ''
        await Promise.all([viewRecipients(currentGroup.value), actions.loadRecipientGroups()])
        actions.showToast(`成功导入 ${res.data.success || 0} 个收件人`)
      } catch (e) { actions.showToast(e.response?.data?.error || '导入失败', 'error') }
    }

    const goToSend = (groupId, mode = 'batch', email = '') => {
      const query = { group: groupId, mode }
      if (email) {
        query.recipient = email
      }
      router.push({ path: '/send', query })
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
    
    return {
      showGroupModal,
      showRecipientModal,
      showImportModal,
      editingGroup,
      currentGroup,
      recipients,
      groupForm,
      recipientForm,
      batchEmails,
      groups,
      selectedGroup,
      totalRecipientCount,
      activeRecipientCount,
      blockedRecipientCount,
      openCreateGroup,
      openEditGroup,
      closeGroupModal,
      saveGroup,
      deleteGroup,
      viewRecipients,
      createRecipient,
      deleteRecipient,
      batchImport,
      goToSend,
      exportRecipients
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';

.header-actions {
  display: flex;
  gap: 8px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.summary-card {
  padding: 18px 20px;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.06);
}

.summary-card strong {
  display: block;
  margin-top: 8px;
  font-size: 28px;
  color: #0f172a;
}

.summary-card small {
  display: block;
  margin-top: 8px;
  color: #64748b;
}

.summary-label {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
}

.recipients-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 16px;
}

.group-panel-head,
.workspace-toolbar,
.workspace-hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.group-panel-head p,
.workspace-copy p,
.workspace-hero p {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.7;
  color: #64748b;
}

.group-list {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.group-card {
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  cursor: pointer;
  transition: all 0.2s;
}

.group-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.08);
}

.group-card.active {
  border-color: #60a5fa;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
}

.group-card:focus-visible {
  outline: 3px solid rgba(59, 130, 246, 0.18);
  outline-offset: 2px;
}

.group-card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.group-copy h4 {
  margin: 0;
  font-size: 15px;
  color: #0f172a;
}

.group-copy p {
  margin: 4px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: #64748b;
}

.group-meta {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 10px;
  font-size: 12px;
  color: #94a3b8;
}

.group-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.workspace-hero {
  padding: 22px;
  border-bottom: 1px solid #f1f5f9;
  background: linear-gradient(135deg, #eff6ff 0%, #ffffff 100%);
}

.workspace-hero.inactive {
  background: linear-gradient(135deg, #f8fafc 0%, #ffffff 100%);
}

.eyebrow {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 999px;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 700;
}

.workspace-hero h3 {
  margin: 10px 0 0;
  font-size: 24px;
  color: #0f172a;
}

.workspace-kpis {
  display: grid;
  grid-template-columns: repeat(3, minmax(76px, 1fr));
  gap: 12px;
  min-width: 270px;
}

.workspace-kpis div {
  padding: 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid #e2e8f0;
}

.workspace-kpis span {
  display: block;
  font-size: 12px;
  color: #64748b;
}

.workspace-kpis strong {
  display: block;
  margin-top: 8px;
  font-size: 24px;
  color: #0f172a;
}

.workspace-toolbar {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.workspace-copy {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.status-pill {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 999px;
  background: #e2e8f0;
  color: #475569;
  font-size: 12px;
  font-weight: 600;
}

.status-pill.active {
  background: #dbeafe;
  color: #1d4ed8;
}

.workspace-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.workspace-table {
  overflow-x: auto;
}

.recipient-main {
  font-weight: 600;
  color: #0f172a;
}

.recipient-sub {
  margin-top: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.panel-empty {
  padding: 24px 20px 28px;
  color: #64748b;
}

.panel-empty h3 {
  margin: 0 0 8px;
  color: #0f172a;
}

.panel-empty p {
  margin: 0 0 16px;
  line-height: 1.7;
}

.workspace-empty {
  min-height: 240px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

@media (max-width: 1100px) {
  .overview-grid,
  .recipients-layout {
    grid-template-columns: 1fr;
  }

  .workspace-hero,
  .workspace-toolbar {
    flex-direction: column;
  }

  .workspace-kpis {
    width: 100%;
    min-width: 0;
  }
}

@media (max-width: 640px) {
  .header-actions,
  .workspace-actions,
  .group-actions {
    width: 100%;
  }

  .header-actions > button,
  .workspace-actions > button,
  .group-actions > button {
    flex: 1;
    justify-content: center;
  }

  .group-meta {
    flex-direction: column;
  }
}
</style>