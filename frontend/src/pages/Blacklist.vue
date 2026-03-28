<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>黑名单</h2>
        <p class="page-desc">禁止向这些邮箱发送邮件</p>
      </div>
      <button class="btn-primary" @click="showModal = true">+ 添加黑名单</button>
    </div>
    
    <div class="card">
      <table class="data-table">
        <thead><tr><th>邮箱</th><th>原因</th><th>添加时间</th><th>操作</th></tr></thead>
        <tbody>
          <tr v-for="b in list" :key="b.id">
            <td class="cell-main">{{ b.email }}</td>
            <td>{{ b.reason || '-' }}</td>
            <td>{{ formatDate(b.created_at) }}</td>
            <td><button class="btn-action danger" @click="removeItem(b.id)">移除</button></td>
          </tr>
          <tr v-if="list.length === 0">
            <td colspan="4" class="empty-cell"><div class="empty-state">暂无黑名单</div></td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal">
        <div class="modal-header"><h3>添加黑名单</h3><button class="modal-close" @click="showModal = false">×</button></div>
        <form @submit.prevent="addItem">
          <div class="field"><label>邮箱 *</label><input v-model="form.email" type="email" required /></div>
          <div class="field"><label>原因</label><input v-model="form.reason" /></div>
          <div class="modal-actions"><button type="button" class="btn-secondary" @click="showModal = false">取消</button><button type="submit" class="btn-primary">添加</button></div>
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
  name: 'Blacklist',
  setup() {
    const showModal = ref(false)
    const form = ref({ email: '', reason: '' })
    const list = computed(() => store.blacklist)
    
    const addItem = async () => {
      try {
        await axios.post(`${API}/blacklist`, form.value, { headers: actions.getHeaders() })
        showModal.value = false
        form.value = { email: '', reason: '' }
        actions.loadBlacklist()
        actions.showToast('已添加')
      } catch (e) { actions.showToast('添加失败', 'error') }
    }
    
    const removeItem = async (id) => {
      if (!confirm('确定移除？')) return
      try {
        await axios.delete(`${API}/blacklist/${id}`, { headers: actions.getHeaders() })
        actions.loadBlacklist()
        actions.showToast('已移除')
      } catch (e) {}
    }
    
    const formatDate = (d) => d ? new Date(d).toLocaleString('zh-CN') : '-'
    
    onMounted(() => actions.loadBlacklist())
    
    return { showModal, form, list, addItem, removeItem, formatDate }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
</style>