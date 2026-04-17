// ============================================================
// API 类型定义 · 与后端 internal/model/model.go 保持一致
// ============================================================

export interface LoginResponse {
  token: string
  username?: string
}

export type SmtpStatus = 'active' | 'disabled' | string

export interface SmtpAccount {
  id: string
  email: string
  smtp_host: string
  smtp_port: number
  daily_limit: number
  daily_used: number
  last_reset_date?: string
  status: SmtpStatus
  last_error?: string
  priority: number
  created_at: string
  updated_at: string
}

export interface SmtpAccountPayload {
  email: string
  password: string
  smtp_host: string
  smtp_port?: number
  daily_limit?: number
  priority?: number
}

export interface ApiKey {
  id: string
  name: string
  key?: string
  key_prefix: string
  last_used_at?: string
  created_at: string
}

export interface Template {
  id: string
  name: string
  subject: string
  body: string
  is_html: boolean
  description?: string
  created_at: string
  updated_at: string
}

export type TemplatePayload = Omit<Template, 'id' | 'created_at' | 'updated_at'>

export interface RecipientGroup {
  id: string
  name: string
  description?: string
  count: number
  created_at: string
  updated_at: string
}

export type RecipientStatus = 'active' | 'blacklisted' | string

export interface Recipient {
  id: string
  group_id: string
  email: string
  name?: string
  status: RecipientStatus
  created_at: string
}

export interface Webhook {
  id: string
  name: string
  url: string
  events: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface BlacklistEntry {
  id: string
  email: string
  reason?: string
  created_at: string
}

export type SendLogStatus = 'sent' | 'success' | 'failed' | 'pending' | string

export interface SendLog {
  id: string
  smtp_account_id?: string
  to_email: string
  subject: string
  status: SendLogStatus
  error_message?: string
  opened: boolean
  opened_at?: string
  clicked: boolean
  clicked_at?: string
  track_id?: string
  batch_id?: string
  created_at: string
}

export interface LogsResponse {
  logs: SendLog[]
  total: number
  page: number
  page_size: number
}

export interface Stats {
  total_sent: number
  success: number
  failed: number
  today_sent: number
  success_rate: number
  opened: number
  clicked: number
  open_rate: number
  click_rate: number
}

export interface QueueStats {
  pending: number
  processing: number
  sent: number
  failed: number
}

export interface VersionInfo {
  version: string
  build_time?: string
  commit?: string
}

export interface UpdateCheck {
  has_update: boolean
  latest_version?: string
  current_version?: string
  force_update?: boolean
  release_notes?: string
}

export interface SendEmailPayload {
  to: string
  subject: string
  body: string
  is_html?: boolean
  from_name?: string
  cc?: string
  bcc?: string
  track_enabled?: boolean
  attachments?: Array<{ filename: string; content: string; size?: number }>
}

export interface SendBatchPayload {
  emails: string[]
  subject: string
  body: string
  is_html?: boolean
  from_name?: string
  track_enabled?: boolean
  attachments?: SendEmailPayload['attachments']
}

export interface SendScheduledPayload extends SendEmailPayload {
  scheduled_at: string
}

export interface SendResult {
  success: boolean
  message: string
  details?: string[]
}
