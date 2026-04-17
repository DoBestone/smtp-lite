import axios, { AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

export const API_BASE = '/api/v1'

const client: AxiosInstance = axios.create({
  baseURL: API_BASE,
  timeout: 30_000
})

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string; message?: string }>) => {
    const status = error.response?.status
    const msg =
      error.response?.data?.error ||
      error.response?.data?.message ||
      error.message ||
      '请求失败'

    if (status === 401) {
      localStorage.removeItem('token')
      if (location.pathname !== '/login') {
        ElMessage.error('登录已过期，请重新登录')
        location.replace('/login')
      }
    } else if (status && status >= 500) {
      ElMessage.error(`服务器错误：${msg}`)
    } else if (status) {
      ElMessage.error(msg)
    } else {
      ElMessage.error('网络异常，请检查连接')
    }

    return Promise.reject(error)
  }
)

export default client
