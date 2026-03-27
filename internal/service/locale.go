package service

import (
	"smtp-lite/internal/config"
	"sync"
)

// 多语言文本
var translations = map[string]map[string]string{
	"zh-CN": {
		// 通用
		"app.title":        "SMTP Lite",
		"app.subtitle":     "个人邮箱聚合系统",
		"login":            "登录",
		"logout":           "退出",
		"username":         "用户名",
		"password":         "密码",
		"submit":           "提交",
		"cancel":           "取消",
		"save":             "保存",
		"delete":           "删除",
		"edit":             "编辑",
		"add":              "添加",
		"search":           "搜索",
		"refresh":          "刷新",
		"success":          "成功",
		"failed":           "失败",
		"loading":          "加载中...",
		"confirm":          "确认",
		"confirm.delete":   "确定删除？",
		"operation.success": "操作成功",
		"operation.failed":  "操作失败",

		// 导航
		"nav.smtp":      "SMTP 账号",
		"nav.apikeys":   "API Key",
		"nav.logs":      "发送日志",
		"nav.stats":     "统计",
		"nav.templates": "邮件模板",
		"nav.recipients": "收件人",
		"nav.blacklist": "黑名单",
		"nav.webhooks":  "Webhook",
		"nav.queue":     "发送队列",
		"nav.settings":  "设置",

		// SMTP 账号
		"smtp.title":       "SMTP 账号",
		"smtp.email":       "邮箱",
		"smtp.host":        "SMTP 服务器",
		"smtp.port":        "端口",
		"smtp.password":    "密码/授权码",
		"smtp.daily_limit": "日限额",
		"smtp.daily_used":  "已发送",
		"smtp.status":      "状态",
		"smtp.active":      "启用",
		"smtp.disabled":    "禁用",
		"smtp.test":        "测试连接",
		"smtp.toggle":      "启用/禁用",
		"smtp.add":         "添加账号",
		"smtp.empty":       "暂无 SMTP 账号",

		// API Key
		"apikey.title":      "API Key",
		"apikey.name":       "名称",
		"apikey.prefix":     "前缀",
		"apikey.last_used":  "最后使用",
		"apikey.created_at": "创建时间",
		"apikey.add":        "创建 Key",
		"apikey.warning":    "请保存此 Key，它只会显示一次！",
		"apikey.empty":      "暂无 API Key",

		// 发送日志
		"log.title":      "发送日志",
		"log.to":         "收件人",
		"log.subject":    "主题",
		"log.status":     "状态",
		"log.time":       "时间",
		"log.opened":     "已打开",
		"log.clicked":    "已点击",
		"log.empty":      "暂无发送记录",

		// 统计
		"stats.title":       "统计信息",
		"stats.total_sent":  "总发送量",
		"stats.success":     "成功",
		"stats.failed":      "失败",
		"stats.today":       "今日发送",
		"stats.success_rate": "成功率",
		"stats.open_rate":   "打开率",
		"stats.click_rate":  "点击率",

		// 模板
		"template.title":   "邮件模板",
		"template.name":    "模板名称",
		"template.subject": "主题",
		"template.body":    "内容",
		"template.html":    "HTML 格式",
		"template.add":     "添加模板",
		"template.empty":   "暂无模板",

		// 收件人
		"recipient.title":   "收件人分组",
		"recipient.group":   "分组",
		"recipient.email":   "邮箱",
		"recipient.name":    "名称",
		"recipient.count":   "数量",
		"recipient.add":     "添加收件人",
		"recipient.import":  "批量导入",
		"recipient.empty":   "暂无收件人",

		// 黑名单
		"blacklist.title":  "黑名单",
		"blacklist.email":  "邮箱",
		"blacklist.reason": "原因",
		"blacklist.add":    "添加黑名单",
		"blacklist.empty":  "暂无黑名单",

		// Webhook
		"webhook.title":  "Webhook",
		"webhook.name":   "名称",
		"webhook.url":    "URL",
		"webhook.events": "事件",
		"webhook.enabled": "启用",
		"webhook.add":    "添加 Webhook",
		"webhook.empty":  "暂无 Webhook",

		// 队列
		"queue.title":    "发送队列",
		"queue.pending":  "待发送",
		"queue.sent":     "已发送",
		"queue.failed":   "发送失败",
		"queue.scheduled": "定时发送",

		// 发送
		"send.title":     "发送邮件",
		"send.to":        "收件人",
		"send.cc":        "抄送",
		"send.bcc":       "密送",
		"send.subject":   "主题",
		"send.body":      "正文",
		"send.html":      "HTML 格式",
		"send.from_name": "发件人名称",
		"send.attachment": "附件",
		"send.track":     "启用追踪",
		"send.schedule":  "定时发送",
		"send.send":      "发送",
		"send.batch":     "批量发送",
		"send.success":   "邮件发送成功",
		"send.failed":    "邮件发送失败",
	},

	"en-US": {
		// Common
		"app.title":        "SMTP Lite",
		"app.subtitle":     "Personal SMTP Aggregator",
		"login":            "Login",
		"logout":           "Logout",
		"username":         "Username",
		"password":         "Password",
		"submit":           "Submit",
		"cancel":           "Cancel",
		"save":             "Save",
		"delete":           "Delete",
		"edit":             "Edit",
		"add":              "Add",
		"search":           "Search",
		"refresh":          "Refresh",
		"success":          "Success",
		"failed":           "Failed",
		"loading":          "Loading...",
		"confirm":          "Confirm",
		"confirm.delete":   "Are you sure to delete?",
		"operation.success": "Operation successful",
		"operation.failed":  "Operation failed",

		// Navigation
		"nav.smtp":      "SMTP Accounts",
		"nav.apikeys":   "API Keys",
		"nav.logs":      "Send Logs",
		"nav.stats":     "Statistics",
		"nav.templates": "Templates",
		"nav.recipients": "Recipients",
		"nav.blacklist": "Blacklist",
		"nav.webhooks":  "Webhooks",
		"nav.queue":     "Send Queue",
		"nav.settings":  "Settings",

		// SMTP Accounts
		"smtp.title":       "SMTP Accounts",
		"smtp.email":       "Email",
		"smtp.host":        "SMTP Server",
		"smtp.port":        "Port",
		"smtp.password":    "Password/App Password",
		"smtp.daily_limit": "Daily Limit",
		"smtp.daily_used":  "Sent Today",
		"smtp.status":      "Status",
		"smtp.active":      "Active",
		"smtp.disabled":    "Disabled",
		"smtp.test":        "Test Connection",
		"smtp.toggle":      "Enable/Disable",
		"smtp.add":         "Add Account",
		"smtp.empty":       "No SMTP accounts",

		// API Keys
		"apikey.title":      "API Keys",
		"apikey.name":       "Name",
		"apikey.prefix":     "Prefix",
		"apikey.last_used":  "Last Used",
		"apikey.created_at": "Created At",
		"apikey.add":        "Create Key",
		"apikey.warning":    "Save this key! It won't be shown again.",
		"apikey.empty":      "No API keys",

		// Send Logs
		"log.title":      "Send Logs",
		"log.to":         "Recipient",
		"log.subject":    "Subject",
		"log.status":     "Status",
		"log.time":       "Time",
		"log.opened":     "Opened",
		"log.clicked":    "Clicked",
		"log.empty":      "No send logs",

		// Statistics
		"stats.title":       "Statistics",
		"stats.total_sent":  "Total Sent",
		"stats.success":     "Success",
		"stats.failed":      "Failed",
		"stats.today":       "Today",
		"stats.success_rate": "Success Rate",
		"stats.open_rate":   "Open Rate",
		"stats.click_rate":  "Click Rate",

		// Templates
		"template.title":   "Email Templates",
		"template.name":    "Template Name",
		"template.subject": "Subject",
		"template.body":    "Body",
		"template.html":    "HTML Format",
		"template.add":     "Add Template",
		"template.empty":   "No templates",

		// Recipients
		"recipient.title":   "Recipient Groups",
		"recipient.group":   "Group",
		"recipient.email":   "Email",
		"recipient.name":    "Name",
		"recipient.count":   "Count",
		"recipient.add":     "Add Recipient",
		"recipient.import":  "Batch Import",
		"recipient.empty":   "No recipients",

		// Blacklist
		"blacklist.title":  "Blacklist",
		"blacklist.email":  "Email",
		"blacklist.reason": "Reason",
		"blacklist.add":    "Add to Blacklist",
		"blacklist.empty":  "No blacklist entries",

		// Webhook
		"webhook.title":  "Webhooks",
		"webhook.name":   "Name",
		"webhook.url":    "URL",
		"webhook.events": "Events",
		"webhook.enabled": "Enabled",
		"webhook.add":    "Add Webhook",
		"webhook.empty":  "No webhooks",

		// Queue
		"queue.title":    "Send Queue",
		"queue.pending":  "Pending",
		"queue.sent":     "Sent",
		"queue.failed":   "Failed",
		"queue.scheduled": "Scheduled",

		// Send
		"send.title":     "Send Email",
		"send.to":        "To",
		"send.cc":        "CC",
		"send.bcc":       "BCC",
		"send.subject":   "Subject",
		"send.body":      "Body",
		"send.html":      "HTML Format",
		"send.from_name": "From Name",
		"send.attachment": "Attachment",
		"send.track":     "Enable Tracking",
		"send.schedule":  "Schedule Send",
		"send.send":      "Send",
		"send.batch":     "Batch Send",
		"send.success":   "Email sent successfully",
		"send.failed":    "Failed to send email",
	},
}

type LocaleService struct {
	currentLocale string
	mu            sync.RWMutex
}

func NewLocaleService() *LocaleService {
	cfg := config.Get()
	locale := cfg.Locale.Default
	if locale == "" {
		locale = "zh-CN"
	}
	return &LocaleService{currentLocale: locale}
}

// Get 获取翻译文本
func (s *LocaleService) Get(key string) string {
	return s.GetWithLocale(key, s.currentLocale)
}

// GetWithLocale 获取指定语言的翻译
func (s *LocaleService) GetWithLocale(key, locale string) string {
	if texts, ok := translations[locale]; ok {
		if text, ok := texts[key]; ok {
			return text
		}
	}
	// 回退到中文
	if texts, ok := translations["zh-CN"]; ok {
		if text, ok := texts[key]; ok {
			return text
		}
	}
	return key
}

// SetLocale 设置当前语言
func (s *LocaleService) SetLocale(locale string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentLocale = locale
}

// GetLocale 获取当前语言
func (s *LocaleService) GetLocale() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLocale
}

// GetAvailableLocales 获取可用语言列表
func (s *LocaleService) GetAvailableLocales() []string {
	locales := []string{}
	for locale := range translations {
		locales = append(locales, locale)
	}
	return locales
}

// GetAllTranslations 获取所有翻译
func (s *LocaleService) GetAllTranslations(locale string) map[string]string {
	if texts, ok := translations[locale]; ok {
		return texts
	}
	return translations["zh-CN"]
}