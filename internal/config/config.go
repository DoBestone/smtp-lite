package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Auth       AuthConfig       `yaml:"auth"`
	JWT        JWTConfig        `yaml:"jwt"`
	Encryption EncryptionConfig `yaml:"encryption"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Track      TrackConfig      `yaml:"track"`
	Locale     LocaleConfig     `yaml:"locale"`
	CORS       CORSConfig       `yaml:"cors"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"` // 支持明文或 bcrypt hash（$2a$ 开头）
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type EncryptionConfig struct {
	Key string `yaml:"key"`
}

type RateLimitConfig struct {
	Enabled      bool `yaml:"enabled"`       // 是否启用限流
	GlobalLimit  int  `yaml:"global_limit"`  // 全局每分钟限制
	AccountLimit int  `yaml:"account_limit"` // 单账号每分钟限制
}

type TrackConfig struct {
	Enabled     bool   `yaml:"enabled"`      // 是否启用邮件追踪
	TrackDomain string `yaml:"track_domain"` // 追踪域名（用于追踪链接）
}

type LocaleConfig struct {
	Default string `yaml:"default"` // 默认语言 zh-CN / en-US
}

type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"` // 允许的域名列表，为空则不允许跨域
}

// tokenVersion 用于密码修改后使已有 token 失效，原子操作安全
var tokenVersion int64

var (
	cfg  *Config
	once sync.Once
	mu   sync.Mutex
)

// 不安全的默认值列表
var insecureDefaults = []string{
	"default-secret-change-me",
	"change-this-to-random-32-byte-string",
	"smtp-lite-default-encryption-32!",
	"smtp-lite-encryption-key-32b!",
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("[config] 无法生成随机密钥: %v", err)
	}
	return hex.EncodeToString(b)[:length]
}

// isInsecureDefault 检查是否为不安全的默认值
func isInsecureDefault(value string) bool {
	for _, d := range insecureDefaults {
		if value == d {
			return true
		}
	}
	return false
}

// Load 加载配置文件
func Load() *Config {
	once.Do(func() {
		cfg = &Config{
			Server: ServerConfig{
				Port: 8090,
				Mode: "release",
			},
			Auth: AuthConfig{
				Username: "admin",
				Password: "admin123",
			},
			JWT: JWTConfig{
				Secret:      "default-secret-change-me",
				ExpireHours: 168,
			},
			Encryption: EncryptionConfig{
				Key: "smtp-lite-default-encryption-32!",
			},
			RateLimit: RateLimitConfig{
				Enabled:      true,
				GlobalLimit:  100, // 全局每分钟100封
				AccountLimit: 30,  // 单账号每分钟30封
			},
			Track: TrackConfig{
				Enabled:     false,
				TrackDomain: "",
			},
			Locale: LocaleConfig{
				Default: "zh-CN",
			},
		}

		// 读取配置文件
		data, err := os.ReadFile("config.yaml")
		if err == nil {
			yaml.Unmarshal(data, cfg)
		}

		// 环境变量覆盖
		if v := os.Getenv("SMTP_USERNAME"); v != "" {
			cfg.Auth.Username = v
		}
		if v := os.Getenv("SMTP_PASSWORD"); v != "" {
			cfg.Auth.Password = v
		}
		if v := os.Getenv("SMTP_JWT_SECRET"); v != "" {
			cfg.JWT.Secret = v
		}
		if v := os.Getenv("SMTP_ENCRYPTION_KEY"); v != "" {
			cfg.Encryption.Key = v
		}
		if v := os.Getenv("SMTP_PORT"); v != "" {
			cfg.Server.Port = 0
			for _, c := range v {
				if c >= '0' && c <= '9' {
					cfg.Server.Port = cfg.Server.Port*10 + int(c-'0')
				}
			}
		}
		if v := os.Getenv("SMTP_RATE_LIMIT_GLOBAL"); v != "" {
			cfg.RateLimit.GlobalLimit = 0
			for _, c := range v {
				if c >= '0' && c <= '9' {
					cfg.RateLimit.GlobalLimit = cfg.RateLimit.GlobalLimit*10 + int(c-'0')
				}
			}
		}

		// 安全检查：自动替换不安全的默认密钥
		configChanged := false
		if isInsecureDefault(cfg.JWT.Secret) {
			cfg.JWT.Secret = generateRandomKey(32)
			log.Println("[security] JWT Secret 为默认值，已自动生成随机密钥")
			configChanged = true
		}
		if isInsecureDefault(cfg.Encryption.Key) {
			cfg.Encryption.Key = generateRandomKey(32)
			log.Println("[security] Encryption Key 为默认值，已自动生成随机密钥")
			configChanged = true
		}

		// 密码 bcrypt 自动迁移：如果密码不是 bcrypt hash 格式，自动加密并写回
		if !strings.HasPrefix(cfg.Auth.Password, "$2a$") && !strings.HasPrefix(cfg.Auth.Password, "$2b$") {
			hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.Auth.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("[config] 无法对密码进行 bcrypt 加密: %v", err)
			}
			cfg.Auth.Password = string(hashed)
			log.Println("[security] 密码已自动迁移为 bcrypt hash")
			configChanged = true
		}

		// 如果有变更，回写配置文件
		if configChanged {
			if yamlData, err := yaml.Marshal(cfg); err == nil {
				os.WriteFile("config.yaml", yamlData, 0600)
			}
		}
	})
	return cfg
}

// Get 获取配置实例
func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}

// GetTokenVersion 获取当前 token 版本号
func GetTokenVersion() int64 {
	return atomic.LoadInt64(&tokenVersion)
}

// IncrementTokenVersion 递增 token 版本号（密码修改后调用）
func IncrementTokenVersion() {
	atomic.AddInt64(&tokenVersion, 1)
}

// VerifyPassword 使用 bcrypt 验证密码
func VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(cfg.Auth.Password), []byte(password)) == nil
}

// UpdateAuthPassword 更新登录密码并持久化到 config.yaml
func UpdateAuthPassword(newPassword string) error {
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		Load()
	}
	// 使用 bcrypt 加密新密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	cfg.Auth.Password = string(hashed)
	// 递增 token 版本号，使旧 token 失效
	IncrementTokenVersion()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile("config.yaml", data, 0600)
}

// UpdateLocale 更新语言配置
func UpdateLocale(locale string) error {
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		Load()
	}
	cfg.Locale.Default = locale
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile("config.yaml", data, 0600)
}
