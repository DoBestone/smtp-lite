import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { pinia } from './stores'
import { i18n } from './i18n'

import './styles/tokens.css'

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: zhCn, size: 'default' })

app.mount('#app')
