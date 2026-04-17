import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const MainLayout = () => import('@/layouts/MainLayout.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { layout: 'blank', public: true }
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/stats',
    children: [
      { path: 'stats', name: 'Stats', component: () => import('@/views/Stats.vue'), meta: { title: 'nav.stats' } },
      { path: 'smtp', name: 'SmtpAccounts', component: () => import('@/views/SmtpAccounts.vue'), meta: { title: 'nav.smtp' } },
      { path: 'send', name: 'SendEmail', component: () => import('@/views/SendEmail.vue'), meta: { title: 'nav.send' } },
      { path: 'keys', name: 'ApiKeys', component: () => import('@/views/ApiKeys.vue'), meta: { title: 'nav.keys' } },
      { path: 'templates', name: 'Templates', component: () => import('@/views/Templates.vue'), meta: { title: 'nav.templates' } },
      { path: 'recipients', name: 'Recipients', component: () => import('@/views/Recipients.vue'), meta: { title: 'nav.recipients' } },
      { path: 'logs', name: 'Logs', component: () => import('@/views/Logs.vue'), meta: { title: 'nav.logs' } },
      { path: 'webhooks', name: 'Webhooks', component: () => import('@/views/Webhooks.vue'), meta: { title: 'nav.webhooks' } },
      { path: 'blacklist', name: 'Blacklist', component: () => import('@/views/Blacklist.vue'), meta: { title: 'nav.blacklist' } },
      { path: 'docs', name: 'Docs', component: () => import('@/views/Docs.vue'), meta: { title: 'nav.docs' } },
      { path: 'settings', name: 'Settings', component: () => import('@/views/Settings.vue'), meta: { title: 'nav.settings' } }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (!to.meta.public && !token) {
    next({ path: '/login', query: to.fullPath !== '/' ? { redirect: to.fullPath } : undefined })
  } else if (to.path === '/login' && token) {
    next('/stats')
  } else {
    next()
  }
})

export default router
