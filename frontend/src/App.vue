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
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
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
                <svg class="input-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.8"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                <input v-model="loginForm.username" placeholder="请输入用户名" type="text" autocomplete="username" required />
              </div>
            </div>
            <div class="field">
              <label>密码</label>
              <div class="input-wrap">
                <svg class="input-icon" width="16" height="16" viewBox="0 0 24 24" fill="none"><rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" stroke-width="1.8"/><path d="M8 11V7a4 4 0 018 0v4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                <input v-model="loginForm.password" placeholder="请输入密码" type="password" autocomplete="current-password" required />
              </div>
            </div>
            <p v-if="loginError" class="form-error">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
              {{ loginError }}
            </p>
            <button type="submit" class="btn-login" :class="{ loading: loginLoading }">
              <span v-if="!loginLoading">登 录</span>
              <span v-else class="spinner"></span>
            </button>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 主界面 ========== -->
    <div v-if="isLoggedIn" class="layout">
      <!-- 移动端顶栏 -->
      <header class="mobile-topbar">
        <div class="topbar-inner">
          <div class="topbar-brand">
            <span class="logo-icon sm">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/>
                <path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
            </span>
            <span class="brand-name">SMTP Lite</span>
          </div>
          <button class="hamburger" @click="sidebarOpen = !sidebarOpen">
            <span></span><span></span><span></span>
          </button>
        </div>
      </header>

      <!-- 移动端侧边栏遮罩 -->
      <transition name="fade">
        <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>
      </transition>

      <!-- 侧边栏 -->
      <aside :class="['sidebar', { open: sidebarOpen }]">
        <div class="sidebar-header">
          <span class="logo-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
              <rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/>
              <path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
            </svg>
          </span>
          <span class="brand-name">SMTP Lite</span>
        </div>

        <div class="sidebar-stats" v-if="stats.today_sent !== undefined">
          <span class="pill-dot"></span>
          今日已发 {{ stats.today_sent || 0 }} 封
        </div>

        <nav class="sidebar-nav">
          <button v-for="item in navItems" :key="item.key"
            :class="['sidebar-nav-btn', { active: tab === item.key }]"
            @click="switchTab(item.key); sidebarOpen = false">
            <span class="nav-icon" v-html="item.icon"></span>
            {{ item.label }}
            <span v-if="item.key === 'system' && updateStatus === 'available'" class="nav-badge">新版本</span>
          </button>
        </nav>

        <div class="sidebar-footer">
          <button class="sidebar-footer-btn" title="修改密码" @click="showChangePwd = true; sidebarOpen = false">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.8"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
            修改密码
          </button>
          <button class="sidebar-footer-btn logout" @click="logout">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
            退出登录
          </button>
        </div>
      </aside>

      <main class="main-content">
        <div class="container">

          <!-- ===== SMTP 账号 ===== -->
          <section v-if="tab === 'smtp'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">SMTP 账号</h1>
                <p class="section-desc">管理用于发送邮件的 SMTP 账号，支持多账号轮询与故障转移</p>
              </div>
              <button @click="openAddSmtp" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                添加账号
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr>
                    <th>邮箱地址</th><th>SMTP 服务器</th><th>端口</th>
                    <th>日限额 / 已用</th><th>状态</th><th>操作</th>
                  </tr></thead>
                  <tbody>
                    <tr v-for="acc in smtpAccounts" :key="acc.id">
                  <td>
                    <span class="cell-main">{{ acc.email }}</span>
                    <p v-if="acc.last_error" class="text-sm text-danger mt-4">{{ acc.last_error }}</p>
                  </td>
                      <td><span class="cell-mono">{{ acc.smtp_host }}</span></td>
                      <td>{{ acc.smtp_port }}</td>
                      <td>
                        <div class="quota-wrap">
                          <span class="quota-text">{{ acc.daily_used }} / {{ acc.daily_limit || '∞' }}</span>
                          <div v-if="acc.daily_limit" class="quota-bar">
                            <div class="quota-fill" :style="{ width: Math.min(100, (acc.daily_used / acc.daily_limit) * 100) + '%', background: acc.daily_used / acc.daily_limit > 0.8 ? 'var(--red)' : 'var(--gradient-blue)' }"></div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span :class="['badge', acc.status === 'active' ? 'badge-success' : 'badge-muted']">
                          <span class="badge-dot"></span>{{ acc.status === 'active' ? '启用' : '禁用' }}
                        </span>
                      </td>
                      <td>
                        <div class="action-group">
                          <button @click="testConnection(acc)" class="btn-action" :disabled="testingId === acc.id">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                            测试连接
                          </button>
                          <button @click="openTestSend(acc)" class="btn-action">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                            发送测试
                          </button>
                          <button @click="toggleSmtp(acc.id)" class="btn-action">
                            {{ acc.status === 'active' ? '禁用' : '启用' }}
                          </button>
                          <button @click="openEditSmtp(acc)" class="btn-action">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                            编辑
                          </button>
                          <button @click="deleteSmtp(acc.id)" class="btn-action danger">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><polyline points="3 6 5 6 21 6" stroke="currentColor" stroke-width="1.8"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6" stroke="currentColor" stroke-width="1.8"/><path d="M10 11v6M14 11v6M9 6V4a1 1 0 011-1h4a1 1 0 011 1v2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="smtpAccounts.length === 0">
                      <td colspan="6" class="empty-cell">
                        <div class="empty-state">
                          <svg width="44" height="44" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.3"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.3"/></svg>
                          <p>暂无 SMTP 账号，点击右上角添加</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <!-- 移动端卡片 -->
              <div class="mobile-list">
                <div v-for="acc in smtpAccounts" :key="acc.id" class="mobile-item">
                  <div class="mobile-item-head">
                    <span class="cell-main">{{ acc.email }}</span>
                    <span :class="['badge', acc.status === 'active' ? 'badge-success' : 'badge-muted']">
                      <span class="badge-dot"></span>{{ acc.status === 'active' ? '启用' : '禁用' }}
                    </span>
                  </div>
                  <div class="mobile-item-meta">
                    <span>{{ acc.smtp_host }}:{{ acc.smtp_port }}</span>
                    <span>已用 {{ acc.daily_used }} / {{ acc.daily_limit || '不限' }}</span>
                  </div>
                  <p v-if="acc.last_error" class="text-sm text-danger mt-4">{{ acc.last_error }}</p>
                  <div class="action-group mt-8">
                    <button @click="testConnection(acc)" class="btn-action">测试连接</button>
                    <button @click="openTestSend(acc)" class="btn-action">发送测试</button>
                    <button @click="toggleSmtp(acc.id)" class="btn-action">{{ acc.status === 'active' ? '禁用' : '启用' }}</button>
                    <button @click="openEditSmtp(acc)" class="btn-action">编辑</button>
                    <button @click="deleteSmtp(acc.id)" class="btn-action danger">删除</button>
                  </div>
                </div>
                <div v-if="smtpAccounts.length === 0" class="empty-state p-20">
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.3"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.3"/></svg>
                  <p>暂无 SMTP 账号</p>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== 发送邮件 ===== -->
          <section v-if="tab === 'send'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">发送邮件</h1>
                <p class="section-desc">在线发送邮件，支持单发、批量发送、定时发送</p>
              </div>
              <div style="display:flex;gap:8px">
                <select v-model="sendMode" class="form-select" @change="resetSendForm">
                  <option value="single">单封发送</option>
                  <option value="batch">批量发送</option>
                  <option value="scheduled">定时发送</option>
                </select>
              </div>
            </div>
            <div class="card">
              <form @submit.prevent="doSendEmail">
                <!-- 单发/定时发送的收件人 -->
                <div v-if="sendMode !== 'batch'" class="field">
                  <label>收件人 <span class="required">*</span></label>
                  <input v-model="sendForm.to" type="email" placeholder="recipient@example.com" required />
                </div>
                <div v-if="sendMode !== 'batch'" class="recipient-link-box">
                  <div class="field-row recipient-link-row">
                    <div class="field">
                      <label>收件人分组</label>
                      <select v-model="sendRecipientGroupId" class="form-select" @change="handleSendRecipientGroupChange">
                        <option value="">选择分组</option>
                        <option v-for="group in recipientGroups" :key="group.id" :value="group.id">{{ group.name }} ({{ group.count || 0 }})</option>
                      </select>
                    </div>
                    <div class="field">
                      <label>分组成员</label>
                      <select v-model="sendRecipientSelection" class="form-select" :disabled="!sendRecipientOptions.length" @change="applySendRecipientSelection">
                        <option value="">选择收件人</option>
                        <option v-for="recipient in sendRecipientOptions" :key="recipient.id" :value="recipient.email">{{ recipient.email }}{{ recipient.name ? ` (${recipient.name})` : '' }}</option>
                      </select>
                    </div>
                  </div>
                  <p class="field-hint">仅显示正常状态成员，选择后会自动带入上方收件人。</p>
                </div>
                <!-- 批量发送的收件人列表 -->
                <div v-if="sendMode === 'batch'" class="field">
                  <label>收件人列表 <span class="required">*</span></label>
                  <textarea v-model="batchEmailsList" rows="5" placeholder="每行一个邮箱地址&#10;user1@example.com&#10;user2@example.com" required></textarea>
                  <p class="field-hint">每行一个邮箱地址，或从收件人分组中选择</p>
                </div>
                <div v-if="sendMode === 'batch'" class="recipient-link-box">
                  <div class="field-row recipient-link-row">
                    <div class="field">
                      <label>从分组导入</label>
                      <select v-model="sendRecipientGroupId" class="form-select" @change="handleSendRecipientGroupChange">
                        <option value="">选择收件人分组</option>
                        <option v-for="group in recipientGroups" :key="group.id" :value="group.id">{{ group.name }} ({{ group.count || 0 }})</option>
                      </select>
                    </div>
                    <div class="field recipient-link-action">
                      <label>操作</label>
                      <button type="button" class="btn-outline recipient-import-btn" :disabled="!sendRecipientGroupId || !sendRecipientOptions.length" @click="importSendGroupRecipients">导入正常成员</button>
                    </div>
                  </div>
                  <p class="field-hint">{{ sendRecipientGroupId ? `当前分组可导入 ${sendRecipientOptions.length} 个正常状态收件人，导入时会自动去重。` : '选择一个收件人分组后，可以一键导入该分组中的正常成员。' }}</p>
                </div>
                <!-- 定时发送时间 -->
                <div v-if="sendMode === 'scheduled'" class="field">
                  <label>发送时间 <span class="required">*</span></label>
                  <input v-model="scheduledTime" type="datetime-local" required />
                </div>
                <div class="field"><label>邮件主题 <span class="required">*</span></label><input v-model="sendForm.subject" required /></div>
                <div class="field">
                  <label>邮件内容 <span class="required">*</span></label>
                  <textarea v-model="sendForm.body" rows="8" placeholder="邮件正文..." required></textarea>
                  <div class="field-actions">
                    <label class="checkbox-label"><input type="checkbox" v-model="sendForm.is_html" /> HTML 格式</label>
                    <label v-if="sendForm.is_html" class="checkbox-label" style="margin-left:16px"><input type="checkbox" v-model="sendForm.track_enabled" /> 启用追踪</label>
                  </div>
                </div>
                <div class="field-row">
                  <div class="field"><label>发件人名称</label><input v-model="sendForm.from_name" placeholder="显示名称（可选）" /></div>
                  <div v-if="sendMode === 'single'" class="field"><label>抄送</label><input v-model="sendForm.cc" placeholder="多人用逗号分隔" /></div>
                </div>
                <div v-if="sendMode === 'single'" class="field"><label>密送</label><input v-model="sendForm.bcc" placeholder="多人用逗号分隔" /></div>
                <!-- 发送结果 -->
                <div v-if="sendResult" class="alert" :class="sendResult.success ? 'alert-success' : 'alert-error'">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <path v-if="sendResult.success" d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    <template v-else><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></template>
                  </svg>
                  <div>
                    <div>{{ sendResult.message }}</div>
                    <div v-if="sendResult.details && sendResult.details.length" style="margin-top:8px;display:flex;flex-direction:column;gap:6px;font-size:0.8125rem;line-height:1.5;opacity:0.95">
                      <div v-for="(detail, index) in sendResult.details" :key="index">- {{ detail }}</div>
                    </div>
                  </div>
                </div>
                <div class="modal-actions" style="margin-top:20px">
                  <button type="button" class="btn-ghost" @click="resetSendForm">清空</button>
                  <button type="submit" class="btn-primary" :disabled="sendLoading">
                    <span v-if="!sendLoading">{{ sendMode === 'single' ? '立即发送' : sendMode === 'batch' ? '批量发送' : '定时发送' }}</span>
                    <span v-else class="spinner"></span>
                  </button>
                </div>
              </form>
            </div>
            <!-- 快速选择模板 -->
            <div v-if="templates.length > 0" class="card mt-16">
              <div class="card-head">快速选择模板</div>
              <div style="display:flex;flex-wrap:wrap;gap:8px;padding:0 24px 24px">
                <button v-for="t in templates" :key="t.id" class="btn-outline btn-sm" @click="applyTemplate(t)">{{ t.name }}</button>
              </div>
            </div>
          </section>

          <!-- ===== API Key ===== -->
          <section v-if="tab === 'keys'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">API Key</h1>
                <p class="section-desc">用于第三方服务调用发信接口的密钥管理</p>
              </div>
              <button @click="createApiKey" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                创建 Key
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr>
                    <th>名称</th><th>Key 前缀</th><th>最后使用</th><th>创建时间</th><th>操作</th>
                  </tr></thead>
                  <tbody>
                    <tr v-for="key in apiKeys" :key="key.id">
                      <td><span class="cell-main">{{ key.name }}</span></td>
                      <td><code class="code-tag">{{ key.key_prefix }}••••••••</code></td>
                      <td>{{ key.last_used_at ? formatDate(key.last_used_at) : '从未使用' }}</td>
                      <td>{{ formatDate(key.created_at) }}</td>
                      <td>
                        <div class="action-group">
                          <button @click="resetApiKey(key.id, key.name)" class="btn-action">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                            重置
                          </button>
                          <button @click="deleteApiKey(key.id)" class="btn-action danger">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><polyline points="3 6 5 6 21 6" stroke="currentColor" stroke-width="1.8"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6" stroke="currentColor" stroke-width="1.8"/></svg>
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="apiKeys.length === 0">
                      <td colspan="5" class="empty-cell">
                        <div class="empty-state">
                          <svg width="44" height="44" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>
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
                    <code class="code-tag">{{ key.key_prefix }}••••</code>
                  </div>
                  <div class="mobile-item-meta">
                    <span>创建 {{ formatDate(key.created_at) }}</span>
                    <span>{{ key.last_used_at ? '最近：' + formatDate(key.last_used_at) : '从未使用' }}</span>
                  </div>
                  <div class="action-group mt-8">
                    <button @click="resetApiKey(key.id, key.name)" class="btn-action">
                      <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                      重置
                    </button>
                    <button @click="deleteApiKey(key.id)" class="btn-action danger">删除</button>
                  </div>
                </div>
                <div v-if="apiKeys.length === 0" class="empty-state p-20">
                  <p>暂无 API Key</p>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== 邮件模板 ===== -->
          <section v-if="tab === 'templates'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">邮件模板</h1>
                <p class="section-desc">保存常用邮件模板，发送时快速选用</p>
              </div>
              <button @click="openCreateTemplate" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                新建模板
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr><th>名称</th><th>主题</th><th>类型</th><th>描述</th><th>创建时间</th><th>操作</th></tr></thead>
                  <tbody>
                    <tr v-for="t in templates" :key="t.id">
                      <td><span class="cell-main">{{ t.name }}</span></td>
                      <td>{{ t.subject || '-' }}</td>
                      <td><span class="badge" :class="t.is_html ? 'badge-info' : ''">{{ t.is_html ? 'HTML' : '纯文本' }}</span></td>
                      <td class="text-muted">{{ t.description || '-' }}</td>
                      <td>{{ formatDate(t.created_at) }}</td>
                      <td>
                        <div class="action-group">
                          <button @click="openEditTemplate(t)" class="btn-action">编辑</button>
                          <button @click="deleteTemplate(t.id)" class="btn-action danger">删除</button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="templates.length === 0">
                      <td colspan="6" class="empty-cell"><div class="empty-state"><p>暂无模板</p></div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          <!-- ===== 收件人分组 ===== -->
          <section v-if="tab === 'recipients'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">收件人分组</h1>
                <p class="section-desc">管理收件人分组，便于批量发送</p>
              </div>
              <button @click="openCreateGroup" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                新建分组
              </button>
            </div>
            <div class="stats-grid recipient-overview">
              <article class="stat-card">
                <div class="stat-icon blue">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V7Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
                </div>
                <div class="stat-num">{{ recipientGroups.length }}</div>
                <div class="stat-label">分组总数</div>
              </article>
              <article class="stat-card">
                <div class="stat-icon green">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M16 21v-2a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4v2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><circle cx="9.5" cy="7" r="4" stroke="currentColor" stroke-width="1.8"/><path d="M20 8v6M17 11h6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ totalRecipientCount }}</div>
                <div class="stat-label">收件人总量</div>
              </article>
              <article class="stat-card">
                <div class="stat-icon orange">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/></svg>
                </div>
                <div class="stat-num">{{ selectedRecipientCount }}</div>
                <div class="stat-label">当前分组人数</div>
              </article>
            </div>
            <div class="recipient-shell">
              <div class="card recipient-panel recipient-group-panel">
                <div class="recipient-panel-head">
                  <div>
                    <h3 class="recipient-panel-title">分组目录</h3>
                    <p class="recipient-panel-desc">先选择分组，再在右侧批量导入、维护名单或清理黑名单成员。</p>
                  </div>
                  <span class="badge badge-info">{{ recipientGroups.length }} 组</span>
                </div>
                <div v-if="recipientGroups.length > 0" class="recipient-group-list">
                  <article
                    v-for="g in recipientGroups"
                    :key="g.id"
                    class="recipient-group-card"
                    :class="{ active: currentGroupId === g.id }"
                    tabindex="0"
                    role="button"
                    @click="loadRecipients(g.id)"
                    @keydown.enter.prevent="loadRecipients(g.id)"
                    @keydown.space.prevent="loadRecipients(g.id)"
                  >
                    <div class="recipient-group-card-head">
                      <div class="recipient-group-card-title">
                        <h3>{{ g.name }}</h3>
                        <p>{{ g.description || '未填写分组描述，适合在这里补充用途或来源。' }}</p>
                      </div>
                      <div class="recipient-group-card-badges">
                        <span class="badge" :class="currentGroupId === g.id ? 'badge-info' : 'badge-muted'">{{ g.count || 0 }} 人</span>
                        <span v-if="currentGroupId === g.id" class="badge badge-info">当前</span>
                      </div>
                    </div>
                    <div class="recipient-group-card-summary">
                      <span class="recipient-summary-pill">创建于 {{ formatDate(g.created_at) }}</span>
                      <span class="recipient-summary-pill">{{ g.description ? '已填写描述' : '未填描述' }}</span>
                    </div>
                    <div class="recipient-group-card-actions">
                      <button @click.stop="loadRecipients(g.id)" class="btn-action">查看</button>
                      <button @click.stop="openEditGroup(g)" class="btn-action">编辑</button>
                      <button @click.stop="deleteGroup(g.id)" class="btn-action danger">删除</button>
                    </div>
                  </article>
                </div>
                <div v-else class="recipient-panel-empty">
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none"><path d="M4 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V7Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
                  <h3>还没有收件人分组</h3>
                  <p>先创建分组，再把目标用户导入进去。后续做群发时可以直接按分组选择。</p>
                  <button @click="openCreateGroup" class="btn-primary">新建第一个分组</button>
                </div>
              </div>

              <div class="card recipient-panel recipient-workspace">
                <div class="recipient-workspace-hero" :class="{ inactive: !currentRecipientGroup }">
                  <div class="recipient-workspace-copy">
                    <span class="recipient-eyebrow">收件人工作区</span>
                    <h3>{{ currentRecipientGroup ? currentRecipientGroup.name : '选择一个分组开始管理名单' }}</h3>
                    <p>
                      {{ currentRecipientGroup
                        ? (currentRecipientGroup.description || '当前分组未填写描述，你可以继续导入邮箱、补充联系人名称，或删除无效地址。')
                        : '左侧先选定一个分组，右侧会切换到该分组的收件人列表，并开放导入和新增操作。' }}
                    </p>
                  </div>
                  <div class="recipient-kpi-grid">
                    <div class="recipient-kpi">
                      <span>分组人数</span>
                      <strong>{{ selectedRecipientCount }}</strong>
                      <small>{{ currentRecipientGroup ? '已加载名单' : '等待选择' }}</small>
                    </div>
                    <div class="recipient-kpi">
                      <span>正常状态</span>
                      <strong>{{ activeRecipientCount }}</strong>
                      <small>可正常发送</small>
                    </div>
                    <div class="recipient-kpi">
                      <span>黑名单</span>
                      <strong>{{ blockedRecipientCount }}</strong>
                      <small>已限制发送</small>
                    </div>
                  </div>
                </div>

                <div class="recipient-toolbar">
                  <div class="recipient-toolbar-copy">
                    <span class="badge" :class="currentRecipientGroup ? 'badge-info' : 'badge-muted'">{{ currentRecipientGroup ? '已选择分组' : '未选择分组' }}</span>
                    <p>
                      {{ currentRecipientGroup
                        ? `当前正在管理 ${currentRecipientGroup.name} 的收件人名单。`
                        : '请选择左侧分组后再执行批量导入或新增收件人。' }}
                    </p>
                  </div>
                  <div class="recipient-toolbar-actions">
                    <button @click="openBatchImport" class="btn-outline" :disabled="!currentGroupId">批量导入</button>
                    <button @click="openCreateRecipient" class="btn-primary" :disabled="!currentGroupId">添加收件人</button>
                  </div>
                </div>

                <template v-if="currentGroupId">
                  <div class="table-wrap">
                    <table>
                      <thead><tr><th>收件人</th><th>状态</th><th>操作</th></tr></thead>
                      <tbody>
                        <tr v-for="r in recipients" :key="r.id">
                          <td>
                            <div class="recipient-entry">
                              <span class="cell-main">{{ r.email }}</span>
                              <div class="recipient-entry-sub">
                                <span>{{ r.name || '未填写名称' }}</span>
                                <span>创建于 {{ formatDate(r.created_at) }}</span>
                              </div>
                            </div>
                          </td>
                          <td><span class="badge" :class="r.status === 'active' ? 'badge-success' : 'badge-danger'">{{ r.status === 'active' ? '正常' : '黑名单' }}</span></td>
                          <td><button @click="deleteRecipient(r.id)" class="btn-action danger">删除</button></td>
                        </tr>
                        <tr v-if="recipients.length === 0">
                          <td colspan="3" class="empty-cell">
                            <div class="empty-state">
                              <p>当前分组还没有收件人，先导入一批邮箱或手动添加一个联系人。</p>
                            </div>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <div class="mobile-list recipient-mobile-list">
                    <div v-for="r in recipients" :key="r.id" class="mobile-item recipient-mobile-item">
                      <div class="mobile-item-head">
                        <span class="cell-main">{{ r.email }}</span>
                        <span class="badge" :class="r.status === 'active' ? 'badge-success' : 'badge-danger'">{{ r.status === 'active' ? '正常' : '黑名单' }}</span>
                      </div>
                      <div class="recipient-entry-sub recipient-mobile-sub">
                        <span>{{ r.name || '未填写名称' }}</span>
                        <span>创建于 {{ formatDate(r.created_at) }}</span>
                      </div>
                      <div class="action-group mt-12">
                        <button @click="deleteRecipient(r.id)" class="btn-action danger">删除</button>
                      </div>
                    </div>
                    <div v-if="recipients.length === 0" class="recipient-panel-empty recipient-mobile-empty">
                      <h3>当前分组还没有收件人</h3>
                      <p>可以先批量导入邮箱列表，或者直接添加一个联系人用于测试发送。</p>
                    </div>
                  </div>
                </template>

                <div v-else class="recipient-panel-empty recipient-workspace-empty">
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none"><path d="M16 21v-2a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4v2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><circle cx="9.5" cy="7" r="4" stroke="currentColor" stroke-width="1.8"/><path d="M20 8v6M17 11h6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                  <h3>右侧工作区暂未激活</h3>
                  <p>选中一个分组后，这里会展示该分组的收件人名单、状态统计以及导入操作入口。</p>
                  <div class="recipient-empty-actions">
                    <button @click="openCreateGroup" class="btn-outline">创建分组</button>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== Webhook ===== -->
          <section v-if="tab === 'webhooks'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">Webhook 回调</h1>
                <p class="section-desc">配置事件回调，邮件发送/打开时自动通知</p>
              </div>
              <button @click="openCreateWebhook" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                新建 Webhook
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr><th>名称</th><th>URL</th><th>事件</th><th>状态</th><th>操作</th></tr></thead>
                  <tbody>
                    <tr v-for="w in webhooks" :key="w.id">
                      <td><span class="cell-main">{{ w.name }}</span></td>
                      <td class="text-muted text-sm" style="max-width:200px">{{ w.url }}</td>
                      <td><span class="badge badge-info">{{ w.events ? JSON.parse(w.events).length : 0 }} 个</span></td>
                      <td><span class="badge" :class="w.enabled ? 'badge-success' : ''">{{ w.enabled ? '启用' : '禁用' }}</span></td>
                      <td>
                        <div class="action-group">
                          <button @click="toggleWebhook(w.id)" class="btn-action">{{ w.enabled ? '禁用' : '启用' }}</button>
                          <button @click="testWebhook(w.id)" class="btn-action">测试</button>
                          <button @click="deleteWebhook(w.id)" class="btn-action danger">删除</button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="webhooks.length === 0">
                      <td colspan="5" class="empty-cell"><div class="empty-state"><p>暂无 Webhook</p></div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          <!-- ===== 黑名单 ===== -->
          <section v-if="tab === 'blacklist'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">黑名单</h1>
                <p class="section-desc">禁止向这些邮箱发送邮件</p>
              </div>
              <button @click="openCreateBlacklist" class="btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                添加黑名单
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr><th>邮箱</th><th>原因</th><th>添加时间</th><th>操作</th></tr></thead>
                  <tbody>
                    <tr v-for="b in blacklist" :key="b.id">
                      <td><span class="cell-main">{{ b.email }}</span></td>
                      <td class="text-muted">{{ b.reason || '-' }}</td>
                      <td>{{ formatDate(b.created_at) }}</td>
                      <td><button @click="deleteBlacklist(b.id)" class="btn-action danger">移除</button></td>
                    </tr>
                    <tr v-if="blacklist.length === 0">
                      <td colspan="4" class="empty-cell"><div class="empty-state"><p>暂无黑名单</p></div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          <!-- ===== 发送日志 ===== -->
          <section v-if="tab === 'logs'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">发送日志</h1>
                <p class="section-desc">查看所有邮件发送记录与状态详情</p>
              </div>
              <button @click="loadLogs" class="btn-outline">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                刷新
              </button>
            </div>
            <div class="card">
              <div class="table-wrap">
                <table>
                  <thead><tr><th>收件人</th><th>主题</th><th>状态</th><th>错误信息</th><th>发送时间</th></tr></thead>
                  <tbody>
                    <tr v-for="log in logs" :key="log.id">
                      <td>{{ log.to_email }}</td>
                      <td class="text-truncate" style="max-width:200px">{{ log.subject || '-' }}</td>
                      <td>
                        <span :class="['badge', log.status === 'success' ? 'badge-success' : 'badge-danger']">
                          <span class="badge-dot"></span>{{ log.status === 'success' ? '成功' : '失败' }}
                        </span>
                      </td>
                      <td class="text-muted text-sm">{{ log.error_message || '-' }}</td>
                      <td>{{ formatDate(log.created_at) }}</td>
                    </tr>
                    <tr v-if="logs.length === 0">
                      <td colspan="5" class="empty-cell">
                        <div class="empty-state">
                          <svg width="44" height="44" viewBox="0 0 24 24" fill="none"><path d="M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
                          <p>暂无发送记录</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-if="logTotal > logPageSize" class="pagination">
                <button :disabled="logPage <= 1" @click="logPage--; loadLogs()" class="page-btn">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                </button>
                <span class="page-info">第 {{ logPage }} / {{ Math.ceil(logTotal / logPageSize) }} 页 · 共 {{ logTotal }} 条</span>
                <button :disabled="logPage >= Math.ceil(logTotal / logPageSize)" @click="logPage++; loadLogs()" class="page-btn">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M9 18l6-6-6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                </button>
              </div>
              <div class="mobile-list">
                <div v-for="log in logs" :key="log.id" class="mobile-item">
                  <div class="mobile-item-head">
                    <span class="cell-main">{{ log.to_email }}</span>
                    <span :class="['badge', log.status === 'success' ? 'badge-success' : 'badge-danger']">
                      <span class="badge-dot"></span>{{ log.status === 'success' ? '成功' : '失败' }}
                    </span>
                  </div>
                  <div class="mobile-item-meta">
                    <span>{{ log.subject || '（无主题）' }}</span>
                    <span>{{ formatDate(log.created_at) }}</span>
                  </div>
                  <p v-if="log.error_message" class="text-sm text-danger mt-4">{{ log.error_message }}</p>
                </div>
                <div v-if="logs.length === 0" class="empty-state p-20"><p>暂无发送记录</p></div>
              </div>
            </div>
          </section>

          <!-- ===== 统计 ===== -->
          <section v-if="tab === 'stats'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">数据统计</h1>
                <p class="section-desc">邮件发送情况概览与实时数据</p>
              </div>
              <button @click="loadStats" class="btn-outline">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10M23 14l-4.64 4.36A9 9 0 0 1 3.51 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                刷新
              </button>
            </div>
            <div class="stats-grid">
              <div class="stat-card">
                <div class="stat-icon blue">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8"/></svg>
                </div>
                <div class="stat-num">{{ stats.total_sent || 0 }}</div>
                <div class="stat-label">累计发送</div>
              </div>
              <div class="stat-card">
                <div class="stat-icon green">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.success || 0 }}</div>
                <div class="stat-label">发送成功</div>
              </div>
              <div class="stat-card">
                <div class="stat-icon red">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.failed || 0 }}</div>
                <div class="stat-label">发送失败</div>
              </div>
              <div class="stat-card">
                <div class="stat-icon cyan">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.8"/><path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ stats.today_sent || 0 }}</div>
                <div class="stat-label">今日发送</div>
              </div>
              <div class="stat-card">
                <div class="stat-icon purple">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M18 20V10M12 20V4M6 20v-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
                </div>
                <div class="stat-num">{{ (stats.success_rate || 0).toFixed(1) }}<span class="stat-unit">%</span></div>
                <div class="stat-label">成功率</div>
              </div>
              <div class="stat-card">
                <div class="stat-icon orange">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" stroke-width="1.8"/><circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.8"/></svg>
                </div>
                <div class="stat-num">{{ stats.opened || 0 }}</div>
                <div class="stat-label">已打开 <span v-if="stats.open_rate">({{ stats.open_rate.toFixed(1) }}%)</span></div>
              </div>
              <div class="stat-card">
                <div class="stat-icon teal">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" stroke="currentColor" stroke-width="1.8"/><polyline points="15 3 21 3 21 9" stroke="currentColor" stroke-width="1.8"/><line x1="10" y1="14" x2="21" y2="3" stroke="currentColor" stroke-width="1.8"/></svg>
                </div>
                <div class="stat-num">{{ stats.clicked || 0 }}</div>
                <div class="stat-label">已点击 <span v-if="stats.click_rate">({{ stats.click_rate.toFixed(1) }}%)</span></div>
              </div>
            </div>
            <div v-if="smtpAccounts.length > 0" class="card">
              <div class="card-head">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8"/></svg>
                账号使用情况
              </div>
              <div class="account-usage-list">
                <div v-for="acc in smtpAccounts" :key="acc.id" class="account-usage-item">
                  <div class="account-usage-head">
                    <span class="cell-main">{{ acc.email }}</span>
                    <span :class="['badge', acc.status === 'active' ? 'badge-success' : 'badge-muted']">
                      <span class="badge-dot"></span>{{ acc.status === 'active' ? '启用' : '禁用' }}
                    </span>
                  </div>
                  <div class="account-usage-bar" v-if="acc.daily_limit">
                    <div class="bar-track">
                      <div class="bar-fill" :style="{ width: Math.min(100, (acc.daily_used / acc.daily_limit) * 100) + '%', background: acc.daily_used / acc.daily_limit > 0.8 ? 'var(--red)' : 'var(--gradient-blue)' }"></div>
                    </div>
                    <span class="bar-label">{{ acc.daily_used }} / {{ acc.daily_limit }}</span>
                  </div>
                  <div v-else class="account-usage-bar">
                    <span class="text-muted text-sm">无限制 · 今日已用 {{ acc.daily_used }}</span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- ===== API 文档 ===== -->
          <section v-if="tab === 'docs'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">API 对接文档</h1>
                <p class="section-desc">HTTP API 接口参考与多语言 SDK 示例</p>
              </div>
              <div class="doc-head-right">
                <div class="base-url-pill">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.8"/><path d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10A15.3 15.3 0 018 12a15.3 15.3 0 014-10z" stroke="currentColor" stroke-width="1.8"/></svg>
                  Base URL: <code>{{ baseUrl }}</code>
                </div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">01</span> 认证方式</div>
              <div class="doc-grid">
                <div class="doc-card">
                  <div class="doc-card-head">
                    <span class="method-tag post">POST</span>
                    <code class="path-tag">/api/v1/auth/login</code>
                  </div>
                  <p class="doc-desc">使用用户名密码获取 JWT Token，用于管理界面 API 调用。</p>
                  <div class="code-block-wrap">
                    <div class="code-block-label">请求示例</div>
                    <div class="copy-wrap">
                      <pre class="code-block" v-text="curlLogin"></pre>
                      <button class="copy-btn" :class="{copied: copiedKey==='curlLogin'}" @click="copyText(curlLogin,'curlLogin')">{{ copiedKey==='curlLogin' ? '✓ 已复制' : '复制' }}</button>
                    </div>
                  </div>
                  <div class="code-block-wrap">
                    <div class="code-block-label">响应示例</div>
                    <div class="copy-wrap">
                      <pre class="code-block">{ "token": "eyJhbGciOiJIUzI1NiIsInR5...", "username": "admin" }</pre>
                      <button class="copy-btn" :class="{copied: copiedKey==='respLogin'}" @click="copyText('{ &quot;token&quot;: &quot;eyJhbGciOiJIUzI1NiIsInR5...&quot;, &quot;username&quot;: &quot;admin&quot; }','respLogin')">{{ copiedKey==='respLogin' ? '✓ 已复制' : '复制' }}</button>
                    </div>
                  </div>
                </div>
                <div class="doc-card">
                  <div class="doc-card-head">
                    <span class="method-tag post">POST</span>
                    <code class="path-tag">/api/v1/auth/change-password</code>
                    <span class="auth-required-tag">🔒 需要 Token</span>
                  </div>
                  <p class="doc-desc">修改登录密码，修改成功后需要重新登录。</p>
                  <div class="code-block-wrap">
                    <div class="code-block-label">请求示例</div>
                    <div class="copy-wrap">
                      <pre class="code-block" v-text="curlChangePassword"></pre>
                      <button class="copy-btn" :class="{copied: copiedKey==='curlChangePwd'}" @click="copyText(curlChangePassword,'curlChangePwd')">{{ copiedKey==='curlChangePwd' ? '✓ 已复制' : '复制' }}</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">02</span> 发送邮件（核心接口）</div>
              <div class="doc-card featured">
                <div class="doc-card-head">
                  <span class="method-tag post">POST</span>
                  <code class="path-tag">/api/v1/send</code>
                  <span class="auth-required-tag">🔑 Token 或 API Key</span>
                </div>
                <p class="doc-desc">统一发信接口，系统自动轮询可用 SMTP 账号发送，支持 HTML 邮件。</p>
                <div class="doc-grid-2">
                  <div class="code-block-wrap">
                    <div class="code-block-label">使用 API Key（推荐）</div>
                    <div class="copy-wrap">
                      <pre class="code-block" v-text="curlSendApiKey"></pre>
                      <button class="copy-btn" :class="{copied: copiedKey==='curlSendKey'}" @click="copyText(curlSendApiKey,'curlSendKey')">{{ copiedKey==='curlSendKey' ? '✓ 已复制' : '复制' }}</button>
                    </div>
                  </div>
                  <div class="code-block-wrap">
                    <div class="code-block-label">发送 HTML 邮件</div>
                    <div class="copy-wrap">
                      <pre class="code-block" v-text="curlSendHtml"></pre>
                      <button class="copy-btn" :class="{copied: copiedKey==='curlSendHtml'}" @click="copyText(curlSendHtml,'curlSendHtml')">{{ copiedKey==='curlSendHtml' ? '✓ 已复制' : '复制' }}</button>
                    </div>
                  </div>
                </div>
                <div class="params-table-wrap">
                  <div class="code-block-label">请求参数</div>
                  <table class="params-table">
                    <thead><tr><th>参数</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
                    <tbody>
                      <tr><td><code>to</code></td><td>string</td><td><span class="req-yes">是</span></td><td>收件人邮箱</td></tr>
                      <tr><td><code>subject</code></td><td>string</td><td><span class="req-yes">是</span></td><td>邮件主题</td></tr>
                      <tr><td><code>body</code></td><td>string</td><td><span class="req-yes">是</span></td><td>邮件正文，is_html=true 时支持 HTML</td></tr>
                      <tr><td><code>is_html</code></td><td>bool</td><td><span class="req-no">否</span></td><td>true 时以 HTML 格式发送（默认 false）</td></tr>
                      <tr><td><code>from_name</code></td><td>string</td><td><span class="req-no">否</span></td><td>发件人显示名称</td></tr>
                      <tr><td><code>cc</code></td><td>string[]</td><td><span class="req-no">否</span></td><td>抄送邮箱列表，收件人可见</td></tr>
                      <tr><td><code>bcc</code></td><td>string[]</td><td><span class="req-no">否</span></td><td>密送邮箱列表，收件人不可见</td></tr>
                    </tbody>
                  </table>
                </div>
                <div class="code-block-wrap">
                  <div class="code-block-label">成功响应</div>
                  <div class="copy-wrap">
                    <pre class="code-block">{ "success": true, "message": "Email sent successfully", "used_smtp": "user***@gmail.com" }</pre>
                    <button class="copy-btn" :class="{copied: copiedKey==='respSend'}" @click="copyText('{ &quot;success&quot;: true, &quot;message&quot;: &quot;Email sent successfully&quot;, &quot;used_smtp&quot;: &quot;user***@gmail.com&quot; }','respSend')">{{ copiedKey==='respSend' ? '✓ 已复制' : '复制' }}</button>
                  </div>
                </div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">03</span> 代码示例</div>
              <div class="tabs-simple">
                <button :class="['tab-simple', { active: codeTab === 'python' }]" @click="codeTab = 'python'">Python</button>
                <button :class="['tab-simple', { active: codeTab === 'nodejs' }]" @click="codeTab = 'nodejs'">Node.js</button>
                <button :class="['tab-simple', { active: codeTab === 'php' }]" @click="codeTab = 'php'">PHP</button>
                <button :class="['tab-simple', { active: codeTab === 'go' }]" @click="codeTab = 'go'">Go</button>
              </div>
              <div class="code-block-wrap">
                <div class="copy-wrap">
                  <pre v-if="codeTab === 'python'" class="code-block" v-text="codeExamplePython"></pre>
                  <pre v-if="codeTab === 'nodejs'" class="code-block" v-text="codeExampleNodejs"></pre>
                  <pre v-if="codeTab === 'php'" class="code-block" v-text="codeExamplePhp"></pre>
                  <pre v-if="codeTab === 'go'" class="code-block" v-text="codeExampleGo"></pre>
                  <button class="copy-btn" :class="{copied: copiedKey==='codeExample'}" @click="copyText(currentCodeExample,'codeExample')">{{ copiedKey==='codeExample' ? '✓ 已复制' : '复制' }}</button>
                </div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">04</span> SMTP 账号管理 <span class="doc-auth-note">（需要 Bearer Token）</span></div>
              <div class="doc-list">
                <div class="doc-list-item"><span class="method-tag get">GET</span><code class="path-tag">/api/v1/smtp-accounts</code><span class="doc-list-desc">获取所有 SMTP 账号列表</span></div>
                <div class="doc-list-item"><span class="method-tag post">POST</span><code class="path-tag">/api/v1/smtp-accounts</code><span class="doc-list-desc">添加新账号：email、password、smtp_host、smtp_port、daily_limit</span></div>
                <div class="doc-list-item"><span class="method-tag put">PUT</span><code class="path-tag">/api/v1/smtp-accounts/:id</code><span class="doc-list-desc">更新账号信息</span></div>
                <div class="doc-list-item"><span class="method-tag del">DELETE</span><code class="path-tag">/api/v1/smtp-accounts/:id</code><span class="doc-list-desc">删除账号</span></div>
                <div class="doc-list-item"><span class="method-tag post">POST</span><code class="path-tag">/api/v1/smtp-accounts/:id/test</code><span class="doc-list-desc">测试 SMTP 连通性（不发送邮件）</span></div>
                <div class="doc-list-item"><span class="method-tag post">POST</span><code class="path-tag">/api/v1/smtp-accounts/:id/test-send</code><span class="doc-list-desc">发送测试邮件，请求体：{"to":"email"}</span></div>
                <div class="doc-list-item"><span class="method-tag post">POST</span><code class="path-tag">/api/v1/smtp-accounts/:id/toggle</code><span class="doc-list-desc">切换账号启用/禁用状态</span></div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">05</span> API Key 管理 <span class="doc-auth-note">（需要 Bearer Token）</span></div>
              <div class="doc-list">
                <div class="doc-list-item"><span class="method-tag get">GET</span><code class="path-tag">/api/v1/api-keys</code><span class="doc-list-desc">获取所有 API Key 列表（完整 Key 仅创建时展示一次）</span></div>
                <div class="doc-list-item"><span class="method-tag post">POST</span><code class="path-tag">/api/v1/api-keys</code><span class="doc-list-desc">创建 Key：请求体 {"name":"my-app"}，响应含完整 Key</span></div>
                <div class="doc-list-item"><span class="method-tag del">DELETE</span><code class="path-tag">/api/v1/api-keys/:id</code><span class="doc-list-desc">删除指定 API Key</span></div>
              </div>
            </div>

            <div class="doc-section">
              <div class="doc-section-title"><span class="doc-num">06</span> 日志与统计 <span class="doc-auth-note">（需要 Bearer Token）</span></div>
              <div class="doc-list">
                <div class="doc-list-item"><span class="method-tag get">GET</span><code class="path-tag">/api/v1/send/logs?page=1&amp;page_size=50</code><span class="doc-list-desc">分页获取发送日志，响应含 logs、total、page</span></div>
                <div class="doc-list-item"><span class="method-tag get">GET</span><code class="path-tag">/api/v1/stats</code><span class="doc-list-desc">统计数据：total_sent、success、failed、today_sent、success_rate</span></div>
              </div>
            </div>
          </section>

          <!-- ===== 系统更新 ===== -->
          <section v-if="tab === 'system'" class="section">
            <div class="section-head">
              <div>
                <h1 class="section-title">系统更新</h1>
                <p class="section-desc">检测并更新至最新版本</p>
              </div>
            </div>

            <div class="card">
              <div class="card-body">
                <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:14px">
                  <div>
                    <div style="font-size:0.8125rem;color:var(--gray-400);font-weight:600;margin-bottom:6px">当前版本</div>
                    <div style="display:flex;align-items:center;gap:10px">
                      <span style="font-size:1.5rem;font-weight:700;color:var(--gray-900);letter-spacing:-0.02em">{{ currentVersion || '加载中...' }}</span>
                      <span class="badge badge-info">SMTP Lite</span>
                    </div>
                  </div>
                  <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap">
                    <button class="btn-outline" @click="checkUpdate" :disabled="updateChecking">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                      {{ updateChecking ? '检测中…' : '检测更新' }}
                    </button>
                  </div>
                </div>

                <div v-if="updateStatus === 'latest'" class="alert alert-success" style="margin-top:16px">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                  已是最新版本，无需更新
                </div>
                <div v-if="updateStatus === 'available'" class="alert alert-warn" style="margin-top:16px;flex-wrap:wrap;gap:10px">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M12 9v4M12 17h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                  <span>发现新版本 <strong>{{ latestVersion }}</strong></span>
                  <a href="https://github.com/DoBestone/smtp-lite/releases" target="_blank" style="color:inherit;font-weight:600;text-decoration:underline">查看更新日志</a>
                  <button class="btn-primary" style="margin-left:auto" @click="doUpdate">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    一键更新
                  </button>
                </div>
                <div v-if="updateStatus === 'error'" class="alert alert-error" style="margin-top:16px">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                  检测更新失败，请稍后重试
                </div>
              </div>
            </div>
          </section>

        </div>
      </main>
    </div>

    <!-- ========== 添加/编辑 SMTP 弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showSmtpModal" class="modal-overlay" @click.self="showSmtpModal = false">
        <div class="modal-box">
          <div class="modal-head">
            <h3>{{ editingSmtp ? '编辑 SMTP 账号' : '添加 SMTP 账号' }}</h3>
            <button class="modal-close" @click="showSmtpModal = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <form @submit.prevent="saveSmtpAccount">
            <div class="field">
              <label>邮箱地址 <span class="required">*</span></label>
              <input v-model="smtpForm.email" placeholder="example@gmail.com" type="email" required />
            </div>
            <div class="field">
              <label>{{ editingSmtp ? '新密码（留空不修改）' : '密码 / 授权码 *' }}</label>
              <input v-model="smtpForm.password" :placeholder="editingSmtp ? '留空则不修改密码' : '应用密码或授权码'" type="password" :required="!editingSmtp" />
            </div>
            <div class="field">
              <label>SMTP 服务器 <span class="required">*</span></label>
              <div class="input-with-presets">
                <input v-model="smtpForm.smtp_host" placeholder="如 smtp.gmail.com" required />
                <div class="preset-btns">
                  <button type="button" class="preset-btn" @click="applyPreset('gmail')">Gmail</button>
                  <button type="button" class="preset-btn" @click="applyPreset('outlook')">Outlook</button>
                  <button type="button" class="preset-btn" @click="applyPreset('qq')">QQ</button>
                  <button type="button" class="preset-btn" @click="applyPreset('163')">163</button>
                </div>
              </div>
            </div>
            <div class="field-row">
              <div class="field">
                <label>SMTP 端口</label>
                <input v-model.number="smtpForm.smtp_port" placeholder="587" type="number" />
              </div>
              <div class="field">
                <label>每日限额</label>
                <input v-model.number="smtpForm.daily_limit" placeholder="500（留空不限）" type="number" />
              </div>
            </div>
            <div v-if="smtpFormError" class="form-error mt-8">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
              {{ smtpFormError }}
            </div>
            <div class="modal-actions">
              <button type="button" class="btn-ghost" @click="showSmtpModal = false">取消</button>
              <button type="submit" class="btn-primary">{{ editingSmtp ? '保存修改' : '添加账号' }}</button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 发送测试邮件弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showTestSend" class="modal-overlay" @click.self="showTestSend = false">
        <div class="modal-box" style="max-width:400px">
          <div class="modal-head">
            <h3>发送测试邮件</h3>
            <button class="modal-close" @click="showTestSend = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <p class="modal-desc">通过 <strong>{{ testSendAccount && testSendAccount.email }}</strong> 发送一封测试邮件，验证配置是否正常。</p>
          <div class="field mt-12">
            <label>收件人邮箱 <span class="required">*</span></label>
            <input v-model="testSendTo" placeholder="recipient@example.com" type="email" />
          </div>
          <div v-if="testSendResult" :class="['alert', testSendResult.success ? 'alert-success' : 'alert-error']" style="margin-top:10px">
            <svg v-if="testSendResult.success" width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            {{ testSendResult.message || testSendResult.error }}
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showTestSend = false">关闭</button>
            <button class="btn-primary" @click="doTestSend" :disabled="testSendLoading || !testSendTo">
              <span v-if="!testSendLoading">发送测试邮件</span>
              <span v-else class="spinner"></span>
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ========== API Key 展示弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="newKeyInfo" class="modal-overlay" @click.self="newKeyInfo = null">
        <div class="modal-box" style="max-width:440px">
          <div class="modal-head">
            <h3>API Key 已创建</h3>
            <button class="modal-close" @click="newKeyInfo = null">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <div class="alert alert-warn">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" stroke="currentColor" stroke-width="1.8"/><line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><line x1="12" y1="17" x2="12.01" y2="17" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            请立即保存此 Key，它只会显示一次！
          </div>
          <div class="key-display-box">
            <code>{{ newKeyInfo.key }}</code>
            <button @click="copyKey" class="copy-btn" title="复制">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="1.8"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="1.8"/></svg>
            </button>
          </div>
          <div class="modal-actions">
            <button @click="copyKey" class="btn-primary">复制 Key</button>
            <button @click="newKeyInfo = null" class="btn-ghost">关闭</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ========== 修改密码弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showChangePwd" class="modal-overlay" @click.self="showChangePwd = false">
        <div class="modal-box" style="max-width:400px">
          <div class="modal-head">
            <h3>修改登录密码</h3>
            <button class="modal-close" @click="showChangePwd = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <form @submit.prevent="changePassword">
            <div class="field">
              <label>当前密码 <span class="required">*</span></label>
              <input v-model="pwdForm.oldPwd" type="password" placeholder="请输入当前密码" required />
            </div>
            <div class="field">
              <label>新密码 <span class="required">*</span></label>
              <input v-model="pwdForm.newPwd" type="password" placeholder="至少 6 位字符" required minlength="6" />
            </div>
            <div class="field">
              <label>确认新密码 <span class="required">*</span></label>
              <input v-model="pwdForm.confirmPwd" type="password" placeholder="再次输入新密码" required />
            </div>
            <div v-if="pwdError" class="form-error mt-8">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
              {{ pwdError }}
            </div>
            <div v-if="pwdSuccess" class="alert alert-success mt-8">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
              {{ pwdSuccess }}
            </div>
            <div class="modal-actions">
              <button type="button" class="btn-ghost" @click="showChangePwd = false">取消</button>
              <button type="submit" class="btn-primary" :disabled="pwdLoading">
                <span v-if="!pwdLoading">保存修改</span>
                <span v-else class="spinner" style="width:16px;height:16px"></span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 创建 API Key 弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showCreateKey" class="modal-overlay" @click.self="showCreateKey = false">
        <div class="modal-box" style="max-width:380px">
          <div class="modal-head">
            <h3>创建 API Key</h3>
            <button class="modal-close" @click="showCreateKey = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <form @submit.prevent="doCreateApiKey">
            <div class="field">
              <label>Key 名称 <span class="required">*</span></label>
              <input v-model="newKeyName" placeholder="如：my-app、production" type="text" required ref="keyNameInput" />
            </div>
            <div class="modal-actions">
              <button type="button" class="btn-ghost" @click="showCreateKey = false">取消</button>
              <button type="submit" class="btn-primary" :disabled="!newKeyName.trim()">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                创建
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 模板弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showTemplateModal" class="modal-overlay" @click.self="showTemplateModal = false">
        <div class="modal-box" style="max-width:500px">
          <div class="modal-head">
            <h3>{{ editingTemplate ? '编辑模板' : '新建模板' }}</h3>
            <button class="modal-close" @click="showTemplateModal = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <form @submit.prevent="saveTemplate">
            <div class="field"><label>模板名称 <span class="required">*</span></label><input v-model="templateForm.name" placeholder="如：验证码模板" required /></div>
            <div class="field"><label>邮件主题</label><input v-model="templateForm.subject" placeholder="邮件主题" /></div>
            <div class="field"><label>邮件内容 <span class="required">*</span></label><textarea v-model="templateForm.body" rows="6" placeholder="邮件正文..." required></textarea></div>
            <div class="field"><label class="checkbox-label"><input type="checkbox" v-model="templateForm.is_html" /> HTML 格式</label></div>
            <div class="field"><label>描述</label><input v-model="templateForm.description" placeholder="可选描述" /></div>
            <div class="modal-actions">
              <button type="button" class="btn-ghost" @click="showTemplateModal = false">取消</button>
              <button type="submit" class="btn-primary">保存</button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 分组弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showGroupModal" class="modal-overlay" @click.self="showGroupModal = false">
        <div class="modal-box" style="max-width:380px">
          <div class="modal-head"><h3>{{ editingGroup ? '编辑分组' : '新建分组' }}</h3><button class="modal-close" @click="showGroupModal = false"><svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
          <form @submit.prevent="saveGroup">
            <div class="field"><label>分组名称 <span class="required">*</span></label><input v-model="groupForm.name" placeholder="例如：测试用户、订阅会员" required /></div>
            <div class="field"><label>描述</label><input v-model="groupForm.description" placeholder="展示在分组卡片和工作区说明中" /></div>
            <div class="modal-actions"><button type="button" class="btn-ghost" @click="showGroupModal = false">取消</button><button type="submit" class="btn-primary">{{ editingGroup ? '保存修改' : '创建分组' }}</button></div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 收件人弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showRecipientModal" class="modal-overlay" @click.self="showRecipientModal = false">
        <div class="modal-box" style="max-width:380px">
          <div class="modal-head"><h3>添加收件人</h3><button class="modal-close" @click="showRecipientModal = false"><svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
          <form @submit.prevent="saveRecipient">
            <div class="field"><label>邮箱 <span class="required">*</span></label><input v-model="recipientForm.email" type="email" required /></div>
            <div class="field"><label>名称</label><input v-model="recipientForm.name" /></div>
            <div class="modal-actions"><button type="button" class="btn-ghost" @click="showRecipientModal = false">取消</button><button type="submit" class="btn-primary">添加</button></div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 批量导入弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showBatchImport" class="modal-overlay" @click.self="showBatchImport = false">
        <div class="modal-box" style="max-width:450px">
          <div class="modal-head"><h3>批量导入收件人</h3><button class="modal-close" @click="showBatchImport = false"><svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
          <form @submit.prevent="doBatchImport">
            <div class="field"><label>邮箱列表（每行一个或逗号分隔）</label><textarea v-model="batchEmails" rows="8" placeholder="user1@example.com&#10;user2@example.com&#10;..." required></textarea></div>
            <div class="modal-actions"><button type="button" class="btn-ghost" @click="showBatchImport = false">取消</button><button type="submit" class="btn-primary">导入</button></div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== Webhook 弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showWebhookModal" class="modal-overlay" @click.self="showWebhookModal = false">
        <div class="modal-box" style="max-width:450px">
          <div class="modal-head"><h3>新建 Webhook</h3><button class="modal-close" @click="showWebhookModal = false"><svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
          <form @submit.prevent="saveWebhook">
            <div class="field"><label>名称 <span class="required">*</span></label><input v-model="webhookForm.name" required /></div>
            <div class="field"><label>URL <span class="required">*</span></label><input v-model="webhookForm.url" type="url" placeholder="https://your-server.com/webhook" required /></div>
            <div class="field"><label>Secret（可选）</label><input v-model="webhookForm.secret" placeholder="用于签名验证" /></div>
            <div class="field"><label>订阅事件</label>
              <div style="display:flex;flex-wrap:wrap;gap:8px;margin-top:4px">
                <label v-for="e in webhookEvents" :key="e.key" class="checkbox-label" style="margin-right:12px">
                  <input type="checkbox" :value="e.key" v-model="webhookForm.events" /> {{ e.label }}
                </label>
              </div>
            </div>
            <div class="modal-actions"><button type="button" class="btn-ghost" @click="showWebhookModal = false">取消</button><button type="submit" class="btn-primary">创建</button></div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 黑名单弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="showBlacklistModal" class="modal-overlay" @click.self="showBlacklistModal = false">
        <div class="modal-box" style="max-width:380px">
          <div class="modal-head"><h3>添加黑名单</h3><button class="modal-close" @click="showBlacklistModal = false"><svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button></div>
          <form @submit.prevent="saveBlacklist">
            <div class="field"><label>邮箱 <span class="required">*</span></label><input v-model="blacklistForm.email" type="email" required /></div>
            <div class="field"><label>原因</label><input v-model="blacklistForm.reason" placeholder="可选" /></div>
            <div class="modal-actions"><button type="button" class="btn-ghost" @click="showBlacklistModal = false">取消</button><button type="submit" class="btn-primary">添加</button></div>
          </form>
        </div>
      </div>
    </transition>

    <!-- ========== 确认弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="confirmDialog.show" class="modal-overlay" @click.self="confirmDialog.show = false">
        <div class="modal-box" style="max-width:360px">
          <div class="modal-head">
            <h3>{{ confirmDialog.title }}</h3>
            <button class="modal-close" @click="confirmDialog.show = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
            </button>
          </div>
          <p class="modal-desc" style="margin-bottom:0">{{ confirmDialog.msg }}</p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="confirmDialog.show = false">取消</button>
            <button class="btn-danger" @click="confirmDialog.onConfirm(); confirmDialog.show = false">
              {{ confirmDialog.confirmText || '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ========== 一键更新进度弹窗 ========== -->
    <transition name="modal-fade">
      <div v-if="updateProgress" class="modal-overlay update-overlay">
        <div class="modal-box update-modal">
          <!-- 更新中 -->
          <template v-if="updateProgress === 'updating'">
            <div class="update-spinner"></div>
            <h3 class="update-modal-title">正在更新...</h3>
            <p class="update-modal-desc">正在拉取代码并重新编译，服务将在完成后自动重启，请稍候。</p>
            <div class="update-steps">
              <div class="update-step" :class="updateStep >= 1 ? 'done' : 'pending'">① git pull</div>
              <div class="update-step" :class="updateStep >= 2 ? 'done' : 'pending'">② go build</div>
              <div class="update-step" :class="updateStep >= 3 ? 'done' : 'pending'">③ 重启服务</div>
            </div>
          </template>
          <!-- 成功 -->
          <template v-else-if="updateProgress === 'done'">
            <div class="update-icon-wrap success">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </div>
            <h3 class="update-modal-title">更新成功！</h3>
            <p class="update-modal-desc">已更新至 <strong>{{ latestVersion }}</strong>，页面即将刷新。</p>
            <button class="btn-primary" style="margin-top:16px" @click="reloadPage()">立即刷新</button>
          </template>
          <!-- 失败 -->
          <template v-else-if="updateProgress === 'error' || updateProgress === 'timeout'">
            <div class="update-icon-wrap error">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"/></svg>
            </div>
            <h3 class="update-modal-title">{{ updateProgress === 'timeout' ? '更新超时' : '更新失败' }}</h3>
            <p class="update-modal-desc">请检查服务器日志，确认 git 和 go 命令可用后手动重试。</p>
            <button class="btn-ghost" style="margin-top:16px" @click="updateProgress = ''">关闭</button>
          </template>
        </div>
      </div>
    </transition>

    <!-- ========== Toast ========== -->
    <transition name="toast-fade">
      <div v-if="toast.show" :class="['toast', 'toast-' + toast.type]">
        <svg v-if="toast.type === 'success'" width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        <svg v-else-if="toast.type === 'error'" width="15" height="15" viewBox="0 0 24 24" fill="none"><path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"/></svg>
        <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        {{ toast.msg }}
      </div>
    </transition>
  </div>
</template>

<script>
import axios from 'axios'

const API = '/api/v1'
const SMTP_PRESETS = {
  gmail:   { host: 'smtp.gmail.com',     port: 587 },
  outlook: { host: 'smtp.office365.com', port: 587 },
  qq:      { host: 'smtp.qq.com',        port: 465 },
  '163':   { host: 'smtp.163.com',       port: 465 },
}

export default {
  data() {
    return {
      isLoggedIn: !!localStorage.getItem('token'),
      loginForm: { username: '', password: '' },
      loginLoading: false,
      loginError: '',
      tab: 'smtp',
      sidebarOpen: false,
      smtpAccounts: [],
      apiKeys: [],
      logs: [],
      logPage: 1,
      logPageSize: 50,
      logTotal: 0,
      stats: {},
      showSmtpModal: false,
      editingSmtp: null,
      smtpForm: { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 },
      smtpFormError: '',
      showTestSend: false,
      testSendAccount: null,
      testSendTo: '',
      testSendLoading: false,
      testSendResult: null,
      testingId: null,
      newKeyInfo: null,
      showCreateKey: false,
      newKeyName: '',
      confirmDialog: { show: false, title: '', msg: '', confirmText: '', onConfirm: () => {} },
      showChangePwd: false,
      pwdForm: { oldPwd: '', newPwd: '', confirmPwd: '' },
      pwdError: '',
      pwdSuccess: '',
      pwdLoading: false,
      codeTab: 'python',
      copiedKey: '',
      currentVersion: '',
      latestVersion: '',
      updateStatus: '',
      updateChecking: false,
      updateProgress: '',  // '' | 'updating' | 'done' | 'error' | 'timeout'
      updateStep: 0,
      toast: { show: false, msg: '', type: 'success' },
      // 模板相关
      templates: [],
      showTemplateModal: false,
      editingTemplate: null,
      templateForm: { name: '', subject: '', body: '', is_html: true, description: '' },
      // 收件人相关
      recipientGroups: [],
      recipients: [],
      currentGroupId: '',
      showGroupModal: false,
      editingGroup: null,
      groupForm: { name: '', description: '' },
      showRecipientModal: false,
      recipientForm: { email: '', name: '' },
      showBatchImport: false,
      batchEmails: '',
      // Webhook 相关
      webhooks: [],
      showWebhookModal: false,
      webhookForm: { name: '', url: '', secret: '', events: [] },
      webhookEvents: [
        { key: 'send_success', label: '发送成功' },
        { key: 'send_failed', label: '发送失败' },
        { key: 'opened', label: '邮件打开' },
        { key: 'clicked', label: '链接点击' },
      ],
      // 黑名单相关
      blacklist: [],
      showBlacklistModal: false,
      blacklistForm: { email: '', reason: '' },
      // 队列状态
      queueStats: {},
      // 发送邮件
      sendMode: 'single', // single / batch / scheduled
      sendForm: {
        to: '',
        subject: '',
        body: '',
        is_html: false,
        from_name: '',
        cc: '',
        bcc: '',
        track_enabled: false,
      },
      sendRecipientGroupId: '',
      sendRecipientOptions: [],
      sendRecipientSelection: '',
      batchEmailsList: '',
      scheduledTime: '',
      sendLoading: false,
      sendResult: null,
      navItems: [
        { key: 'smtp', label: 'SMTP 账号', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><rect x="2" y="4" width="20" height="16" rx="3" stroke="currentColor" stroke-width="1.8"/><path d="M2 8l10 6 10-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>' },
        { key: 'send', label: '发送邮件', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
        { key: 'keys', label: 'API Key', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>' },
        { key: 'templates', label: '模板', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5"/><polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="1.5"/></svg>' },
        { key: 'recipients', label: '收件人', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" stroke="currentColor" stroke-width="1.5"/><circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="1.5"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" stroke="currentColor" stroke-width="1.5"/></svg>' },
        { key: 'logs', label: '发送日志', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M9 11l3 3L22 4M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>' },
        { key: 'stats', label: '统计', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M18 20V10M12 20V4M6 20v-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>' },
        { key: 'webhooks', label: 'Webhook', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" stroke="currentColor" stroke-width="1.5"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" stroke="currentColor" stroke-width="1.5"/></svg>' },
        { key: 'blacklist', label: '黑名单', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07" stroke="currentColor" stroke-width="1.5"/></svg>' },
        { key: 'docs', label: 'API 文档', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5"/><polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="1.5"/><line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>' },
        { key: 'system', label: '系统更新', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
      ]
    }
  },
  mounted() {
    const token = localStorage.getItem('token')
    if (token) { this.isLoggedIn = true; this.loadData(); this.checkUpdate() }
    // 从 URL hash 恢复当前 tab
    const validTabs = this.navItems.map(i => i.key)
    const hash = window.location.hash.replace('#', '')
    if (hash && validTabs.includes(hash)) {
      this.switchTab(hash)
    }
    // 监听浏览器前进/后退
    window.addEventListener('popstate', () => {
      const h = window.location.hash.replace('#', '')
      if (h && validTabs.includes(h) && h !== this.tab) {
        this.tab = h
      }
    })
  },
  computed: {
    baseUrl() {
      return window.location.origin
    },
    curlLogin() {
      return `curl -X POST ${this.baseUrl}/api/v1/auth/login \\\n  -H "Content-Type: application/json" \\\n  -d '{"username":"admin","password":"admin123"}'`
    },
    curlChangePassword() {
      return `curl -X POST ${this.baseUrl}/api/v1/auth/change-password \\\n  -H "Authorization: Bearer {token}" \\\n  -H "Content-Type: application/json" \\\n  -d '{"old_password":"admin123","new_password":"new-pass"}'`
    },
    curlSendApiKey() {
      return `curl -X POST ${this.baseUrl}/api/v1/send \\\n  -H "X-API-Key: smtp_xxxxxxxxxxxx" \\\n  -H "Content-Type: application/json" \\\n  -d '{\n    "to": "recipient@example.com",\n    "subject": "验证码通知",\n    "body": "您的验证码是：123456",\n    "from_name": "我的服务",\n    "is_html": false\n  }'`
    },
    curlSendHtml() {
      return `curl -X POST ${this.baseUrl}/api/v1/send \\\n  -H "X-API-Key: smtp_xxxxxxxxxxxx" \\\n  -H "Content-Type: application/json" \\\n  -d '{\n    "to": "user@example.com",\n    "subject": "欢迎注册",\n    "body": "<h1>欢迎</h1><p>感谢注册！</p>",\n    "from_name": "我的服务",\n    "is_html": true\n  }'`
    },
    codeExamplePython() {
      return `import requests\n\nAPI_URL = "${this.baseUrl}/api/v1/send"\nAPI_KEY = "smtp_xxxxxxxxxxxx"\n\ndef send_email(to, subject, body, is_html=False):\n    resp = requests.post(API_URL,\n        headers={"X-API-Key": API_KEY},\n        json={"to": to, "subject": subject, "body": body, "is_html": is_html},\n        timeout=30)\n    return resp.json()\n\nresult = send_email("user@example.com", "验证码", "您的验证码是：123456")\nprint(result)`
    },
    codeExampleNodejs() {
      return `const axios = require('axios');\n\nconst API_URL = '${this.baseUrl}/api/v1/send';\nconst API_KEY = 'smtp_xxxxxxxxxxxx';\n\nasync function sendEmail(to, subject, body, isHtml = false) {\n  const resp = await axios.post(API_URL,\n    { to, subject, body, is_html: isHtml },\n    { headers: { 'X-API-Key': API_KEY } });\n  return resp.data;\n}\n\nsendEmail('user@example.com', '验证码', '您的验证码是：123456').then(console.log);`
    },
    codeExamplePhp() {
      return `<?php\nfunction sendEmail($to, $subject, $body) {\n    $ch = curl_init('${this.baseUrl}/api/v1/send');\n    curl_setopt_array($ch, [\n        CURLOPT_POST => true,\n        CURLOPT_POSTFIELDS => json_encode(['to'=>$to,'subject'=>$subject,'body'=>$body]),\n        CURLOPT_RETURNTRANSFER => true,\n        CURLOPT_HTTPHEADER => ['Content-Type: application/json', 'X-API-Key: smtp_xxxx'],\n    ]);\n    $r = curl_exec($ch); curl_close($ch);\n    return json_decode($r, true);\n}\nvar_dump(sendEmail('user@example.com', '验证码', '您的验证码是：123456'));`
    },
    codeExampleGo() {
      return `package main\n\nimport (\n    "bytes"; "encoding/json"; "fmt"; "net/http"\n)\n\nfunc sendEmail(to, subject, body string) {\n    payload, _ := json.Marshal(map[string]interface{}{\n        "to": to, "subject": subject, "body": body,\n    })\n    req, _ := http.NewRequest("POST", "${this.baseUrl}/api/v1/send", bytes.NewReader(payload))\n    req.Header.Set("Content-Type", "application/json")\n    req.Header.Set("X-API-Key", "smtp_xxxxxxxxxxxx")\n    resp, _ := http.DefaultClient.Do(req)\n    defer resp.Body.Close()\n    fmt.Println(resp.Status)\n}\n\nfunc main() { sendEmail("user@example.com", "验证码", "您的验证码是：123456") }`
    },
    currentCodeExample() {
      const map = { python: this.codeExamplePython, nodejs: this.codeExampleNodejs, php: this.codeExamplePhp, go: this.codeExampleGo }
      return map[this.codeTab] || ''
    },
    currentRecipientGroup() {
      return this.recipientGroups.find(group => group.id === this.currentGroupId) || null
    },
    totalRecipientCount() {
      return this.recipientGroups.reduce((total, group) => total + Number(group.count || 0), 0)
    },
    selectedRecipientCount() {
      return this.currentGroupId ? this.recipients.length : 0
    },
    activeRecipientCount() {
      return this.recipients.filter(recipient => recipient.status === 'active').length
    },
    blockedRecipientCount() {
      return this.recipients.filter(recipient => recipient.status !== 'active').length
    },
  },
  methods: {
    showToast(msg, type = 'success') {
      this.toast = { show: true, msg, type }
      setTimeout(() => { this.toast.show = false }, 3000)
    },
    copyText(text, key) {
      navigator.clipboard.writeText(text).then(() => {
        this.copiedKey = key
        setTimeout(() => { this.copiedKey = '' }, 2000)
      })
    },
    async loadVersion() {
      try {
        const res = await axios.get(`${API}/version`)
        this.currentVersion = res.data.version
      } catch(e) {}
    },
    compareVersions(a, b) {
      // 比较语义化版本 a > b 返回 1, a < b 返回 -1, 相等返回 0
      const pa = (a || '').replace(/^v/, '').split('.').map(Number)
      const pb = (b || '').replace(/^v/, '').split('.').map(Number)
      for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
        const na = pa[i] || 0, nb = pb[i] || 0
        if (na > nb) return 1
        if (na < nb) return -1
      }
      return 0
    },
    async checkUpdate() {
      if (!this.currentVersion) await this.loadVersion()
      this.updateChecking = true
      this.updateStatus = ''
      try {
        const res = await axios.get('https://api.github.com/repos/DoBestone/smtp-lite/releases/latest')
        this.latestVersion = res.data.tag_name
        this.updateStatus = this.compareVersions(this.latestVersion, this.currentVersion) > 0 ? 'available' : 'latest'
      } catch(e) {
        this.updateStatus = 'error'
      } finally {
        this.updateChecking = false
      }
    },
    async doUpdate() {
      this.updateProgress = 'updating'
      this.updateStep = 1
      try {
        const prepRes = await axios.post(`${API}/system/update-prepare`, {}, { headers: this.getHeaders() })
        const confirmToken = prepRes.data.confirm_token
        await axios.post(`${API}/system/update`, { confirm_token: confirmToken }, { headers: this.getHeaders() })
        this.updateStep = 2
        this.pollForNewVersion()
      } catch(e) {
        this.updateProgress = 'error'
      }
    },
    reloadPage() { window.location.reload() },
    pollForNewVersion() {
      const start = Date.now()
      const timeout = 120000
      const target = this.latestVersion
      const poll = async () => {
        if (Date.now() - start > timeout) { this.updateProgress = 'timeout'; return }
        try {
          const res = await axios.get(`${API}/version`)
          if (res.data.version === target) {
            this.currentVersion = target
            this.updateStep = 3
            this.updateProgress = 'done'
            this.updateStatus = 'latest'
            setTimeout(() => window.location.reload(), 3000)
            return
          }
        } catch(e) { /* 服务重启中，正常 */ }
        setTimeout(poll, 2000)
      }
      setTimeout(poll, 5000) // 给服务器 5 秒开始编译
    },
    switchTab(key) {
      this.tab = key
      window.location.hash = key
      if (key === 'logs') this.loadLogs()
      if (key === 'stats') this.loadStats()
      if (key === 'smtp') this.loadSmtpAccounts()
      if (key === 'docs' && !this.currentVersion) this.loadVersion()
      if (key === 'system') { this.loadVersion(); this.checkUpdate() }
    },
    async login() {
      this.loginLoading = true; this.loginError = ''
      try {
        const res = await axios.post(`${API}/auth/login`, this.loginForm)
        localStorage.setItem('token', res.data.token)
        this.isLoggedIn = true
        this.loadData()
      } catch (e) {
        this.loginError = e.response?.data?.error || '用户名或密码错误'
      } finally { this.loginLoading = false }
    },
    logout() {
      localStorage.removeItem('token')
      this.isLoggedIn = false
      this.loginForm = { username: '', password: '' }
      this.smtpAccounts = []; this.apiKeys = []; this.logs = []; this.stats = {}
    },
    getHeaders() { return { Authorization: `Bearer ${localStorage.getItem('token')}` } },
    async loadData() {
      await Promise.all([
        this.loadSmtpAccounts(),
        this.loadApiKeys(),
        this.loadStats(),
        this.loadTemplates(),
        this.loadRecipientGroups(),
        this.loadWebhooks(),
        this.loadBlacklist(),
        this.loadQueueStats(),
      ])
    },
    async loadSmtpAccounts() {
      try {
        const res = await axios.get(`${API}/smtp-accounts`, { headers: this.getHeaders() })
        this.smtpAccounts = res.data || []
      } catch (e) { this.handleAuthError(e) }
    },
    async loadApiKeys() {
      try {
        const res = await axios.get(`${API}/api-keys`, { headers: this.getHeaders() })
        this.apiKeys = res.data || []
      } catch (e) { this.handleAuthError(e) }
    },
    async loadLogs() {
      try {
        const res = await axios.get(`${API}/send/logs`, {
          headers: this.getHeaders(),
          params: { page: this.logPage, page_size: this.logPageSize }
        })
        this.logs = res.data.logs || []; this.logTotal = res.data.total || 0
      } catch (e) { this.handleAuthError(e) }
    },
    async loadStats() {
      try {
        const res = await axios.get(`${API}/stats`, { headers: this.getHeaders() })
        this.stats = res.data || {}
      } catch (e) { this.handleAuthError(e) }
    },
    handleAuthError(e) {
      if (e.response?.status === 401) { this.logout(); this.showToast('登录已过期，请重新登录', 'error') }
    },
    openAddSmtp() {
      this.editingSmtp = null
      this.smtpForm = { email: '', password: '', smtp_host: '', smtp_port: 587, daily_limit: 500 }
      this.smtpFormError = ''; this.showSmtpModal = true
    },
    openEditSmtp(acc) {
      this.editingSmtp = acc
      this.smtpForm = { email: acc.email, password: '', smtp_host: acc.smtp_host, smtp_port: acc.smtp_port, daily_limit: acc.daily_limit }
      this.smtpFormError = ''; this.showSmtpModal = true
    },
    applyPreset(name) {
      const p = SMTP_PRESETS[name]
      if (p) { this.smtpForm.smtp_host = p.host; this.smtpForm.smtp_port = p.port }
    },
    async saveSmtpAccount() {
      this.smtpFormError = ''
      try {
        if (this.editingSmtp) {
          const payload = { email: this.smtpForm.email, smtp_host: this.smtpForm.smtp_host, smtp_port: this.smtpForm.smtp_port, daily_limit: this.smtpForm.daily_limit }
          if (this.smtpForm.password) payload.password = this.smtpForm.password
          await axios.put(`${API}/smtp-accounts/${this.editingSmtp.id}`, payload, { headers: this.getHeaders() })
          this.showToast('账号已更新')
        } else {
          await axios.post(`${API}/smtp-accounts`, this.smtpForm, { headers: this.getHeaders() })
          this.showToast('账号已添加')
        }
        this.showSmtpModal = false; this.loadSmtpAccounts()
      } catch (e) { this.smtpFormError = e.response?.data?.error || '操作失败' }
    },
    async testConnection(acc) {
      this.testingId = acc.id
      try {
        const res = await axios.post(`${API}/smtp-accounts/${acc.id}/test`, {}, { headers: this.getHeaders() })
        if (res.data.success) {
          acc.last_error = ''
          await this.loadSmtpAccounts()
        }
        this.showToast(res.data.success ? `✓ ${acc.email} 连接成功` : '连接失败：' + res.data.error, res.data.success ? 'success' : 'error')
      } catch (e) {
        this.showToast('测试失败：' + (e.response?.data?.error || e.message), 'error')
      } finally { this.testingId = null }
    },
    openTestSend(acc) {
      this.testSendAccount = acc; this.testSendTo = ''; this.testSendResult = null; this.showTestSend = true
    },
    async doTestSend() {
      if (!this.testSendTo) return
      this.testSendLoading = true; this.testSendResult = null
      try {
        const res = await axios.post(`${API}/smtp-accounts/${this.testSendAccount.id}/test-send`, { to: this.testSendTo }, { headers: this.getHeaders() })
        this.testSendAccount.last_error = ''
        await this.loadSmtpAccounts()
        this.testSendResult = { success: true, message: res.data.message }
      } catch (e) {
        this.testSendResult = { success: false, error: e.response?.data?.error || '发送失败' }
      } finally { this.testSendLoading = false }
    },
    async toggleSmtp(id) {
      try {
        await axios.post(`${API}/smtp-accounts/${id}/toggle`, {}, { headers: this.getHeaders() })
        await this.loadSmtpAccounts()
      } catch (e) { this.showToast('操作失败', 'error') }
    },
    deleteSmtp(id) {
      this.confirmDialog = {
        show: true,
        title: '删除 SMTP 账号',
        msg: '确定要删除此 SMTP 账号？删除后无法恢复。',
        confirmText: '确认删除',
        onConfirm: async () => {
          try {
            await axios.delete(`${API}/smtp-accounts/${id}`, { headers: this.getHeaders() })
            this.showToast('账号已删除'); this.loadSmtpAccounts()
          } catch (e) { this.showToast('删除失败', 'error') }
        }
      }
    },
    createApiKey() {
      this.newKeyName = ''
      this.showCreateKey = true
      this.$nextTick(() => { this.$refs.keyNameInput && this.$refs.keyNameInput.focus() })
    },
    async doCreateApiKey() {
      if (!this.newKeyName.trim()) return
      try {
        const res = await axios.post(`${API}/api-keys`, { name: this.newKeyName.trim() }, { headers: this.getHeaders() })
        this.showCreateKey = false
        this.newKeyInfo = res.data; this.loadApiKeys()
      } catch (e) { this.showToast('创建失败：' + (e.response?.data?.error || e.message), 'error') }
    },
    deleteApiKey(id) {
      this.confirmDialog = {
        show: true,
        title: '删除 API Key',
        msg: '确定要删除此 API Key？删除后接口调用将立即失效。',
        confirmText: '确认删除',
        onConfirm: async () => {
          try {
            await axios.delete(`${API}/api-keys/${id}`, { headers: this.getHeaders() })
            this.showToast('已删除'); this.loadApiKeys()
          } catch (e) { this.showToast('删除失败', 'error') }
        }
      }
    },
    resetApiKey(id, name) {
      this.confirmDialog = {
        show: true,
        title: '重置 API Key',
        msg: `确定要重置「${name}」的 Key？原 Key 将立即失效，系统会生成新的 Key。`,
        confirmText: '确认重置',
        onConfirm: async () => {
          try {
            const res = await axios.post(`${API}/api-keys/${id}/reset`, {}, { headers: this.getHeaders() })
            this.newKeyInfo = res.data
            this.loadApiKeys()
          } catch (e) { this.showToast('重置失败：' + (e.response?.data?.error || e.message), 'error') }
        }
      }
    },
    copyKey() {
      navigator.clipboard.writeText(this.newKeyInfo.key)
        .then(() => this.showToast('已复制到剪贴板'))
        .catch(() => this.showToast('请手动复制', 'error'))
    },
    async changePassword() {
      this.pwdError = ''; this.pwdSuccess = ''
      if (this.pwdForm.newPwd !== this.pwdForm.confirmPwd) { this.pwdError = '两次输入的新密码不一致'; return }
      if (this.pwdForm.newPwd.length < 6) { this.pwdError = '新密码至少需要 6 位字符'; return }
      this.pwdLoading = true
      try {
        const res = await axios.post(`${API}/auth/change-password`, {
          old_password: this.pwdForm.oldPwd,
          new_password: this.pwdForm.newPwd,
        }, { headers: this.getHeaders() })
        this.pwdSuccess = res.data.message || '密码修改成功，即将重新登录...'
        this.pwdForm = { oldPwd: '', newPwd: '', confirmPwd: '' }
        setTimeout(() => { this.showChangePwd = false; this.logout() }, 1500)
      } catch (e) {
        this.pwdError = e.response?.data?.error || '修改失败'
      } finally { this.pwdLoading = false }
    },
    formatDate(date) {
      if (!date) return '-'
      return new Date(date).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
    },
    // ===== 模板管理 =====
    async loadTemplates() {
      try {
        const res = await axios.get(`${API}/templates`, { headers: this.getHeaders() })
        this.templates = res.data
      } catch(e) { console.error('加载模板失败', e) }
    },
    openCreateTemplate() {
      this.editingTemplate = null
      this.templateForm = { name: '', subject: '', body: '', is_html: true, description: '' }
      this.showTemplateModal = true
    },
    openEditTemplate(t) {
      this.editingTemplate = t
      this.templateForm = { name: t.name, subject: t.subject, body: t.body, is_html: t.is_html, description: t.description }
      this.showTemplateModal = true
    },
    async saveTemplate() {
      try {
        if (this.editingTemplate) {
          await axios.put(`${API}/templates/${this.editingTemplate.id}`, this.templateForm, { headers: this.getHeaders() })
        } else {
          await axios.post(`${API}/templates`, this.templateForm, { headers: this.getHeaders() })
        }
        this.showTemplateModal = false
        this.loadTemplates()
        this.showToast(this.editingTemplate ? '模板已更新' : '模板已创建')
      } catch(e) {
        this.showToast(e.response?.data?.error || '保存失败', 'error')
      }
    },
    async deleteTemplate(id) {
      if (!confirm('确定删除此模板？')) return
      try {
        await axios.delete(`${API}/templates/${id}`, { headers: this.getHeaders() })
        this.loadTemplates()
        this.showToast('模板已删除')
      } catch(e) {
        this.showToast('删除失败', 'error')
      }
    },
    // ===== 收件人管理 =====
    async loadRecipientGroups() {
      try {
        const res = await axios.get(`${API}/recipient-groups`, { headers: this.getHeaders() })
        this.recipientGroups = res.data
      } catch(e) { console.error('加载分组失败', e) }
    },
    async loadRecipients(groupId) {
      if (!groupId) return
      this.currentGroupId = groupId
      try {
        const res = await axios.get(`${API}/recipients?group_id=${groupId}`, { headers: this.getHeaders() })
        this.recipients = res.data
      } catch(e) { console.error('加载收件人失败', e) }
    },
    openCreateGroup() {
      this.editingGroup = null
      this.groupForm = { name: '', description: '' }
      this.showGroupModal = true
    },
    openEditGroup(group) {
      this.editingGroup = group
      this.groupForm = { name: group.name || '', description: group.description || '' }
      this.showGroupModal = true
    },
    async saveGroup() {
      const payload = {
        name: this.groupForm.name.trim(),
        description: this.groupForm.description.trim(),
      }
      if (!payload.name) {
        this.showToast('分组名称不能为空', 'error')
        return
      }
      try {
        if (this.editingGroup) {
          await axios.put(`${API}/recipient-groups/${this.editingGroup.id}`, payload, { headers: this.getHeaders() })
        } else {
          await axios.post(`${API}/recipient-groups`, payload, { headers: this.getHeaders() })
        }
        this.showGroupModal = false
        await this.loadRecipientGroups()
        this.showToast(this.editingGroup ? '分组已更新' : '分组已创建')
        this.editingGroup = null
      } catch(e) {
        this.showToast(e.response?.data?.error || (this.editingGroup ? '更新失败' : '创建失败'), 'error')
      }
    },
    async deleteGroup(id) {
      if (!confirm('确定删除此分组？分组内的收件人也会被删除。')) return
      try {
        await axios.delete(`${API}/recipient-groups/${id}`, { headers: this.getHeaders() })
        this.loadRecipientGroups()
        if (this.currentGroupId === id) {
          this.currentGroupId = ''
          this.recipients = []
        }
        this.showToast('分组已删除')
      } catch(e) {
        this.showToast('删除失败', 'error')
      }
    },
    openCreateRecipient() {
      this.recipientForm = { email: '', name: '' }
      this.showRecipientModal = true
    },
    async saveRecipient() {
      try {
        await axios.post(`${API}/recipients`, { ...this.recipientForm, group_id: this.currentGroupId }, { headers: this.getHeaders() })
        this.showRecipientModal = false
        await Promise.all([this.loadRecipients(this.currentGroupId), this.loadRecipientGroups()])
        this.showToast('收件人已添加')
      } catch(e) {
        this.showToast(e.response?.data?.error || '添加失败', 'error')
      }
    },
    async deleteRecipient(id) {
      if (!confirm('确定删除此收件人？')) return
      try {
        await axios.delete(`${API}/recipients/${id}`, { headers: this.getHeaders() })
        await Promise.all([this.loadRecipients(this.currentGroupId), this.loadRecipientGroups()])
        this.showToast('收件人已删除')
      } catch(e) {
        this.showToast('删除失败', 'error')
      }
    },
    openBatchImport() {
      this.batchEmails = ''
      this.showBatchImport = true
    },
    async doBatchImport() {
      try {
        const res = await axios.post(`${API}/recipients/batch`, { group_id: this.currentGroupId, emails: this.batchEmails }, { headers: this.getHeaders() })
        this.showBatchImport = false
        await Promise.all([this.loadRecipients(this.currentGroupId), this.loadRecipientGroups()])
        this.showToast(`成功导入 ${res.data.success} 个收件人`)
      } catch(e) {
        this.showToast(e.response?.data?.error || '导入失败', 'error')
      }
    },
    // ===== Webhook 管理 =====
    async loadWebhooks() {
      try {
        const res = await axios.get(`${API}/webhooks`, { headers: this.getHeaders() })
        this.webhooks = res.data
      } catch(e) { console.error('加载 Webhook 失败', e) }
    },
    openCreateWebhook() {
      this.editingWebhook = null
      this.webhookForm = { name: '', url: '', secret: '', events: [] }
      this.showWebhookModal = true
    },
    async saveWebhook() {
      try {
        await axios.post(`${API}/webhooks`, this.webhookForm, { headers: this.getHeaders() })
        this.showWebhookModal = false
        this.loadWebhooks()
        this.showToast('Webhook 已创建')
      } catch(e) {
        this.showToast(e.response?.data?.error || '保存失败', 'error')
      }
    },
    async toggleWebhook(id) {
      try {
        await axios.post(`${API}/webhooks/${id}/toggle`, {}, { headers: this.getHeaders() })
        this.loadWebhooks()
      } catch(e) {
        this.showToast('操作失败', 'error')
      }
    },
    async deleteWebhook(id) {
      if (!confirm('确定删除此 Webhook？')) return
      try {
        await axios.delete(`${API}/webhooks/${id}`, { headers: this.getHeaders() })
        this.loadWebhooks()
        this.showToast('Webhook 已删除')
      } catch(e) {
        this.showToast('删除失败', 'error')
      }
    },
    async testWebhook(id) {
      try {
        await axios.post(`${API}/webhooks/${id}/test`, {}, { headers: this.getHeaders() })
        this.showToast('测试事件已发送')
      } catch(e) {
        this.showToast('发送失败', 'error')
      }
    },
    // ===== 黑名单管理 =====
    async loadBlacklist() {
      try {
        const res = await axios.get(`${API}/blacklist`, { headers: this.getHeaders() })
        this.blacklist = res.data
      } catch(e) { console.error('加载黑名单失败', e) }
    },
    openCreateBlacklist() {
      this.blacklistForm = { email: '', reason: '' }
      this.showBlacklistModal = true
    },
    async saveBlacklist() {
      try {
        await axios.post(`${API}/blacklist`, this.blacklistForm, { headers: this.getHeaders() })
        this.showBlacklistModal = false
        this.loadBlacklist()
        this.showToast('已添加到黑名单')
      } catch(e) {
        this.showToast(e.response?.data?.error || '添加失败', 'error')
      }
    },
    async deleteBlacklist(id) {
      if (!confirm('确定移除此黑名单？')) return
      try {
        await axios.delete(`${API}/blacklist/${id}`, { headers: this.getHeaders() })
        this.loadBlacklist()
        this.showToast('已从黑名单移除')
      } catch(e) {
        this.showToast('操作失败', 'error')
      }
    },
    // ===== 队列状态 =====
    async loadQueueStats() {
      try {
        const res = await axios.get(`${API}/queue/stats`, { headers: this.getHeaders() })
        this.queueStats = res.data
      } catch(e) { console.error('加载队列状态失败', e) }
    },
    // ===== 发送邮件 =====
    resetSendForm() {
      this.sendForm = { to: '', subject: '', body: '', is_html: false, from_name: '', cc: '', bcc: '', track_enabled: false }
      this.sendRecipientGroupId = ''
      this.sendRecipientOptions = []
      this.sendRecipientSelection = ''
      this.batchEmailsList = ''
      this.scheduledTime = ''
      this.sendResult = null
    },
    async fetchSendRecipientOptions(groupId) {
      if (!groupId) {
        this.sendRecipientOptions = []
        return []
      }
      const res = await axios.get(`${API}/recipients?group_id=${groupId}`, { headers: this.getHeaders() })
      const recipients = (res.data || []).filter(recipient => recipient.status === 'active')
      this.sendRecipientOptions = recipients
      return recipients
    },
    async handleSendRecipientGroupChange() {
      this.sendRecipientSelection = ''
      if (!this.sendRecipientGroupId) {
        this.sendRecipientOptions = []
        return
      }
      try {
        await this.fetchSendRecipientOptions(this.sendRecipientGroupId)
      } catch (e) {
        this.sendRecipientOptions = []
        this.showToast('加载收件人分组失败', 'error')
      }
    },
    applySendRecipientSelection() {
      if (!this.sendRecipientSelection) return
      this.sendForm.to = this.sendRecipientSelection
    },
    async importSendGroupRecipients() {
      if (!this.sendRecipientGroupId) return
      try {
        const recipients = this.sendRecipientOptions.length > 0
          ? this.sendRecipientOptions
          : await this.fetchSendRecipientOptions(this.sendRecipientGroupId)
        if (!recipients.length) {
          this.showToast('该分组没有可导入的正常状态收件人', 'error')
          return
        }

        const currentEmails = this.batchEmailsList
          .split('\n')
          .map(email => email.trim())
          .filter(Boolean)
        const existingEmails = new Set(currentEmails.map(email => email.toLowerCase()))
        const appendedEmails = []
        let skippedCount = 0

        for (const recipient of recipients) {
          const email = (recipient.email || '').trim()
          const normalizedEmail = email.toLowerCase()
          if (!email) continue
          if (existingEmails.has(normalizedEmail)) {
            skippedCount++
            continue
          }
          existingEmails.add(normalizedEmail)
          appendedEmails.push(email)
        }

        if (!appendedEmails.length) {
          this.showToast('当前分组成员已全部在列表中', 'error')
          return
        }

        this.batchEmailsList = [...currentEmails, ...appendedEmails].join('\n')
        this.showToast(`已导入 ${appendedEmails.length} 个收件人${skippedCount ? `，跳过 ${skippedCount} 个重复项` : ''}`)
      } catch (e) {
        this.showToast('导入分组成员失败', 'error')
      }
    },
    async doSendEmail() {
      this.sendLoading = true
      this.sendResult = null
      try {
        if (this.sendMode === 'single') {
          const data = {
            to: this.sendForm.to,
            subject: this.sendForm.subject,
            body: this.sendForm.body,
            is_html: this.sendForm.is_html,
            from_name: this.sendForm.from_name || undefined,
            track_enabled: this.sendForm.track_enabled,
          }
          if (this.sendForm.cc) data.cc = this.sendForm.cc.split(',').map(e => e.trim()).filter(Boolean)
          if (this.sendForm.bcc) data.bcc = this.sendForm.bcc.split(',').map(e => e.trim()).filter(Boolean)
          const res = await axios.post(`${API}/send`, data, { headers: this.getHeaders() })
          this.sendResult = res.data
        } else if (this.sendMode === 'batch') {
          const emails = this.batchEmailsList.split('\n').map(e => e.trim()).filter(Boolean)
          const res = await axios.post(`${API}/send/batch`, {
            name: `批量发送 ${new Date().toLocaleString()}`,
            emails,
            subject: this.sendForm.subject,
            body: this.sendForm.body,
            is_html: this.sendForm.is_html,
            from_name: this.sendForm.from_name,
          }, { headers: this.getHeaders() })
          this.sendResult = { success: true, message: `已加入队列，共 ${emails.length} 封`, batch_id: res.data.id }
        } else if (this.sendMode === 'scheduled') {
          const res = await axios.post(`${API}/send/scheduled`, {
            to: this.sendForm.to,
            subject: this.sendForm.subject,
            body: this.sendForm.body,
            is_html: this.sendForm.is_html,
            from_name: this.sendForm.from_name,
            scheduled_at: this.scheduledTime,
          }, { headers: this.getHeaders() })
          this.sendResult = { success: true, message: `已加入定时队列`, scheduled_at: this.scheduledTime }
        }
        if (this.sendResult?.success) {
          this.showToast(this.sendResult.message)
          this.loadStats()
        }
      } catch(e) {
        this.sendResult = { success: false, message: e.response?.data?.error || e.response?.data?.message || '发送失败' }
      } finally {
        this.sendLoading = false
      }
    },
    applyTemplate(t) {
      this.sendForm.subject = t.subject || ''
      this.sendForm.body = t.body || ''
      this.sendForm.is_html = t.is_html || false
      this.tab = 'send'
      this.showToast('已应用模板')
    }
  }
}
</script>

<style>
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

:root {
  --blue: #4f46e5; --blue-l: #6366f1;
  --blue-50: #eef2ff; --blue-100: #e0e7ff;
  --cyan: #0891b2; --cyan-50: #ecfeff;
  --green: #059669; --green-50: #ecfdf5; --green-100: #d1fae5;
  --red: #dc2626; --red-50: #fef2f2; --red-100: #fee2e2;
  --purple: #7c3aed; --purple-50: #f5f3ff;
  --gray-50: #f8fafc; --gray-100: #f1f5f9; --gray-200: #e2e8f0;
  --gray-300: #cbd5e1; --gray-400: #94a3b8; --gray-500: #64748b;
  --gray-600: #475569; --gray-700: #334155; --gray-800: #1e293b; --gray-900: #0f172a;
  --gradient-blue: linear-gradient(135deg, #4f46e5, #6366f1);
  --gradient-indigo: linear-gradient(135deg, #4338ca 0%, #6366f1 50%, #818cf8 100%);
  --radius: 10px; --radius-lg: 14px; --radius-xl: 18px;
  --shadow: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.02);
  --shadow-md: 0 4px 16px rgba(0,0,0,0.06), 0 1px 3px rgba(0,0,0,0.04);
  --shadow-lg: 0 8px 30px rgba(0,0,0,0.08), 0 2px 8px rgba(0,0,0,0.04);
  --shadow-card: 0 1px 3px rgba(0,0,0,0.03), 0 1px 2px rgba(0,0,0,0.02), 0 0 0 1px rgba(0,0,0,0.03);
  --topbar-h: 58px;
  --sidebar-w: 250px;
}

html { font-size: 15px; -webkit-text-size-adjust: 100%; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Inter', 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: #f0f2f5; color: var(--gray-700); line-height: 1.6; min-height: 100vh;
}
input, button, select, textarea { font-family: inherit; font-size: inherit; }
button { cursor: pointer; }

/* ====== Login ====== */
.login-page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(160deg, #eef2ff 0%, #e8ecf7 30%, #f0f2ff 60%, #e8f0fe 100%);
  position: relative; overflow: hidden; padding: 24px;
}
.login-orb { position: absolute; border-radius: 50%; filter: blur(80px); opacity: 0.25; pointer-events: none; }
.login-orb-1 { width: 500px; height: 500px; background: radial-gradient(#818cf8, #4f46e5); top: -120px; left: -100px; }
.login-orb-2 { width: 420px; height: 420px; background: radial-gradient(#a5f3fc, #06b6d4); bottom: -100px; right: -80px; }
.login-card {
  background: rgba(255,255,255,0.96); backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
  border: 1px solid rgba(255,255,255,0.9); border-radius: 20px;
  padding: clamp(32px,5vw,48px); width: 100%; max-width: 420px;
  box-shadow: 0 20px 60px rgba(79,70,229,0.08), 0 8px 24px rgba(0,0,0,0.04), 0 0 0 1px rgba(79,70,229,0.04);
  position: relative; z-index: 1;
}
.login-logo { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.logo-text { font-size: 1.35rem; font-weight: 700; color: var(--gray-900); letter-spacing: -0.02em; }
.login-subtitle { color: var(--gray-400); font-size: 0.875rem; margin-bottom: 28px; margin-left: 48px; }
.login-form { display: flex; flex-direction: column; gap: 16px; }
.login-hint { text-align: center; color: var(--gray-400); font-size: 0.8rem; margin-top: 18px; }
.btn-login {
  width: 100%; height: 46px; background: var(--gradient-indigo); color: white;
  border: none; border-radius: var(--radius);
  font-size: 0.9375rem; font-weight: 600;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  transition: all 0.2s ease; margin-top: 6px;
  box-shadow: 0 4px 14px rgba(79,70,229,0.25);
}
.btn-login:hover { transform: translateY(-1px); box-shadow: 0 6px 20px rgba(79,70,229,0.35); }
.btn-login:active { transform: translateY(0); }
.btn-login.loading { opacity: 0.7; pointer-events: none; }

.logo-icon {
  width: 38px; height: 38px; background: var(--gradient-indigo); border-radius: 10px;
  display: flex; align-items: center; justify-content: center; color: white; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(79,70,229,0.25);
}
.logo-icon.sm { width: 30px; height: 30px; border-radius: 8px; }

/* ====== Sidebar ====== */
.topbar-brand { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.brand-name { font-size: 1.05rem; font-weight: 700; color: var(--gray-900); letter-spacing: -0.02em; }
.nav-icon { display: flex; align-items: center; width: 18px; justify-content: center; }
.pill-dot { width: 7px; height: 7px; background: #22c55e; border-radius: 50%; box-shadow: 0 0 6px rgba(34,197,94,0.4); }
.sidebar {
  position: fixed; top: 0; left: 0; bottom: 0;
  width: var(--sidebar-w); background: #fff;
  border-right: 1px solid var(--gray-100);
  display: flex; flex-direction: column;
  z-index: 150; overflow-y: auto;
  transition: transform 0.3s cubic-bezier(.4,0,.2,1);
}
.sidebar-header {
  display: flex; align-items: center; gap: 10px;
  padding: 22px 20px 14px; flex-shrink: 0;
}
.sidebar-stats {
  display: flex; align-items: center; gap: 7px;
  margin: 0 14px 10px; padding: 8px 14px;
  background: linear-gradient(135deg, #eef2ff, #e0e7ff);
  color: var(--blue); font-size: 0.8rem; font-weight: 600;
  border-radius: 10px; border: 1px solid rgba(79,70,229,0.08);
}
.sidebar-nav {
  flex: 1; display: flex; flex-direction: column; gap: 2px;
  padding: 6px 10px; overflow-y: auto;
}
.sidebar-nav-btn {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border: none; border-radius: var(--radius);
  background: transparent; color: var(--gray-500);
  font-size: 0.875rem; font-weight: 500;
  transition: all 0.2s ease; white-space: nowrap;
  text-align: left; cursor: pointer;
  position: relative;
}
.sidebar-nav-btn:hover { background: var(--gray-50); color: var(--gray-700); }
.sidebar-nav-btn.active {
  background: linear-gradient(135deg, #eef2ff, #e0e7ff);
  color: var(--blue); font-weight: 600;
  box-shadow: 0 1px 4px rgba(79,70,229,0.08);
}
.sidebar-nav-btn.active::before {
  content: '';
  position: absolute; left: 0; top: 50%; transform: translateY(-50%);
  width: 3px; height: 20px; background: var(--blue);
  border-radius: 0 3px 3px 0;
}
.nav-badge {
  margin-left: auto; font-size: 0.6875rem; font-weight: 600;
  padding: 1px 7px; border-radius: 10px;
  background: var(--red); color: #fff;
  line-height: 1.5; letter-spacing: 0.01em;
  animation: badge-pulse 2s ease-in-out infinite;
}
@keyframes badge-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
.sidebar-footer {
  border-top: 1px solid var(--gray-100); padding: 10px;
  display: flex; flex-direction: column; gap: 2px; flex-shrink: 0;
}
.sidebar-footer-btn {
  display: flex; align-items: center; gap: 8px;
  padding: 9px 14px; border: none; border-radius: var(--radius);
  background: transparent; color: var(--gray-500);
  font-size: 0.8125rem; font-weight: 500; cursor: pointer;
  transition: all 0.2s ease; text-align: left;
}
.sidebar-footer-btn:hover { background: var(--gray-50); color: var(--gray-700); }
.sidebar-footer-btn.logout:hover { color: var(--red); background: var(--red-50); }
.sidebar-overlay {
  position: fixed; inset: 0; background: rgba(15,23,42,0.35);
  backdrop-filter: blur(2px);
  z-index: 140; display: none;
}

/* Mobile topbar */
.mobile-topbar { display: none; }
.mobile-topbar .topbar-inner {
  padding: 0 16px; height: 56px;
  display: flex; align-items: center; justify-content: space-between;
}
.hamburger { display: flex; flex-direction: column; gap: 5px; padding: 8px; border: none; background: transparent; cursor: pointer; }
.hamburger span { display: block; width: 18px; height: 2px; background: var(--gray-600); border-radius: 2px; transition: all 0.2s; }

/* ====== Layout ====== */
.layout { min-height: 100vh; display: flex; background: #f0f2f5; }
.main-content { flex: 1; margin-left: var(--sidebar-w); padding: clamp(24px,3vw,40px) 0 80px; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 clamp(16px,3vw,36px); }

/* ====== Section ====== */
.section { display: flex; flex-direction: column; gap: 22px; }
.section-head {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 14px; flex-wrap: wrap;
}
.section-title {
  font-size: clamp(1.25rem,2.5vw,1.55rem); font-weight: 700;
  color: var(--gray-900); letter-spacing: -0.025em; line-height: 1.3;
}
.section-desc { font-size: 0.875rem; color: var(--gray-400); margin-top: 4px; }

/* ====== Card ====== */
.card {
  background: white; border: 1px solid rgba(0,0,0,0.04);
  border-radius: var(--radius-lg); box-shadow: var(--shadow-card);
  overflow: hidden;
}
.card-head {
  display: flex; align-items: center; gap: 8px;
  padding: 16px 20px 0; color: var(--gray-700);
  font-weight: 600; font-size: 0.9rem;
}
.card-head svg { color: var(--blue); }

/* Form inside card needs padding */
.card > form,
.card > .card-body {
  padding: 24px;
  display: flex; flex-direction: column; gap: 18px;
}

/* ====== Buttons ====== */
.btn-primary {
  display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px;
  background: var(--gradient-indigo); color: white; border: none;
  border-radius: var(--radius); font-size: 0.875rem; font-weight: 600;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(79,70,229,0.2); white-space: nowrap;
}
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(79,70,229,0.3); }
.btn-primary:active { transform: translateY(0); }
.btn-primary:disabled { opacity: 0.6; pointer-events: none; transform: none; }
.btn-outline {
  display: inline-flex; align-items: center; gap: 6px; padding: 8px 16px;
  background: white; color: var(--gray-600); border: 1.5px solid var(--gray-200);
  border-radius: var(--radius); font-size: 0.875rem; font-weight: 500;
  transition: all 0.2s ease; white-space: nowrap;
}
.btn-outline:hover { border-color: var(--blue-l); color: var(--blue); background: var(--blue-50); }
.btn-ghost {
  display: inline-flex; align-items: center; gap: 6px; padding: 9px 18px;
  background: transparent; color: var(--gray-500); border: 1.5px solid var(--gray-200);
  border-radius: var(--radius); font-size: 0.875rem; font-weight: 500;
  transition: all 0.2s ease;
}
.btn-ghost:hover { background: var(--gray-50); color: var(--gray-700); border-color: var(--gray-300); }
.btn-danger {
  display: inline-flex; align-items: center; gap: 6px; padding: 9px 18px;
  background: var(--red); color: white; border: none;
  border-radius: var(--radius); font-size: 0.875rem; font-weight: 600;
  transition: all 0.2s ease;
}
.btn-danger:hover { opacity: 0.9; transform: translateY(-1px); }
.action-group { display: flex; flex-wrap: wrap; gap: 6px; }
.btn-action {
  display: inline-flex; align-items: center; gap: 4px; padding: 5px 12px;
  border: 1.5px solid var(--gray-200); background: white; color: var(--gray-600);
  border-radius: 8px; font-size: 0.8rem; font-weight: 500;
  transition: all 0.15s ease; white-space: nowrap;
}
.btn-action:hover { border-color: var(--blue-l); color: var(--blue); background: var(--blue-50); transform: translateY(-1px); }
.btn-action.danger { color: var(--red); border-color: #fecaca; }
.btn-action.danger:hover { background: var(--red-50); border-color: #f87171; }
.btn-action:disabled { opacity: 0.5; pointer-events: none; }

/* ====== Form ====== */
.field { display: flex; flex-direction: column; gap: 6px; }
.field label {
  font-size: 0.8125rem; font-weight: 600; color: var(--gray-600);
  letter-spacing: 0.01em;
}
.required { color: var(--red); margin-left: 2px; }
.input-wrap { position: relative; display: flex; align-items: center; }
.input-icon { position: absolute; left: 14px; color: var(--gray-400); pointer-events: none; z-index: 1; }
.field .input-wrap input { padding-left: 40px; }
.field input, .modal-box input, .field textarea, .modal-box textarea {
  width: 100%; padding: 10px 14px;
  border: 1.5px solid var(--gray-200);
  border-radius: var(--radius); background: var(--gray-50);
  color: var(--gray-900); font-size: 0.9rem;
  outline: none; transition: all 0.2s ease;
}
.field input:hover, .modal-box input:hover,
.field textarea:hover, .modal-box textarea:hover {
  border-color: var(--gray-300);
}
.field input:focus, .modal-box input:focus,
.field textarea:focus, .modal-box textarea:focus {
  border-color: var(--blue-l); background: white;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}
.field input::placeholder, .field textarea::placeholder { color: var(--gray-400); }
.field textarea { resize: vertical; line-height: 1.65; }
.field-hint { font-size: 0.75rem; color: var(--gray-400); margin-top: 2px; }
.field-actions { display: flex; align-items: center; margin-top: 8px; gap: 4px; }
.recipient-link-box {
  padding: 14px 16px;
  border: 1px dashed var(--blue-100);
  border-radius: var(--radius);
  background: linear-gradient(180deg, rgba(238,242,255,0.45) 0%, rgba(255,255,255,0.9) 100%);
}
.recipient-link-row {
  align-items: end;
}
.recipient-link-action {
  max-width: 180px;
}
.recipient-import-btn {
  width: 100%;
  justify-content: center;
  min-height: 40px;
}

/* Custom checkbox */
.checkbox-label {
  display: inline-flex; align-items: center; gap: 8px;
  font-size: 0.8125rem; color: var(--gray-600); cursor: pointer;
  font-weight: 500; user-select: none;
}
.checkbox-label input[type="checkbox"] {
  width: 16px; height: 16px; accent-color: var(--blue);
  cursor: pointer; border-radius: 4px;
}

.form-select {
  padding: 9px 14px; border: 1.5px solid var(--gray-200);
  border-radius: var(--radius); background: white;
  font-size: 0.875rem; color: var(--gray-700);
  cursor: pointer; outline: none; transition: all 0.2s;
}
.form-select:focus { border-color: var(--blue-l); box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }
.btn-sm { padding: 5px 12px; font-size: 0.8rem; }
.mt-16 { margin-top: 16px; }
.mt-20 { margin-top: 20px; }
.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.input-with-presets { display: flex; flex-direction: column; gap: 6px; }
.preset-btns { display: flex; gap: 6px; flex-wrap: wrap; }
.preset-btn {
  padding: 4px 12px; font-size: 0.775rem; border: 1px solid var(--gray-200);
  background: white; color: var(--gray-500); border-radius: 6px;
  transition: all 0.15s;
}
.preset-btn:hover { border-color: var(--blue); color: var(--blue); background: var(--blue-50); }
.form-error {
  display: flex; align-items: center; gap: 6px;
  color: var(--red); font-size: 0.8125rem;
  background: var(--red-50); padding: 8px 12px;
  border-radius: var(--radius); border: 1px solid var(--red-100);
}

/* ====== Table ====== */
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
thead tr { border-bottom: 1.5px solid var(--gray-100); }
th {
  padding: 12px 16px; text-align: left;
  font-size: 0.72rem; font-weight: 700; color: var(--gray-400);
  text-transform: uppercase; letter-spacing: 0.05em;
  background: var(--gray-50); white-space: nowrap;
}
td { padding: 14px 16px; border-bottom: 1px solid var(--gray-50); vertical-align: middle; }
tbody tr:last-child td { border-bottom: none; }
tbody tr { transition: background 0.15s; }
tbody tr:hover { background: #f8f9ff; }
.row-selected { background: var(--blue-50) !important; }
.cell-main { font-weight: 600; color: var(--gray-900); }
.cell-mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8125rem; color: var(--gray-600); }
.text-truncate { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.text-muted { color: var(--gray-400); }
.text-sm { font-size: 0.8125rem; }
.text-danger { color: var(--red); }
.mt-4 { margin-top: 4px; } .mt-8 { margin-top: 8px; } .mt-12 { margin-top: 12px; }
.p-20 { padding: 20px; }

.quota-wrap { display: flex; flex-direction: column; gap: 4px; min-width: 100px; }
.quota-text { font-size: 0.8125rem; color: var(--gray-600); }
.quota-bar { height: 5px; background: var(--gray-100); border-radius: 9px; overflow: hidden; }
.quota-fill { height: 100%; border-radius: 9px; transition: width 0.5s cubic-bezier(.4,0,.2,1); }

/* ====== Badge ====== */
.badge {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 3px 10px; border-radius: 9999px;
  font-size: 0.75rem; font-weight: 600; white-space: nowrap;
}
.badge-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; opacity: 0.7; }
.badge-success { background: var(--green-50); color: var(--green); border: 1px solid var(--green-100); }
.badge-danger  { background: var(--red-50);   color: var(--red);   border: 1px solid var(--red-100); }
.badge-muted   { background: var(--gray-100); color: var(--gray-500); border: 1px solid var(--gray-200); }
.badge-info    { background: var(--blue-50);  color: var(--blue);  border: 1px solid var(--blue-100); }

.code-tag { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; background: var(--gray-100); color: var(--gray-600); padding: 2px 8px; border-radius: 5px; }
.code-block {
  background: var(--gray-900); color: #e2e8f0; border-radius: var(--radius);
  padding: 16px 18px; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.8125rem; line-height: 1.75; overflow-x: auto; white-space: pre; tab-size: 2;
}

.empty-cell { text-align: center; padding: 48px 16px !important; }
.empty-state {
  display: flex; flex-direction: column; align-items: center;
  gap: 12px; color: var(--gray-400); padding: 20px;
}
.empty-state svg { opacity: 0.3; }
.empty-state p { font-size: 0.875rem; }

/* ====== Pagination ====== */
.pagination {
  display: flex; align-items: center; justify-content: center;
  gap: 12px; padding: 16px; border-top: 1px solid var(--gray-100);
}
.page-btn {
  width: 32px; height: 32px; border: 1.5px solid var(--gray-200);
  background: white; color: var(--gray-500); border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.page-btn:hover:not(:disabled) { border-color: var(--blue); color: var(--blue); background: var(--blue-50); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { font-size: 0.8125rem; color: var(--gray-500); }

/* ====== Stats ====== */
.stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(180px,100%),1fr)); gap: 16px; }
.stat-card {
  background: white; border: 1px solid rgba(0,0,0,0.04);
  border-radius: var(--radius-lg);
  padding: 20px; display: flex; flex-direction: column; gap: 10px;
  box-shadow: var(--shadow-card); transition: all 0.25s ease;
}
.stat-card:hover { box-shadow: var(--shadow-md); transform: translateY(-3px); }
.stat-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
}
.stat-icon.blue   { background: var(--blue-50);   color: var(--blue); }
.stat-icon.green  { background: var(--green-50);  color: var(--green); }
.stat-icon.red    { background: var(--red-50);    color: var(--red); }
.stat-icon.cyan   { background: var(--cyan-50);   color: var(--cyan); }
.stat-icon.purple { background: var(--purple-50); color: var(--purple); }
.stat-icon.orange  { background: #fff7ed; color: #ea580c; }
.stat-icon.teal    { background: #f0fdfa; color: #0d9488; }
.stat-num { font-size: clamp(1.75rem,3vw,2.25rem); font-weight: 700; color: var(--gray-900); letter-spacing: -0.03em; line-height: 1; }
.stat-unit { font-size: 1.2rem; font-weight: 600; }
.stat-label { font-size: 0.8125rem; color: var(--gray-400); font-weight: 500; }

.account-usage-list { padding: 14px 20px 20px; display: flex; flex-direction: column; gap: 16px; }
.account-usage-item { display: flex; flex-direction: column; gap: 6px; }
.account-usage-head { display: flex; justify-content: space-between; align-items: center; }
.account-usage-bar { display: flex; align-items: center; gap: 10px; }
.bar-track { flex: 1; height: 6px; background: var(--gray-100); border-radius: 9px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 9px; transition: width 0.5s cubic-bezier(.4,0,.2,1); }
.bar-label { font-size: 0.8rem; color: var(--gray-500); white-space: nowrap; }

.recipient-overview { margin-bottom: 20px; }
.recipient-shell {
  display: grid;
  grid-template-columns: minmax(280px, 320px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}
.recipient-panel { padding: 0; overflow: hidden; }
.recipient-panel-head {
  display: flex; justify-content: space-between; align-items: flex-start; gap: 16px;
  padding: 18px 18px 16px; border-bottom: 1px solid var(--gray-100);
}
.recipient-panel-title { margin: 0; font-size: 1rem; color: var(--gray-900); }
.recipient-panel-desc { margin: 6px 0 0; font-size: 0.84rem; color: var(--gray-400); line-height: 1.7; }
.recipient-group-list { display: flex; flex-direction: column; gap: 10px; padding: 12px; }
.recipient-group-card {
  padding: 14px; border: 1px solid var(--gray-100); border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8faff 100%);
  cursor: pointer; transition: all 0.2s ease;
}
.recipient-group-card:hover {
  border-color: var(--blue-100); box-shadow: 0 8px 20px rgba(79,70,229,0.08);
  transform: translateY(-1px);
}
.recipient-group-card:focus-visible {
  outline: 3px solid rgba(99,102,241,0.16);
  outline-offset: 2px;
}
.recipient-group-card.active {
  border-color: rgba(79,70,229,0.28);
  background: linear-gradient(180deg, #eef2ff 0%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(79,70,229,0.12);
}
.recipient-group-card-head {
  display: flex; justify-content: space-between; align-items: flex-start; gap: 14px;
}
.recipient-group-card-title { min-width: 0; }
.recipient-group-card-head h3 { margin: 0; font-size: 0.95rem; color: var(--gray-900); }
.recipient-group-card-head p {
  margin: 4px 0 0; font-size: 0.8rem; color: var(--gray-400); line-height: 1.55;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.recipient-group-card-badges {
  display: flex; flex-direction: column; gap: 6px; align-items: flex-end;
}
.recipient-group-card-summary {
  display: flex; flex-wrap: wrap; gap: 8px;
  margin-top: 10px;
}
.recipient-summary-pill {
  display: inline-flex; align-items: center;
  padding: 4px 8px; border-radius: 9999px;
  background: rgba(148,163,184,0.1); color: var(--gray-500);
  font-size: 0.74rem; font-weight: 600;
}
.recipient-group-card-actions {
  margin-top: 12px; padding-top: 10px; border-top: 1px solid rgba(148,163,184,0.16);
  display: flex; gap: 8px; flex-wrap: wrap;
}
.recipient-workspace-hero {
  display: flex; justify-content: space-between; gap: 20px; align-items: stretch;
  padding: 22px; border-bottom: 1px solid var(--gray-100);
  background:
    radial-gradient(circle at top left, rgba(79,70,229,0.12), transparent 44%),
    linear-gradient(135deg, #ffffff 0%, #f8faff 100%);
}
.recipient-workspace-hero.inactive {
  background:
    radial-gradient(circle at top left, rgba(148,163,184,0.12), transparent 40%),
    linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
}
.recipient-workspace-copy { max-width: 620px; }
.recipient-eyebrow {
  display: inline-flex; align-items: center; padding: 5px 10px;
  border-radius: 9999px; background: rgba(79,70,229,0.08); color: var(--blue);
  font-size: 0.75rem; font-weight: 700; letter-spacing: 0.04em;
}
.recipient-workspace-copy h3 {
  margin: 12px 0 8px; font-size: clamp(1.15rem, 2vw, 1.45rem);
  line-height: 1.2; color: var(--gray-900);
}
.recipient-workspace-copy p {
  margin: 0; font-size: 0.88rem; color: var(--gray-500); line-height: 1.75;
}
.recipient-kpi-grid {
  display: grid; grid-template-columns: repeat(3, minmax(88px, 1fr));
  gap: 12px; min-width: 320px;
}
.recipient-kpi {
  display: flex; flex-direction: column; justify-content: space-between; gap: 8px;
  padding: 16px 14px; border-radius: 16px;
  background: rgba(255,255,255,0.82); border: 1px solid rgba(226,232,240,0.8);
}
.recipient-kpi span { font-size: 0.76rem; font-weight: 700; color: var(--gray-500); letter-spacing: 0.03em; }
.recipient-kpi strong { font-size: 1.6rem; line-height: 1; color: var(--gray-900); }
.recipient-kpi small { font-size: 0.76rem; color: var(--gray-400); }
.recipient-toolbar {
  display: flex; justify-content: space-between; align-items: center; gap: 16px;
  padding: 18px 22px; border-bottom: 1px solid var(--gray-100);
}
.recipient-toolbar-copy {
  display: flex; flex-direction: column; align-items: flex-start; gap: 8px;
}
.recipient-toolbar-copy p { margin: 0; font-size: 0.84rem; color: var(--gray-500); }
.recipient-toolbar-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.recipient-entry { display: flex; flex-direction: column; gap: 6px; }
.recipient-entry-sub {
  display: flex; flex-wrap: wrap; gap: 10px; font-size: 0.78rem; color: var(--gray-400);
}
.recipient-mobile-list { padding: 8px 0; }
.recipient-mobile-item { padding: 16px 22px; }
.recipient-mobile-sub { margin-top: 8px; }
.recipient-panel-empty {
  display: flex; flex-direction: column; align-items: flex-start; gap: 12px;
  padding: 24px 22px 28px; color: var(--gray-400);
}
.recipient-panel-empty h3 { margin: 0; font-size: 1rem; color: var(--gray-700); }
.recipient-panel-empty p { margin: 0; font-size: 0.84rem; line-height: 1.7; }
.recipient-workspace-empty { min-height: 240px; justify-content: center; }
.recipient-empty-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.btn-outline:disabled { opacity: 0.6; pointer-events: none; }

/* Mobile list */
.mobile-list { display: none; }
.mobile-item { padding: 16px 18px; border-bottom: 1px solid var(--gray-50); }
.mobile-item:last-child { border-bottom: none; }
.mobile-item-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; margin-bottom: 6px; }
.mobile-item-meta { display: flex; justify-content: space-between; font-size: 0.8rem; color: var(--gray-400); gap: 8px; }

/* ====== Modal ====== */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(15,23,42,0.45);
  backdrop-filter: blur(6px); -webkit-backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  z-index: 200; padding: 20px;
}
.modal-box {
  background: white; border-radius: var(--radius-xl);
  box-shadow: 0 24px 70px rgba(0,0,0,0.16), 0 0 0 1px rgba(0,0,0,0.03);
  width: 100%; max-width: 480px; max-height: 90vh; overflow-y: auto;
  padding: clamp(24px,4vw,32px);
}
.modal-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.modal-head h3 { font-size: 1.05rem; font-weight: 700; color: var(--gray-900); }
.modal-close {
  width: 30px; height: 30px; border: none; background: var(--gray-100);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  color: var(--gray-500); transition: all 0.15s;
}
.modal-close:hover { background: var(--gray-200); color: var(--gray-700); transform: rotate(90deg); }
.modal-box form { display: flex; flex-direction: column; gap: 14px; }
.modal-actions {
  display: flex; justify-content: flex-end; gap: 10px;
  margin-top: 10px; padding-top: 16px;
  border-top: 1px solid var(--gray-100);
}
.modal-desc { font-size: 0.875rem; color: var(--gray-500); margin-bottom: 6px; line-height: 1.6; }

/* ====== Alerts ====== */
.alert {
  display: flex; align-items: center; gap: 8px;
  padding: 11px 14px; border-radius: var(--radius);
  font-size: 0.875rem; font-weight: 500;
}
.alert-success { background: var(--green-50); color: var(--green); border: 1px solid var(--green-100); }
.alert-error   { background: var(--red-50);   color: var(--red);   border: 1px solid var(--red-100); }
.alert-warn    { background: #fffbeb; color: #92400e; border: 1px solid #fde68a; }

.key-display-box {
  background: var(--gray-50); border: 1.5px dashed var(--gray-200);
  border-radius: var(--radius); padding: 14px 16px;
  display: flex; align-items: center; justify-content: space-between;
  gap: 10px; margin: 16px 0 6px;
}
.key-display-box code { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; color: var(--gray-700); word-break: break-all; flex: 1; }
.copy-btn {
  padding: 5px; border: 1px solid var(--gray-200); background: white;
  border-radius: 6px; color: var(--gray-400); display: flex;
  align-items: center; transition: all 0.15s; flex-shrink: 0;
}
.copy-btn:hover { color: var(--blue); border-color: var(--blue-l); }

.spinner { display: inline-block; width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ====== Toast ====== */
.toast {
  position: fixed; bottom: 28px; left: 50%; transform: translateX(-50%);
  display: flex; align-items: center; gap: 8px; padding: 11px 20px;
  border-radius: 9999px; font-size: 0.875rem; font-weight: 500;
  box-shadow: 0 8px 30px rgba(0,0,0,0.12); z-index: 999; white-space: nowrap;
}
.toast-success { background: #fff; color: var(--green); border: 1.5px solid var(--green-100); }
.toast-error   { background: #fff; color: var(--red);   border: 1.5px solid var(--red-100); }
.toast-info    { background: #fff; color: var(--blue);  border: 1.5px solid var(--blue-100); }

/* ====== API Docs ====== */
.base-url-pill {
  display: inline-flex; align-items: center; gap: 6px;
  background: var(--gray-50); border: 1px solid var(--gray-200);
  padding: 6px 14px; border-radius: 8px;
  font-size: 0.8125rem; color: var(--gray-600);
}
.base-url-pill code { font-family: 'SF Mono', 'Fira Code', monospace; }
.doc-section { display: flex; flex-direction: column; gap: 14px; }
.doc-section-title {
  display: flex; align-items: center; gap: 10px;
  font-size: 1rem; font-weight: 700; color: var(--gray-900);
}
.doc-num {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; background: var(--gradient-indigo);
  color: white; font-size: 0.72rem; font-weight: 700; border-radius: 8px;
}
.doc-auth-note { font-size: 0.8rem; color: var(--gray-400); font-weight: 400; }
.doc-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(340px,100%),1fr)); gap: 14px; }
.doc-grid-2 { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(300px,100%),1fr)); gap: 12px; margin: 12px 0; }
.doc-card {
  background: white; border: 1px solid rgba(0,0,0,0.04);
  border-radius: var(--radius-lg); padding: 20px;
  box-shadow: var(--shadow-card); display: flex; flex-direction: column; gap: 12px;
}
.doc-card.featured {
  border-color: var(--blue-100);
  background: linear-gradient(to bottom right, white, var(--blue-50));
  box-shadow: var(--shadow-card), 0 0 0 1px rgba(79,70,229,0.04);
}
.doc-card-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.doc-desc { font-size: 0.875rem; color: var(--gray-500); line-height: 1.6; }
.method-tag {
  display: inline-flex; align-items: center; padding: 3px 9px;
  border-radius: 5px; font-size: 0.72rem; font-weight: 700;
  letter-spacing: 0.03em;
}
.method-tag.get  { background: var(--green-50);  color: var(--green); }
.method-tag.post { background: var(--blue-50);   color: var(--blue); }
.method-tag.put  { background: #fff7ed; color: #c2410c; }
.method-tag.del  { background: var(--red-50);    color: var(--red); }
.path-tag { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.825rem; color: var(--gray-700); background: var(--gray-100); padding: 3px 9px; border-radius: 5px; }
.auth-required-tag { font-size: 0.8rem; color: var(--gray-500); }
.code-block-wrap { display: flex; flex-direction: column; gap: 6px; }
.code-block-label { font-size: 0.75rem; font-weight: 600; color: var(--gray-400); text-transform: uppercase; letter-spacing: 0.04em; }
.params-table-wrap { display: flex; flex-direction: column; gap: 6px; overflow-x: auto; }
.params-table { width: 100%; border-collapse: collapse; font-size: 0.8125rem; }
.params-table th { background: var(--gray-50); padding: 8px 14px; font-size: 0.72rem; font-weight: 700; color: var(--gray-400); text-align: left; text-transform: uppercase; letter-spacing: 0.04em; }
.params-table td { padding: 9px 14px; border-bottom: 1px solid var(--gray-50); }
.params-table tbody tr:last-child td { border-bottom: none; }
.req-yes { color: var(--red); font-weight: 600; }
.req-no  { color: var(--gray-400); }
.doc-list {
  display: flex; flex-direction: column; gap: 0;
  background: white; border: 1px solid rgba(0,0,0,0.04);
  border-radius: var(--radius-lg); overflow: hidden;
  box-shadow: var(--shadow-card);
}
.doc-list-item {
  display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
  padding: 12px 18px; border-bottom: 1px solid var(--gray-50); font-size: 0.875rem;
}
.doc-list-item:last-child { border-bottom: none; }
.doc-list-desc { color: var(--gray-500); }
.tabs-simple { display: flex; gap: 4px; margin-bottom: 8px; flex-wrap: wrap; }
.tab-simple {
  padding: 6px 16px; border: 1.5px solid var(--gray-200);
  background: white; color: var(--gray-500); border-radius: 8px;
  font-size: 0.8125rem; font-weight: 500; transition: all 0.15s;
}
.tab-simple.active { border-color: var(--blue); color: var(--blue); background: var(--blue-50); }
.tab-simple:hover:not(.active) { border-color: var(--gray-300); color: var(--gray-700); }
/* Copy button inside code blocks */
.copy-wrap { position: relative; }
.copy-btn { position: absolute; top: 8px; right: 8px; padding: 3px 10px; font-size: 0.72rem; font-weight: 600; border: 1px solid rgba(255,255,255,0.18); border-radius: 5px; background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.7); cursor: pointer; transition: all 0.15s; letter-spacing: 0.02em; }
.copy-btn:hover { background: rgba(255,255,255,0.2); color: #fff; }
.copy-btn.copied { background: rgba(34,197,94,0.25); color: #4ade80; border-color: rgba(74,222,128,0.3); }
/* Version / update */
.doc-head-right { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.version-pill {
  display: inline-flex; align-items: center; gap: 8px;
  background: var(--gray-50); border: 1px solid var(--gray-200);
  border-radius: 20px; padding: 5px 14px; font-size: 0.8rem; color: var(--gray-600);
}
.version-tag { font-weight: 700; color: var(--gray-800); font-family: 'SF Mono','Fira Code',monospace; }
.check-update-btn {
  padding: 3px 12px; font-size: 0.72rem; font-weight: 600;
  border: 1px solid var(--gray-300); border-radius: 12px;
  background: white; color: var(--gray-600); cursor: pointer; transition: all 0.15s;
}
.check-update-btn:hover:not(:disabled) { border-color: var(--blue); color: var(--blue); }
.check-update-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.update-badge { display: inline-flex; align-items: center; gap: 6px; padding: 5px 14px; border-radius: 20px; font-size: 0.78rem; font-weight: 600; }
.update-badge.latest { background: var(--green-50); color: var(--green); border: 1px solid rgba(34,197,94,0.2); }
.update-badge.available { background: #fff7ed; color: #c2410c; border: 1px solid #fed7aa; }
.update-badge.error { background: var(--red-50); color: var(--red); border: 1px solid rgba(239,68,68,0.2); }
.update-badge a { color: inherit; font-weight: 700; }
.one-click-update-btn { margin-left: 4px; padding: 3px 12px; font-size: 0.72rem; font-weight: 700; border: none; border-radius: 10px; background: #c2410c; color: white; cursor: pointer; transition: background 0.15s; }
.one-click-update-btn:hover { background: #9a3412; }
/* Update modal */
.update-overlay { background: rgba(0,0,0,0.6); backdrop-filter: blur(6px); }
.update-modal { max-width: 420px; text-align: center; padding: 44px 36px; }
.update-spinner { width: 48px; height: 48px; border: 3px solid var(--gray-200); border-top-color: var(--blue); border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 20px; }
.update-modal-title { font-size: 1.2rem; font-weight: 700; color: var(--gray-900); margin: 0 0 8px; }
.update-modal-desc { font-size: 0.875rem; color: var(--gray-500); margin: 0; line-height: 1.6; }
.update-steps { display: flex; justify-content: center; gap: 8px; margin-top: 20px; flex-wrap: wrap; }
.update-step { padding: 5px 16px; border-radius: 20px; font-size: 0.8rem; font-weight: 600; border: 1.5px solid var(--gray-200); color: var(--gray-400); transition: all 0.3s; }
.update-step.done { border-color: var(--green); color: var(--green); background: var(--green-50); }
.update-icon-wrap { width: 64px; height: 64px; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px; }
.update-icon-wrap.success { background: var(--green-50); color: var(--green); }
.update-icon-wrap.error { background: var(--red-50); color: var(--red); }

/* ====== Transitions ====== */
.fade-enter-active, .fade-leave-active { transition: opacity 0.25s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s ease; }
.modal-fade-enter-active .modal-box, .modal-fade-leave-active .modal-box { transition: transform 0.25s cubic-bezier(.4,0,.2,1), opacity 0.2s ease; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-from .modal-box { transform: scale(0.95) translateY(10px); opacity: 0; }
.slide-down-enter-active, .slide-down-leave-active { transition: all 0.18s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-6px); }
.toast-fade-enter-active, .toast-fade-leave-active { transition: all 0.3s ease; }
.toast-fade-enter-from, .toast-fade-leave-to { opacity: 0; transform: translateX(-50%) translateY(10px); }

/* ====== Responsive ====== */
@media (min-width: 992px) {
  .mobile-topbar { display: none !important; }
  .sidebar-overlay { display: none !important; }
  .sidebar { transform: translateX(0); }
  .table-wrap { display: block; }
  .mobile-list { display: none !important; }
}
@media (max-width: 991px) {
  .sidebar { transform: translateX(-100%); }
  .sidebar.open { transform: translateX(0); box-shadow: 4px 0 30px rgba(0,0,0,0.12); }
  .sidebar-overlay { display: block !important; }
  .mobile-topbar {
    display: block; position: sticky; top: 0; z-index: 100;
    background: rgba(255,255,255,0.95); backdrop-filter: blur(16px);
    border-bottom: 1px solid var(--gray-100);
  }
  .main-content { margin-left: 0; }
  .recipient-shell { grid-template-columns: 1fr; }
}
@media (max-width: 767px) {
  .table-wrap { display: none !important; }
  .mobile-list { display: block !important; }
  .section-head { flex-direction: column; align-items: flex-start; }
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .field-row { grid-template-columns: 1fr; }
  .modal-box { padding: 22px; border-radius: 16px; }
  .doc-grid, .doc-grid-2 { grid-template-columns: 1fr; }
  .recipient-link-action { max-width: none; }
  .recipient-shell { gap: 16px; }
  .recipient-panel-head, .recipient-workspace-hero, .recipient-toolbar, .recipient-panel-empty { padding-left: 18px; padding-right: 18px; }
  .recipient-panel-head, .recipient-toolbar, .recipient-workspace-hero { flex-direction: column; align-items: flex-start; }
  .recipient-group-card-badges { align-items: flex-start; }
  .recipient-kpi-grid { width: 100%; min-width: 0; }
  .recipient-toolbar-actions { width: 100%; }
  .recipient-toolbar-actions .btn-outline,
  .recipient-toolbar-actions .btn-primary { flex: 1; justify-content: center; }
  .recipient-mobile-item { padding: 16px 18px; }
}
@media (max-width: 479px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .stat-num { font-size: 1.5rem; }
  .login-card { padding: 26px 20px; }
  .recipient-overview { grid-template-columns: 1fr; }
  .recipient-kpi-grid { grid-template-columns: 1fr; }
  .recipient-toolbar-actions { flex-direction: column; }
  .recipient-toolbar-actions .btn-outline,
  .recipient-toolbar-actions .btn-primary { width: 100%; }
}
</style>
