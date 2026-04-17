import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') ?? '')
  const username = ref<string>(localStorage.getItem('username') ?? '')

  const isLoggedIn = computed(() => !!token.value)

  async function login(user: string, password: string) {
    const res = await authApi.login(user, password)
    token.value = res.token
    username.value = res.username ?? user
    localStorage.setItem('token', res.token)
    localStorage.setItem('username', username.value)
    return res
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  function setUsername(name: string) {
    username.value = name
    localStorage.setItem('username', name)
  }

  return { token, username, isLoggedIn, login, logout, setUsername }
})
