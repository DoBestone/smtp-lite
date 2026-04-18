import client from './client'
import type {
  ApiKey,
  BlacklistEntry,
  LoginResponse,
  LogsResponse,
  QueueStats,
  Recipient,
  RecipientGroup,
  SendBatchPayload,
  SendEmailPayload,
  SendResult,
  SendScheduledPayload,
  SmtpAccount,
  SmtpAccountPayload,
  Stats,
  Template,
  TemplatePayload,
  UpdateCheck,
  UpdatePrepareResp,
  VersionInfo,
  Webhook
} from './types'

// ---------- Auth ----------
export const authApi = {
  login: (username: string, password: string) =>
    client.post<LoginResponse>('/auth/login', { username, password }).then((r) => r.data),
  changePassword: (oldPassword: string, newPassword: string) =>
    client
      .post('/auth/change-password', { old_password: oldPassword, new_password: newPassword })
      .then((r) => r.data),
  changeUsername: (newUsername: string, password: string) =>
    client
      .post('/auth/change-username', { new_username: newUsername, password })
      .then((r) => r.data)
}

// ---------- SMTP Accounts ----------
export const smtpApi = {
  list: () => client.get<SmtpAccount[]>('/smtp-accounts').then((r) => r.data ?? []),
  create: (data: SmtpAccountPayload) =>
    client.post<SmtpAccount>('/smtp-accounts', data).then((r) => r.data),
  update: (id: string, data: Partial<SmtpAccountPayload>) =>
    client.put<SmtpAccount>(`/smtp-accounts/${id}`, data).then((r) => r.data),
  remove: (id: string) => client.delete(`/smtp-accounts/${id}`).then((r) => r.data),
  toggle: (id: string) => client.post(`/smtp-accounts/${id}/toggle`).then((r) => r.data),
  test: (id: string) =>
    client.post<{ success: boolean; error?: string }>(`/smtp-accounts/${id}/test`).then((r) => r.data)
}

// ---------- API Keys ----------
export const apiKeysApi = {
  list: () => client.get<ApiKey[]>('/api-keys').then((r) => r.data ?? []),
  create: (name: string) => client.post<ApiKey>('/api-keys', { name }).then((r) => r.data),
  reset: (id: string) => client.post<ApiKey>(`/api-keys/${id}/reset`).then((r) => r.data),
  remove: (id: string) => client.delete(`/api-keys/${id}`).then((r) => r.data)
}

// ---------- Templates ----------
export const templatesApi = {
  list: () => client.get<Template[]>('/templates').then((r) => r.data ?? []),
  create: (data: TemplatePayload) =>
    client.post<Template>('/templates', data).then((r) => r.data),
  update: (id: string, data: Partial<TemplatePayload>) =>
    client.put<Template>(`/templates/${id}`, data).then((r) => r.data),
  remove: (id: string) => client.delete(`/templates/${id}`).then((r) => r.data)
}

// ---------- Recipient Groups & Recipients ----------
export const recipientsApi = {
  listGroups: () =>
    client.get<RecipientGroup[]>('/recipient-groups').then((r) => r.data ?? []),
  createGroup: (data: { name: string; description?: string }) =>
    client.post<RecipientGroup>('/recipient-groups', data).then((r) => r.data),
  updateGroup: (id: string, data: Partial<{ name: string; description: string }>) =>
    client.put<RecipientGroup>(`/recipient-groups/${id}`, data).then((r) => r.data),
  removeGroup: (id: string) => client.delete(`/recipient-groups/${id}`).then((r) => r.data),
  listByGroup: (groupId: string) =>
    client
      .get<Recipient[]>(`/recipients`, { params: { group_id: groupId } })
      .then((r) => r.data ?? []),
  create: (data: { group_id: string; email: string; name?: string }) =>
    client.post<Recipient>('/recipients', data).then((r) => r.data),
  remove: (id: string) => client.delete(`/recipients/${id}`).then((r) => r.data),
  batchImport: (groupId: string, emails: string) =>
    client
      .post<{ imported: number; skipped: number }>('/recipients/batch', {
        group_id: groupId,
        emails
      })
      .then((r) => r.data)
}

// ---------- Webhooks ----------
export interface WebhookCreatePayload {
  name: string
  url: string
  secret?: string
  events: string
  enabled?: boolean
}
export const webhooksApi = {
  list: () => client.get<Webhook[]>('/webhooks').then((r) => r.data ?? []),
  create: (data: WebhookCreatePayload) =>
    client.post<Webhook>('/webhooks', data).then((r) => r.data),
  toggle: (id: string) => client.post(`/webhooks/${id}/toggle`).then((r) => r.data),
  test: (id: string) => client.post(`/webhooks/${id}/test`).then((r) => r.data),
  remove: (id: string) => client.delete(`/webhooks/${id}`).then((r) => r.data)
}

// ---------- Blacklist ----------
export const blacklistApi = {
  list: () => client.get<BlacklistEntry[]>('/blacklist').then((r) => r.data ?? []),
  create: (data: { email: string; reason?: string }) =>
    client.post<BlacklistEntry>('/blacklist', data).then((r) => r.data),
  remove: (id: string) => client.delete(`/blacklist/${id}`).then((r) => r.data)
}

// ---------- Logs ----------
export const logsApi = {
  list: (page = 1, pageSize = 50) =>
    client
      .get<LogsResponse>('/send/logs', { params: { page, page_size: pageSize } })
      .then((r) => r.data),
  exportCsv: (params?: Record<string, string | number>) =>
    client
      .get<Blob>('/send/logs/export', { params, responseType: 'blob' })
      .then((r) => r.data)
}

// ---------- Stats ----------
export const statsApi = {
  get: () => client.get<Stats>('/stats').then((r) => r.data),
  queue: () => client.get<QueueStats>('/queue/stats').then((r) => r.data)
}

// ---------- Send ----------
export const sendApi = {
  send: (data: SendEmailPayload) =>
    client.post<SendResult>('/send', data).then((r) => r.data),
  sendBatch: (data: SendBatchPayload) =>
    client.post<SendResult>('/send/batch', data).then((r) => r.data),
  sendScheduled: (data: SendScheduledPayload) =>
    client.post<SendResult>('/send/scheduled', data).then((r) => r.data)
}

// ---------- System ----------
export const systemApi = {
  version: () => client.get<VersionInfo>('/version').then((r) => r.data),
  updateCheck: () => client.get<UpdateCheck>('/system/update-check').then((r) => r.data),
  changelog: () => client.get<{ changelog: string }>('/system/changelog').then((r) => r.data),
  updatePrepare: () =>
    client.post<UpdatePrepareResp>('/system/update-prepare').then((r) => r.data),
  update: (confirmToken: string) =>
    client
      .post<{ message: string }>('/system/update', { confirm_token: confirmToken })
      .then((r) => r.data),
  exportAccountsUrl: '/api/v1/export/accounts'
}
