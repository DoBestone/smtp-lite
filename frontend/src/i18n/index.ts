import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

export type MessageSchema = typeof zhCN

const initialLocale = (localStorage.getItem('locale') as 'zh-CN' | 'en-US') || 'zh-CN'

export const i18n = createI18n<[MessageSchema], 'zh-CN' | 'en-US'>({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: 'en-US',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

export function setI18nLocale(locale: 'zh-CN' | 'en-US') {
  ;(i18n.global.locale as unknown as { value: string }).value = locale
  document.documentElement.lang = locale === 'zh-CN' ? 'zh-CN' : 'en'
}

setI18nLocale(initialLocale)
