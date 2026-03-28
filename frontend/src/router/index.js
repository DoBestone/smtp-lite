import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/Login.vue'),
    meta: { layout: 'blank' }
  },
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', redirect: '/smtp' },
      { path: 'smtp', name: 'SmtpAccounts', component: () => import('@/pages/SmtpAccounts.vue') },
      { path: 'send', name: 'SendEmail', component: () => import('@/pages/SendEmail.vue') },
      { path: 'keys', name: 'ApiKeys', component: () => import('@/pages/ApiKeys.vue') },
      { path: 'templates', name: 'Templates', component: () => import('@/pages/Templates.vue') },
      { path: 'recipients', name: 'Recipients', component: () => import('@/pages/Recipients.vue') },
      { path: 'logs', name: 'Logs', component: () => import('@/pages/Logs.vue') },
      { path: 'stats', name: 'Stats', component: () => import('@/pages/Stats.vue') },
      { path: 'webhooks', name: 'Webhooks', component: () => import('@/pages/Webhooks.vue') },
      { path: 'blacklist', name: 'Blacklist', component: () => import('@/pages/Blacklist.vue') },
      { path: 'docs', name: 'Docs', component: () => import('@/pages/Docs.vue') },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 检查登录状态
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/smtp')
  } else {
    next()
  }
})

export default router