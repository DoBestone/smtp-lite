<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h3>修改登录账号</h3>
        <button class="modal-close" @click="$emit('close')">×</button>
      </div>
      <form @submit.prevent="changeUsername">
        <div class="field">
          <label>当前密码</label>
          <input v-model="form.password" type="password" placeholder="请输入当前密码" required />
        </div>
        <div class="field">
          <label>新用户名</label>
          <input v-model="form.newUsername" type="text" placeholder="至少 3 位字符" required />
        </div>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <p v-if="success" class="success-msg">{{ success }}</p>
        <div class="modal-actions">
          <button type="button" class="btn-secondary" @click="$emit('close')">取消</button>
          <button type="submit" class="btn-primary" :disabled="loading">
            {{ loading ? '处理中...' : '保存修改' }}
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
  name: 'ChangeUsernameModal',
  emits: ['close'],
  setup(props, { emit }) {
    const router = useRouter()
    const form = ref({ password: '', newUsername: '' })
    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const changeUsername = async () => {
      error.value = ''
      if (form.value.newUsername.length < 3) {
        error.value = '用户名至少需要 3 位字符'
        return
      }

      loading.value = true
      try {
        await axios.post('/api/v1/auth/change-username', {
          password: form.value.password,
          new_username: form.value.newUsername
        }, { headers: actions.getHeaders() })

        success.value = '用户名修改成功，即将重新登录...'
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

    return { form, loading, error, success, changeUsername }
  }
}
</script>

<style scoped>
@import '@/assets/styles.css';
.error-msg { color: #dc2626; background: #fef2f2; padding: 10px; border-radius: 6px; font-size: 14px; }
.success-msg { color: #16a34a; background: #dcfce7; padding: 10px; border-radius: 6px; font-size: 14px; }
</style>
