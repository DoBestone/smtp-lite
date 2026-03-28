<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h3>修改密码</h3>
        <button class="modal-close" @click="$emit('close')">×</button>
      </div>
      <form @submit.prevent="changePassword">
        <div class="field">
          <label>旧密码</label>
          <input v-model="form.oldPwd" type="password" required />
        </div>
        <div class="field">
          <label>新密码</label>
          <input v-model="form.newPwd" type="password" required />
        </div>
        <div class="field">
          <label>确认新密码</label>
          <input v-model="form.confirmPwd" type="password" required />
        </div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <p v-if="success" class="success-msg">{{ success }}</p>
        <div class="modal-actions">
          <button type="button" class="btn-secondary" @click="$emit('close')">取消</button>
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '处理中...' : '保存' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { actions } from '@/store'
import axios from 'axios'

export default {
  name: 'ChangePasswordModal',
  emits: ['close'],
  setup(props, { emit }) {
    const router = useRouter()
    const form = ref({ oldPwd: '', newPwd: '', confirmPwd: '' })
    const loading = ref(false)
    const error = ref('')
    const success = ref('')
    
    const changePassword = async () => {
      error.value = ''
      if (form.value.newPwd !== form.value.confirmPwd) {
        error.value = '两次输入的新密码不一致'
        return
      }
      if (form.value.newPwd.length < 6) {
        error.value = '新密码至少需要 6 位字符'
        return
      }
      
      loading.value = true
      try {
        await axios.post('/api/v1/auth/change-password', {
          old_password: form.value.oldPwd,
          new_password: form.value.newPwd
        }, { headers: actions.getHeaders() })
        
        success.value = '密码修改成功，即将重新登录...'
        setTimeout(() => {
          actions.logout()
          router.push('/login')
          emit('close')
        }, 1500)
      } catch (e) {
        error.value = e.response?.data?.error || '修改失败'
      } finally {
        loading.value = false
      }
    }
    
    return { form, loading, error, success, changePassword }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
.error-msg { color: #dc2626; background: #fef2f2; padding: 10px; border-radius: 6px; font-size: 14px; }
.success-msg { color: #16a34a; background: #dcfce7; padding: 10px; border-radius: 6px; font-size: 14px; }
</style>