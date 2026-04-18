import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { systemApi } from '@/api'

const DISMISS_KEY = 'force-update-dismissed-at'
const DISMISS_TTL_MS = 30 * 60 * 1000

export const useVersionStore = defineStore('version', () => {
  const currentVersion = ref('')
  const latestVersion = ref('')
  const changelog = ref('')
  const hasUpdate = ref(false)
  const forceUpdate = ref(false)
  const checking = ref(false)
  const updating = ref(false)
  const lastError = ref('')

  const forceUpdateDismissed = ref(isDismissed())

  const updateStatus = computed<'available' | 'latest' | ''>(() => {
    if (!currentVersion.value) return ''
    return hasUpdate.value ? 'available' : 'latest'
  })

  // 强制更新横幅：有更新 + force_update + 未被临时忽略
  const shouldShowForceBanner = computed(
    () => hasUpdate.value && forceUpdate.value && !forceUpdateDismissed.value
  )

  async function loadVersion() {
    try {
      const v = await systemApi.version()
      currentVersion.value = v.version
    } catch {
      /* ignore */
    }
  }

  async function checkUpdate(silent = false) {
    checking.value = true
    lastError.value = ''
    try {
      const r = await systemApi.updateCheck()
      currentVersion.value = r.current || currentVersion.value
      latestVersion.value = r.latest || ''
      hasUpdate.value = !!r.has_update
      forceUpdate.value = !!r.force_update
      changelog.value = r.changelog?.trim() ?? ''
      if (r.error) lastError.value = r.error
    } catch (e) {
      if (!silent) lastError.value = (e as Error).message || 'check failed'
    } finally {
      checking.value = false
    }
  }

  function dismissForceBanner() {
    forceUpdateDismissed.value = true
    localStorage.setItem(DISMISS_KEY, String(Date.now()))
  }

  function isDismissed(): boolean {
    const raw = localStorage.getItem(DISMISS_KEY)
    if (!raw) return false
    const ts = Number(raw)
    if (!Number.isFinite(ts)) return false
    return Date.now() - ts < DISMISS_TTL_MS
  }

  // 触发更新 → 启动后端更新 → 轮询版本号直到变化 → 刷新页面
  async function runUpdate(): Promise<void> {
    updating.value = true
    try {
      const prep = await systemApi.updatePrepare()
      if (!prep.confirm_token) throw new Error('no confirm_token')
      await systemApi.update(prep.confirm_token)
      await pollUntilUpgraded(currentVersion.value)
      // 版本已变化，刷新拿新 bundle
      window.location.reload()
    } finally {
      updating.value = false
    }
  }

  async function pollUntilUpgraded(oldVersion: string, timeoutMs = 180_000) {
    const started = Date.now()
    while (Date.now() - started < timeoutMs) {
      await sleep(3000)
      try {
        const v = await systemApi.version()
        if (v.version && v.version !== oldVersion) {
          currentVersion.value = v.version
          return
        }
      } catch {
        /* 服务重启中请求会失败，继续轮询 */
      }
    }
  }

  return {
    currentVersion,
    latestVersion,
    changelog,
    hasUpdate,
    forceUpdate,
    checking,
    updating,
    lastError,
    forceUpdateDismissed,
    updateStatus,
    shouldShowForceBanner,
    loadVersion,
    checkUpdate,
    runUpdate,
    dismissForceBanner
  }
})

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms))
}
