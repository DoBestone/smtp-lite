<template>
  <div class="app">
    <!-- ========== 登录页 ========== -->
    <transition name="fade">
      <div v-if="!isLoggedIn" class="login-page">
        <div class="login-orb login-orb-1"></div>
        <div class="login-orb login-orb-2"></div>
        <div class="login-card">
          <div class="login-logo">
            <span class="logo-icon">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
                <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/>
                <path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
            </span>
            <span class="logo-text">SMTP Lite</span>
          </div>
          <p class="login-subtitle">个人邮箱聚合发送平台</p>

          <form @submit.prevent="login" class="login-form">
            <div class="field">
              <label>用户名</label>
              <div class="input-wrap">
                <svg class="input-icon" width="16" height="16" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.8"/>
                  <path d="M4 20c0-4 3.6-7 8-7s8 3 8 7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
                </svg>
                <input v-model="username" placeholder="请输入用户名" type="text" autocomplete="username" required />
              </div>
            </div>
            <div class="field">
              <label>密码</label>
              <div class="input-wrap">
                <svg class="input-icon" width="16" height="16" viewBox="0 0 24 24" fill="none">
                  <rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" stroke-width="1.8"/>
                  <path d="M8 11V7a4 4 0 018 0v4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
                </svg>
                <input v-model="password" placeholder="请输入密码" type="password" autocomplete="current-password" required />
              </div>
            </div>
            <p v-if="error" class="login-error">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
              {{ error }}
            </p>
            <button type="submit" class="btn-login" :class="{ loading }">
              <span v-if="!loading">登录</span>
              <span v-else class="spinner"></span>
            </button>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 主界面 ========== -->
    <div v-if="isLoggedIn" class="layout">
      <!-- 顶部导航 -->
      <header class="topbar">
        <div class="topbar-inner">
          <div class="topbar-brand">
            <span class="logo-icon sm">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/>
                <path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
            </span>
            <span class="brand-name">SMTP Lite</span>
          </div>

          <!-- 桌面端导航 -->
          <nav class="desktop-nav">
            <button
              v-for="item in navItems" :key="item.key"
              :class="['nav-btn', { active: tab === item.key }]"
              @click="switchTab(item.key)"
            >
              <component :is="'span'" class="nav-icon" v-html="item.icon"></component>
              {{ item.label }}
            </button>
          </nav>

          <div class="topbar-right">
            <div class="stats-pill" v-if="stats.total_sent !== undefined">
              <span class="pill-dot"></span>
              今日 {{ stats.today_sent || 0 }} 封
            </div>
            <button @click="logout" class="btn-logout">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              退出
            </button>
            <!-- 移动端汉堡菜单 -->
            <button class="hamburger" @click="mobileNavOpen = !mobileNavOpen">
              <span></span><span></span><span></span>
            </button>
          </div>
        </div>

        <!-- 移动端导航抽屉 -->
        <transition name="slide-down">
          <div v-if="mobileNavOpen" class="mobile-nav">
            <button
              v-for="item in navItems" :key="item.key"
              :class="['mobile-nav-btn', { active: tab === item.key }]"
              @click="switchTab(item.key); mobileNavOpen = false"
            >
              <span v-html="item.icon"></span>
              {{ item.label }}
            </button>
          </div>
        </transition>
      </header>

      <main class="main-content">
        <div class="container">

          <!-- ===== SMTP 账号 ===== -->
          <section v-if="tab === 'smtp'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">SMTP 账号</h1>
                <p class="section-desc">管理用于发送邮件的 SMTP 账号</p>
              </div>
              <button @click="showAddSmtp = true" class="btn-primary">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                添加账号
              </button>
            </div>

            <div class="card">
              <!-- 桌面表格 -->
              <div class="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>邮箱地址</th>
                      <th>SMTP 服务器</th>
                      <th>端口</th>
                      <th>日限额 / 已用</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="acc in smtpAccounts" :key="acc.id">
                      <td>
                        <div class="cell-main">{{ acc.email }}</div>
                      </td>
                      <td><span class="cell-mono">{{ acc.smtp_host }}</span></td>
                      <td>{{ acc.smtp_port }}</td>
                      <td>
                        <div class="quota-bar-wrap">
                          <span class="quota-text">{{ acc.daily_used }} / {{ acc.daily_limit || '∞' }}</span>
                          <div v-if="acc.daily_limit" class="quota-bar">
                            <div class="quota-fill" :style="{ width: Math.min(100, acc.daily_used / acc.daily_limit * 100) + '%' }"></div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span :class="['badge', acc.status === 'active' ? 'badge-success' : 'badge-muted']">
                          {{ acc.status === 'active' ? '启用' : '禁用' }}
                        </span>
                      </td>
                      <td>
                        <div class="action-group">
                          <button @click="testSmtp(acc.id)" class="btn-action">测试</button>
                          <button @click="toggleSmtp(acc.id)" class="btn-action">
                            {{ acc.status === 'active' ? '禁用' : '启用' }}
                          </button>
                          <button @click="deleteSmtp(acc.id)" class="btn-action danger">删除</button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="smtpAccounts.length === 0">
                      <td colspan="6" class="empty-cell">
                        <div class="empty-state">
                          <svg width="40" height="40" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.5"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                          <p>暂无 SMTP 账号，点击右上角添加</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- 移动端卡片列表 -->
              <div class="mobile-list">
                <div v-for="acc in smtpAccounts" :key="acc.id" class="mobile-item">
                  <div class="mobile-item-head">
                    <span class="cell-main">{{ acc.email }}</span>
                    <span :class="['badge', acc.status === 'active' ? 'badge-success' : 'badge-muted']">
                      {{ acc.status === 'active' ? '启用' : '禁用' }}
                    </span>
                  </div>
                  <div class="mobile-item-meta">
                    <span>{{ acc.smtp_host }}:{{ acc.smtp_port }}</span>
                    <span>已用 {{ acc.daily_used }} / {{ acc.daily_limit || '不限' }}</span>
                  </div>
                  <div class="action-group mt-8">
                    <button @click="testSmtp(acc.id)" class="btn-action">测试</button>
                    <button @click="toggleSmtp(acc.id)" class="btn-action">{{ acc.status === 'active' ? '禁用' : '启用' }}</button>
                    <button @click="deleteSmtp(acc.id)" class="btn-action danger">删除</button>
                  </div>
                </div>
                <div v-if="smtpAccounts.length === 0" class="empty-state">
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.5"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.5"/></svg>
                  <p>暂无 SMTP 账号</p>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== API Key ===== -->
          <section v-if="tab === 'keys'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">API Key</h1>
                <p class="section-desc">用于第三方调用发信接口的密钥</p>
              </div>
              <button @click="createApiKey" class="btn-primary">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                创建 Key
              </button>
            </div>

            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>名称</th>
                      <th>Key 前缀</th>
                      <th>最后使用</th>
                      <th>创建时间</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="key in apiKeys" :key="key.id">
                      <td><span class="cell-main">{{ key.name }}</span></td>
                      <td><code class="code-tag">{{ key.key_prefix }}****</code></td>
                      <td>{{ key.last_used_at ? formatDate(key.last_used_at) : '从未使用' }}</td>
                      <td>{{ formatDate(key.created_at) }}</td>
                      <td>
                        <button @click="deleteApiKey(key.id)" class="btn-action danger">删除</button>
                      </td>
                    </tr>
                    <tr v-if="apiKeys.length === 0">
                      <td colspan="5" class="empty-cell">
                        <div class="empty-state">
                          <svg width="40" height="40" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                          <p>暂无 API Key</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="mobile-list">
                <div v-for="key in apiKeys" :key="key.id" class="mobile-item">
                  <div class="mobile-item-head">
                    <span class="cell-main">{{ key.name }}</span>
                    <code class="code-tag">{{ key.key_prefix }}****</code>
                  </div>
                  <div class="mobile-item-meta">
                    <span>创建于 {{ formatDate(key.created_at) }}</span>
                    <span>{{ key.last_used_at ? '最近：' + formatDate(key.last_used_at) : '从未使用' }}</span>
                  </div>
                  <div class="action-group mt-8">
                    <button @click="deleteApiKey(key.id)" class="btn-action danger">删除</button>
                  </div>
                </div>
                <div v-if="apiKeys.length === 0" class="empty-state">
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.5"/></svg>
                  <p>暂无 API Key</p>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== 发送日志 ===== -->
          <section v-if="tab === 'logs'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">发送日志</h1>
                <p class="section-desc">查看所有邮件发送记录</p>
              </div>
              <button @click="loadLogs" class="btn-outline">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                刷新
              </button>
            </div>

            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>收件人</th>
                      <th>主题</th>
                      <th>状态</th>
                      <th>发送时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="log in logs" :key="log.id">
                      <td>{{ log.to_email }}</td>
                      <td>{{ log.subject || '-' }}</td>
                      <td>
                        <span :class="['badge', log.status === 'success' ? 'badge-success' : 'badge-danger']">
                          {{ log.status === 'success' ? '成功' : '失败' }}
                        </span>
                      </td>
                      <td>{{ formatDate(log.created_at) }}</td>
                    </tr>
                    <tr v-if="logs.length === 0">
                      <td colspan="4" class="empty-cell">
                        <div class="empty-state">
                          <svg width="40" height="40" viewBox="0 0 24 24" fill="none"><path d="M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                          <p>暂无发送记录</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="mobile-list">
                <div v-for="log in logs" :key="log.id" class="mobile-item">
                  <div class="mobile-item-head">
                    <span class="cell-main">{{ log.to_email }}</span>
                    <span :class="['badge', log.status === 'success' ? 'badge-success' : 'badge-danger']">
                      {{ log.status === 'success' ? '成功' : '失败' }}
                    </span>
                  </div>
                  <div class="mobile-item-meta">
                    <span>{{ log.subject || '（无主题）' }}</span>
                    <span>{{ formatDate(log.created_at) }}</span>
                  </div>
                </div>
                <div v-if="logs.length === 0" class="empty-state">
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="none"><path d="M9 11l3 3L22 4" stroke="currentColor" stroke-width="1.5"/></svg>
                  <p>暂无发送记录</p>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== 统计 ===== -->
          <section v-if="tab === 'stats'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">数据统计</h1>
                <p class="section-desc">邮件发送情况概览</p>
              </div>
              <button @click="loadStats" class="btn-outline">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                刷新
              </button>
            </div>

            <div class="stats-grid">
              <div class="stat-card blue">
                <div class="stat-icon">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8"/></svg>
                </div>
                <div class="stat-num">{{ stats.total_sent || 0 }}</div>
                <div class="stat-label">累计发送</div>
              </div>
              <div class="stat-card green">
                <div class="stat-icon">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.success || 0 }}</div>
                <div class="stat-label">发送成功</div>
              </div>
              <div class="stat-card red">
                <div class="stat-icon">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.failed || 0 }}</div>
                <div class="stat-label">发送失败</div>
              </div>
              <div class="stat-card cyan">
                <div class="stat-icon">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.8"/><path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.today_sent || 0 }}</div>
                <div class="stat-label">今日发送</div>
              </div>
              <div class="stat-card purple">
                <div class="stat-icon">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </div>
                <div class="stat-num">{{ (stats.success_rate || 0).toFixed(1) }}<span class="stat-unit">%</span></div>
                <div class="stat-label">成功率</div>
              </div>
            </div>

            <!-- API 使用说明 -->
            <div class="card api-doc-card">
              <div class="api-doc-head">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.5"/></svg>
                <span>发信 API 使用示例</span>
              </div>
              <pre class="code-block">POST /api/v1/send
X-API-Key: your-api-key

{
  "to": "recipient@example.com",
  "subject": "Hello",
  "body": "邮件内容",
  "html": false
}</pre>
            </div>
          </section>

        </div>
      </main>
    </div>

    <!-- ========== 添加 SMTP 弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showAddSmtp" class="modal-overlay" @click.self="showAddSmtp = false">
        <div class="modal-box">
          <div class="modal-head">
            <h3>添加 SMTP 账号</h3>
            <button class="modal-close" @click="showAddSmtp = false">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <form @submit.prevent="addSmtpAccount">
            <div class="field">
              <label>邮箱地址 <span class="required">*</span></label>
              <input v-model="newSmtp.email" placeholder="example@gmail.com" type="email" required />
            </div>
            <div class="field">
              <label>密码 / 授权码 <span class="required">*</span></label>
              <input v-model="newSmtp.password" placeholder="应用密码或授权码" type="password" required />
            </div>
            <div class="field">
              <label>SMTP 服务器 <span class="required">*</span></label>
              <input v-model="newSmtp.smtp_host" placeholder="如 smtp.gmail.com" required />
            </div>
            <div class="field-row">
              <div class="field">
                <label>端口</label>
                <input v-model.number="newSmtp.smtp_port" placeholder="587" type="number" />
              </div>
              <div class="field">
                <label>每日限额</label>
                <input v-model.number="newSmtp.daily_limit" placeholder="500（留空不限）" type="number" />
              </div>
            </div>
            <div class="modal-actions">
              <button type="button" class="btn-ghost" @click="showAddSmtp = false">取消</button>
              <button type="submit" class="btn-primary">添加账号</button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== API Key 展示弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="newKeyInfo" class="modal-overlay" @click.self="newKeyInfo = null">
        <div class="modal-box">
          <div class="modal-head">
            <h3>API Key 已创建</h3>
            <button class="modal-close" @click="newKeyInfo = null">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <div class="alert-warn">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" stroke="currentColor" stroke-width="1.8"/><line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><line x1="12" y1="17" x2="12.01" y2="17" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            请立即保存此 Key，它只会显示一次！
          </div>
          <div class="key-display-box">
            <code>{{ newKeyInfo.key }}</code>
            <button @click="copyKey" class="copy-btn" title="复制">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="1.8"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="1.8"/></svg>
            </button>
          </div>
          <div class="modal-actions">
            <button @click="copyKey" class="btn-primary">复制 Key</button>
            <button @click="newKeyInfo = null" class="btn-ghost">关闭</button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
import axios from 'axios'

const API = '/api/v1'

export default {
  data() {
    return {
      isLoggedIn: false,
      username: '',
      password: '',
      loading: false,
      error: '',
      tab: 'smtp',
      mobileNavOpen: false,
      smtpAccounts: [],
      apiKeys: [],
      logs: [],
      stats: {},
      showAddSmtp: false,
      newSmtp: { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 },
      newKeyInfo: null,
      navItems: [
        {
          key: 'smtp',
          label: 'SMTP 账号',
          icon: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>'
        },
        {
          key: 'keys',
          label: 'API Key',
          icon: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>'
        },
        {
          key: 'logs',
          label: '发送日志',
          icon: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>'
        },
        {
          key: 'stats',
          label: '统计',
          icon: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M18 20V10M12 20V4M6 20v-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>'
        }
      ]
    }
  },
  mounted() {
    const token = localStorage.getItem('token')
    if (token) {
      this.isLoggedIn = true
      this.loadData()
    }
  },
  methods: {
    switchTab(key) {
      this.tab = key
      if (key === 'logs') this.loadLogs()
    },
    async login() {
      this.loading = true
      this.error = ''
      try {
        const res = await axios.post(`${API}/auth/login`, {
          username: this.username,
          password: this.password
        })
        localStorage.setItem('token', res.data.token)
        this.isLoggedIn = true
        this.loadData()
      } catch (e) {
        this.error = e.response?.data?.error || '用户名或密码错误'
      } finally {
        this.loading = false
      }
    },
    logout() {
      localStorage.removeItem('token')
      this.isLoggedIn = false
      this.username = ''
      this.password = ''
    },
    getHeaders() {
      return { Authorization: `Bearer ${localStorage.getItem('token')}` }
    },
    async loadData() {
      this.loadSmtpAccounts()
      this.loadApiKeys()
      this.loadStats()
    },
    async loadSmtpAccounts() {
      try {
        const res = await axios.get(`${API}/smtp-accounts`, { headers: this.getHeaders() })
        this.smtpAccounts = res.data
      } catch (e) { console.error(e) }
    },
    async loadApiKeys() {
      try {
        const res = await axios.get(`${API}/api-keys`, { headers: this.getHeaders() })
        this.apiKeys = res.data
      } catch (e) { console.error(e) }
    },
    async loadLogs() {
      try {
        const res = await axios.get(`${API}/logs`, { headers: this.getHeaders() })
        this.logs = res.data.logs || []
      } catch (e) { console.error(e) }
    },
    async loadStats() {
      try {
        const res = await axios.get(`${API}/stats`, { headers: this.getHeaders() })
        this.stats = res.data
      } catch (e) { console.error(e) }
    },
    async addSmtpAccount() {
      try {
        await axios.post(`${API}/smtp-accounts`, this.newSmtp, { headers: this.getHeaders() })
        this.showAddSmtp = false
        this.newSmtp = { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 }
        this.loadSmtpAccounts()
      } catch (e) {
        alert(e.response?.data?.error || '添加失败')
      }
    },
    async testSmtp(id) {
      try {
        const res = await axios.post(`${API}/smtp-accounts/${id}/test`, {}, { headers: this.getHeaders() })
        alert(res.data.success ? '✅ 连接成功！' : '❌ 失败：' + res.data.error)
      } catch (e) {
        alert('测试失败：' + (e.response?.data?.error || e.message))
      }
    },
    async toggleSmtp(id) {
      try {
        await axios.post(`${API}/smtp-accounts/${id}/toggle`, {}, { headers: this.getHeaders() })
        this.loadSmtpAccounts()
      } catch (e) { alert('操作失败') }
    },
    async deleteSmtp(id) {
      if (!confirm('确定要删除此 SMTP 账号？')) return
      try {
        await axios.delete(`${API}/smtp-accounts/${id}`, { headers: this.getHeaders() })
        this.loadSmtpAccounts()
      } catch (e) { alert('删除失败') }
    },
    async createApiKey() {
      const name = prompt('请输入 Key 名称：')
      if (!name) return
      try {
        const res = await axios.post(`${API}/api-keys`, { name }, { headers: this.getHeaders() })
        this.newKeyInfo = res.data
        this.loadApiKeys()
      } catch (e) { alert('创建失败') }
    },
    async deleteApiKey(id) {
      if (!confirm('确定要删除此 API Key？')) return
      try {
        await axios.delete(`${API}/api-keys/${id}`, { headers: this.getHeaders() })
        this.loadApiKeys()
      } catch (e) { alert('删除失败') }
    },
    copyKey() {
      navigator.clipboard.writeText(this.newKeyInfo.key).then(() => alert('已复制到剪贴板'))
    },
    formatDate(date) {
      if (!date) return '-'
      return new Date(date).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style>
/* ===== Reset & Base ===== */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

:root {
  --blue: #2563eb;
  --blue-light: #3b82f6;
  --blue-50: #eff6ff;
  --blue-100: #dbeafe;
  --cyan: #0891b2;
  --cyan-50: #ecfeff;
  --green: #16a34a;
  --green-50: #f0fdf4;
  --red: #dc2626;
  --red-50: #fef2f2;
  --purple: #7c3aed;
  --purple-50: #f5f3ff;
  --gray-50: #f8fafc;
  --gray-100: #f1f5f9;
  --gray-200: #e2e8f0;
  --gray-300: #cbd5e1;
  --gray-400: #94a3b8;
  --gray-500: #64748b;
  --gray-600: #475569;
  --gray-700: #334155;
  --gray-900: #0f172a;
  --radius-sm: 6px;
  --radius: 10px;
  --radius-lg: 14px;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow: 0 1px 3px rgba(0,0,0,0.07), 0 4px 12px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 16px rgba(0,0,0,0.08), 0 1px 4px rgba(0,0,0,0.04);
  --topbar-h: 60px;
}

html { font-size: 15px; -webkit-text-size-adjust: 100%; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Inter', 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: var(--gray-50);
  color: var(--gray-700);
  line-height: 1.6;
  min-height: 100vh;
}

input, button, select { font-family: inherit; font-size: inherit; }
button { cursor: pointer; }
a { color: inherit; text-decoration: none; }

/* ===== 登录页 ===== */
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f0f7ff 0%, #e8f4fd 40%, #f0f0ff 100%);
  position: relative;
  overflow: hidden;
  padding: 24px;
}

.login-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.35;
  pointer-events: none;
}
.login-orb-1 {
  width: 400px; height: 400px;
  background: radial-gradient(circle, #93c5fd, #3b82f6);
  top: -100px; left: -80px;
}
.login-orb-2 {
  width: 350px; height: 350px;
  background: radial-gradient(circle, #a5f3fc, #06b6d4);
  bottom: -80px; right: -60px;
}

.login-card {
  background: rgba(255,255,255,0.9);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255,255,255,0.8);
  border-radius: 20px;
  padding: clamp(28px, 5vw, 44px);
  width: 100%;
  max-width: 420px;
  box-shadow: var(--shadow-md), 0 0 0 1px rgba(37,99,235,0.05);
  position: relative;
  z-index: 1;
}

.login-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.logo-icon {
  width: 40px; height: 40px;
  background: linear-gradient(135deg, var(--blue), var(--cyan));
  border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  color: white;
  flex-shrink: 0;
}
.logo-icon.sm { width: 32px; height: 32px; border-radius: 8px; }
.logo-icon.sm svg { width: 16px; height: 16px; }

.brand-name {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--gray-900);
  letter-spacing: -0.02em;
}

.logo-text {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--gray-900);
  letter-spacing: -0.02em;
}

.login-subtitle {
  color: var(--gray-500);
  font-size: 0.875rem;
  margin-bottom: 28px;
  margin-left: 50px;
}

.login-form { display: flex; flex-direction: column; gap: 14px; }

.field { display: flex; flex-direction: column; gap: 6px; }

.field label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--gray-600);
  letter-spacing: 0.01em;
}

.required { color: var(--red); }

.input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 12px;
  color: var(--gray-400);
  pointer-events: none;
  flex-shrink: 0;
}

.input-wrap input,
.field > input {
  width: 100%;
  padding: 10px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  background: white;
  color: var(--gray-900);
  font-size: 0.9375rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  outline: none;
}
.input-wrap input { padding-left: 38px; }
.input-wrap input:focus,
.field > input:focus {
  border-color: var(--blue-light);
  box-shadow: 0 0 0 3px rgba(59,130,246,0.12);
}

.login-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--red);
  font-size: 0.8125rem;
  background: var(--red-50);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid #fecaca;
}

.btn-login {
  width: 100%;
  padding: 11px;
  background: linear-gradient(135deg, var(--blue), var(--blue-light));
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.9375rem;
  font-weight: 600;
  letter-spacing: 0.01em;
  transition: opacity 0.2s, transform 0.15s;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  height: 44px;
  margin-top: 4px;
}
.btn-login:hover { opacity: 0.92; transform: translateY(-1px); }
.btn-login:active { transform: translateY(0); }
.btn-login.loading { opacity: 0.7; pointer-events: none; }

.spinner {
  width: 18px; height: 18px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 顶部导航 ===== */
.topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--gray-100);
  box-shadow: 0 1px 0 var(--gray-100);
}

.topbar-inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 clamp(16px, 3vw, 32px);
  height: var(--topbar-h);
  display: flex;
  align-items: center;
  gap: 16px;
}

.topbar-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-right: 8px;
}

.desktop-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
}

.nav-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--gray-500);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
}
.nav-btn:hover { background: var(--gray-100); color: var(--gray-700); }
.nav-btn.active {
  background: var(--blue-50);
  color: var(--blue);
  font-weight: 600;
}

.nav-icon { display: flex; align-items: center; }

.topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.stats-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--blue-50);
  color: var(--blue);
  font-size: 0.8125rem;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 999px;
  white-space: nowrap;
}
.pill-dot {
  width: 6px; height: 6px;
  background: #22c55e;
  border-radius: 50%;
}

.btn-logout {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1.5px solid var(--gray-200);
  background: white;
  color: var(--gray-500);
  border-radius: var(--radius-sm);
  font-size: 0.8125rem;
  font-weight: 500;
  transition: border-color 0.2s, color 0.2s;
  white-space: nowrap;
}
.btn-logout:hover { border-color: var(--gray-300); color: var(--gray-700); }

.hamburger {
  display: none;
  flex-direction: column;
  gap: 5px;
  padding: 8px;
  border: none;
  background: transparent;
}
.hamburger span {
  display: block;
  width: 20px; height: 2px;
  background: var(--gray-600);
  border-radius: 2px;
  transition: 0.2s;
}

.mobile-nav {
  border-top: 1px solid var(--gray-100);
  padding: 8px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mobile-nav-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--gray-600);
  font-size: 0.9375rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
  text-align: left;
}
.mobile-nav-btn:hover { background: var(--gray-50); }
.mobile-nav-btn.active { background: var(--blue-50); color: var(--blue); font-weight: 600; }

/* ===== 主内容区 ===== */
.layout { min-height: 100vh; display: flex; flex-direction: column; }

.main-content {
  flex: 1;
  padding: clamp(20px, 3vw, 36px) 0 48px;
}

.container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 clamp(16px, 3vw, 32px);
}

/* ===== Section ===== */
.section { display: flex; flex-direction: column; gap: 20px; }

.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.section-title {
  font-size: clamp(1.25rem, 2.5vw, 1.5rem);
  font-weight: 700;
  color: var(--gray-900);
  letter-spacing: -0.02em;
  line-height: 1.3;
}

.section-desc {
  font-size: 0.875rem;
  color: var(--gray-400);
  margin-top: 3px;
}

/* ===== 卡片 ===== */
.card {
  background: white;
  border: 1px solid var(--gray-100);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  overflow: hidden;
}

/* ===== 按钮 ===== */
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 18px;
  background: linear-gradient(135deg, var(--blue), var(--blue-light));
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 600;
  transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 1px 3px rgba(37,99,235,0.25);
  white-space: nowrap;
}
.btn-primary:hover { opacity: 0.92; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(37,99,235,0.3); }
.btn-primary:active { transform: translateY(0); }

.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: white;
  color: var(--gray-600);
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
  transition: border-color 0.2s, color 0.2s, background 0.2s;
  white-space: nowrap;
}
.btn-outline:hover { border-color: var(--blue-light); color: var(--blue); background: var(--blue-50); }

.btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: transparent;
  color: var(--gray-500);
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
}
.btn-ghost:hover { background: var(--gray-50); color: var(--gray-700); }

/* ===== 操作按钮 ===== */
.action-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.btn-action {
  padding: 4px 12px;
  border: 1.5px solid var(--gray-200);
  background: white;
  color: var(--gray-600);
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
  white-space: nowrap;
}
.btn-action:hover { border-color: var(--blue-light); color: var(--blue); background: var(--blue-50); }
.btn-action.danger { color: var(--red); border-color: #fecaca; }
.btn-action.danger:hover { background: var(--red-50); border-color: #f87171; }

.mt-8 { margin-top: 8px; }

/* ===== 表格 ===== */
.table-wrap { overflow-x: auto; }

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

thead tr {
  border-bottom: 1.5px solid var(--gray-100);
}

th {
  padding: 12px 16px;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--gray-400);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  background: var(--gray-50);
}

td {
  padding: 13px 16px;
  border-bottom: 1px solid var(--gray-50);
  color: var(--gray-700);
  vertical-align: middle;
}

tbody tr:last-child td { border-bottom: none; }

tbody tr {
  transition: background 0.12s;
}
tbody tr:hover { background: var(--blue-50); }

.cell-main {
  font-weight: 500;
  color: var(--gray-900);
}

.cell-mono {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
  color: var(--gray-600);
}

/* 配额进度条 */
.quota-bar-wrap { display: flex; flex-direction: column; gap: 4px; min-width: 90px; }
.quota-text { font-size: 0.8125rem; color: var(--gray-600); }
.quota-bar {
  height: 4px;
  background: var(--gray-100);
  border-radius: 9px;
  overflow: hidden;
}
.quota-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--blue), var(--cyan));
  border-radius: 9px;
  transition: width 0.4s;
}

/* ===== Badge ===== */
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  white-space: nowrap;
}
.badge-success { background: var(--green-50); color: var(--green); border: 1px solid #bbf7d0; }
.badge-danger  { background: var(--red-50);   color: var(--red);   border: 1px solid #fecaca; }
.badge-muted   { background: var(--gray-100); color: var(--gray-500); border: 1px solid var(--gray-200); }

/* ===== Empty state ===== */
.empty-cell { text-align: center; padding: 40px 16px !important; }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--gray-400);
  padding: 16px;
}
.empty-state svg { opacity: 0.4; }
.empty-state p { font-size: 0.875rem; }

/* ===== 移动端卡片列表（隐藏） ===== */
.mobile-list { display: none; }

.mobile-item {
  padding: 14px 16px;
  border-bottom: 1px solid var(--gray-50);
}
.mobile-item:last-child { border-bottom: none; }

.mobile-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.mobile-item-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.8125rem;
  color: var(--gray-400);
  gap: 8px;
}

/* ===== 统计卡片 ===== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(180px, 100%), 1fr));
  gap: 16px;
}

.stat-card {
  background: white;
  border: 1px solid var(--gray-100);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-shadow: var(--shadow-sm);
  transition: box-shadow 0.2s, transform 0.15s;
}
.stat-card:hover { box-shadow: var(--shadow-md); transform: translateY(-2px); }

.stat-icon {
  width: 40px; height: 40px;
  border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
}

.stat-card.blue  .stat-icon { background: var(--blue-50); color: var(--blue); }
.stat-card.green .stat-icon { background: var(--green-50); color: var(--green); }
.stat-card.red   .stat-icon { background: var(--red-50); color: var(--red); }
.stat-card.cyan  .stat-icon { background: var(--cyan-50); color: var(--cyan); }
.stat-card.purple .stat-icon { background: var(--purple-50); color: var(--purple); }

.stat-num {
  font-size: clamp(1.75rem, 3vw, 2.25rem);
  font-weight: 700;
  color: var(--gray-900);
  letter-spacing: -0.03em;
  line-height: 1;
}
.stat-unit { font-size: 1.25rem; font-weight: 600; }

.stat-label {
  font-size: 0.8125rem;
  color: var(--gray-400);
  font-weight: 500;
}

/* ===== API 文档卡片 ===== */
.api-doc-card { padding: 20px; }
.api-doc-head {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--gray-700);
  font-weight: 600;
  font-size: 0.9375rem;
  margin-bottom: 14px;
}
.api-doc-head svg { color: var(--blue); }

.code-block {
  background: var(--gray-50);
  border: 1px solid var(--gray-100);
  border-radius: var(--radius-sm);
  padding: 16px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.8125rem;
  color: var(--gray-700);
  line-height: 1.7;
  overflow-x: auto;
  white-space: pre;
}

/* ===== 弹窗 ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 20px;
}

.modal-box {
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  padding: clamp(20px, 4vw, 28px);
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.modal-head h3 {
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--gray-900);
}

.modal-close {
  width: 30px; height: 30px;
  border: none;
  background: var(--gray-100);
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  color: var(--gray-500);
  transition: background 0.15s, color 0.15s;
}
.modal-close:hover { background: var(--gray-200); color: var(--gray-700); }

.modal-box form { display: flex; flex-direction: column; gap: 14px; }

.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
  padding-top: 16px;
  border-top: 1px solid var(--gray-100);
}

.alert-warn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  color: #92400e;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 14px;
}

.key-display-box {
  background: var(--gray-50);
  border: 1.5px dashed var(--gray-200);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 6px;
}
.key-display-box code {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
  color: var(--gray-700);
  word-break: break-all;
  flex: 1;
}
.copy-btn {
  padding: 6px;
  border: 1px solid var(--gray-200);
  background: white;
  border-radius: 6px;
  color: var(--gray-400);
  display: flex; align-items: center;
  transition: color 0.15s, border-color 0.15s;
  flex-shrink: 0;
}
.copy-btn:hover { color: var(--blue); border-color: var(--blue-light); }

.code-tag {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
  background: var(--gray-100);
  color: var(--gray-600);
  padding: 2px 8px;
  border-radius: 4px;
}

/* ===== 过渡动画 ===== */
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.modal-fade-enter-active, .modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-active .modal-box,
.modal-fade-leave-active .modal-box {
  transition: transform 0.2s ease, opacity 0.2s ease;
}
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-from .modal-box { transform: scale(0.96) translateY(8px); }

.slide-down-enter-active, .slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }

/* ===== 响应式 ===== */

/* >= 992px: 桌面端 */
@media (min-width: 992px) {
  .hamburger { display: none; }
  .mobile-nav { display: none !important; }
  .table-wrap { display: block; }
  .mobile-list { display: none !important; }
  .stats-pill { display: inline-flex; }
}

/* 768px ~ 991px: 平板 */
@media (max-width: 991px) {
  .desktop-nav { display: none; }
  .hamburger { display: flex; }
  .stats-pill { display: none; }
  .table-wrap { display: block; }
  .mobile-list { display: none !important; }
  th, td { padding: 10px 12px; }
}

/* < 768px: 手机 */
@media (max-width: 767px) {
  :root { --topbar-h: 56px; }
  .desktop-nav { display: none; }
  .hamburger { display: flex; }
  .stats-pill { display: none; }

  /* 隐藏桌面表格，显示移动端卡片 */
  .table-wrap { display: none !important; }
  .mobile-list { display: block !important; }

  .section-head { flex-direction: column; align-items: flex-start; }
  .section-head .btn-primary, .section-head .btn-outline { align-self: flex-start; }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-card { padding: 16px; }

  .field-row { grid-template-columns: 1fr; }

  .modal-box { padding: 20px; border-radius: 16px; }

  .btn-logout span { display: none; }
  .btn-logout { padding: 6px 10px; }
}

/* < 480px: 小屏手机 */
@media (max-width: 479px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .stat-num { font-size: 1.5rem; }
  .login-card { padding: 24px 20px; }
}
</style>
