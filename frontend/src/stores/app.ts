import { defineStore } from 'pinia'
import { ref } from 'vue'

export type Locale = 'zh-CN' | 'en-US'

export const useAppStore = defineStore('app', () => {
  const locale = ref<Locale>((localStorage.getItem('locale') as Locale) || 'zh-CN')
  const sidebarCollapsed = ref<boolean>(
    localStorage.getItem('sidebar-collapsed') === '1'
  )

  function setLocale(next: Locale) {
    locale.value = next
    localStorage.setItem('locale', next)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('sidebar-collapsed', sidebarCollapsed.value ? '1' : '0')
  }

  return { locale, sidebarCollapsed, setLocale, toggleSidebar }
})
