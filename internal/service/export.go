package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"smtp-lite/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExportService struct {
	db *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{db: db}
}

// ExportSendLogsCSV 导出发送日志为CSV
func (s *ExportService) ExportSendLogsCSV(start, end *time.Time) ([]byte, error) {
	query := s.db.Model(&model.SendLog{}).Order("created_at desc")

	if start != nil {
		query = query.Where("created_at >= ?", start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", end)
	}

	var logs []model.SendLog
	if err := query.Limit(50000).Find(&logs).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"ID", "收件人", "主题", "状态", "错误信息", "打开", "点击", "创建时间"})

	// 写入数据
	for _, log := range logs {
		writer.Write([]string{
			log.ID.String(),
			log.ToEmail,
			log.Subject,
			log.Status,
			log.ErrorMessage,
			fmt.Sprintf("%v", log.Opened),
			fmt.Sprintf("%v", log.Clicked),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportSendLogsJSON 导出发送日志为JSON
func (s *ExportService) ExportSendLogsJSON(start, end *time.Time) ([]byte, error) {
	query := s.db.Model(&model.SendLog{}).Order("created_at desc")

	if start != nil {
		query = query.Where("created_at >= ?", start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", end)
	}

	var logs []model.SendLog
	if err := query.Limit(50000).Find(&logs).Error; err != nil {
		return nil, err
	}

	return json.MarshalIndent(logs, "", "  ")
}

// ExportSmtpAccountsCSV 导出SMTP账号为CSV
func (s *ExportService) ExportSmtpAccountsCSV() ([]byte, error) {
	var accounts []model.SmtpAccount
	if err := s.db.Order("created_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"ID", "邮箱", "SMTP服务器", "端口", "日限额", "已用", "状态", "创建时间"})

	// 写入数据
	for _, acc := range accounts {
		writer.Write([]string{
			acc.ID.String(),
			acc.Email,
			acc.SmtpHost,
			fmt.Sprintf("%d", acc.SmtpPort),
			fmt.Sprintf("%d", acc.DailyLimit),
			fmt.Sprintf("%d", acc.DailyUsed),
			acc.Status,
			acc.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportApiKeysCSV 导出API Key为CSV
func (s *ExportService) ExportApiKeysCSV() ([]byte, error) {
	var keys []model.APIKey
	if err := s.db.Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"ID", "名称", "前缀", "最后使用", "创建时间"})

	// 写入数据
	for _, key := range keys {
		lastUsed := ""
		if key.LastUsedAt != nil {
			lastUsed = key.LastUsedAt.Format("2006-01-02 15:04:05")
		}
		writer.Write([]string{
			key.ID.String(),
			key.Name,
			key.KeyPrefix,
			lastUsed,
			key.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportRecipientsCSV 导出收件人为CSV
func (s *ExportService) ExportRecipientsCSV(groupID *uuid.UUID) ([]byte, error) {
	query := s.db.Model(&model.Recipient{})
	if groupID != nil {
		query = query.Where("group_id = ?", groupID)
	}

	var recipients []model.Recipient
	if err := query.Order("created_at desc").Find(&recipients).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"邮箱", "名称", "状态", "创建时间"})

	// 写入数据
	for _, r := range recipients {
		writer.Write([]string{
			r.Email,
			r.Name,
			r.Status,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportBlacklistCSV 导出黑名单为CSV
func (s *ExportService) ExportBlacklistCSV() ([]byte, error) {
	var list []model.Blacklist
	if err := s.db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"邮箱", "原因", "创建时间"})

	// 写入数据
	for _, item := range list {
		writer.Write([]string{
			item.Email,
			item.Reason,
			item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportTemplatesCSV 导出模板为CSV
func (s *ExportService) ExportTemplatesCSV() ([]byte, error) {
	var templates []model.EmailTemplate
	if err := s.db.Order("created_at desc").Find(&templates).Error; err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入头部
	writer.Write([]string{"名称", "主题", "内容", "HTML", "描述", "创建时间"})

	// 写入数据
	for _, t := range templates {
		writer.Write([]string{
			t.Name,
			t.Subject,
			t.Body,
			fmt.Sprintf("%v", t.IsHTML),
			t.Description,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportReport 导出综合报告
func (s *ExportService) ExportReport(start, end *time.Time) ([]byte, error) {
	report := make(map[string]interface{})

	// 发送统计
	var totalSent, successCount, failedCount int64
	query := s.db.Model(&model.SendLog{})
	if start != nil {
		query = query.Where("created_at >= ?", start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", end)
	}
	query.Count(&totalSent)

	s.db.Model(&model.SendLog{}).Where("status = ?", "success").Count(&successCount)
	s.db.Model(&model.SendLog{}).Where("status = ?", "failed").Count(&failedCount)

	report["total_sent"] = totalSent
	report["success"] = successCount
	report["failed"] = failedCount

	var successRate float64
	if totalSent > 0 {
		successRate = float64(successCount) / float64(totalSent) * 100
	}
	report["success_rate"] = successRate

	// 账号统计
	var totalAccounts, activeAccounts int64
	s.db.Model(&model.SmtpAccount{}).Count(&totalAccounts)
	s.db.Model(&model.SmtpAccount{}).Where("status = ?", "active").Count(&activeAccounts)
	report["total_accounts"] = totalAccounts
	report["active_accounts"] = activeAccounts

	// 时间范围
	report["start_time"] = start
	report["end_time"] = end
	report["generated_at"] = time.Now()

	return json.MarshalIndent(report, "", "  ")
}