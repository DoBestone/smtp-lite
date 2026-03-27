package config

import (
	"os"
	"sync"

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
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type EncryptionConfig struct {
	Key string `yaml:"key"`
}

type RateLimitConfig struct {
	Enabled       bool `yaml:"enabled"`        // 是否启用限流
	GlobalLimit   int  `yaml:"global_limit"`   // 全局每分钟限制
	AccountLimit  int  `yaml:"account_limit"`  // 单账号每分钟限制
}

type TrackConfig struct {
	Enabled     bool   `yaml:"enabled"`      // 是否启用邮件追踪
	TrackDomain string `yaml:"track_domain"` // 追踪域名（用于追踪链接）
}

type LocaleConfig struct {
	Default string `yaml:"default"` // 默认语言 zh-CN / en-US
}

var (
	cfg  *Config
	once sync.Once
	mu   sync.Mutex
)

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
				GlobalLimit:  100,  // 全局每分钟100封
				AccountLimit: 30,   // 单账号每分钟30封
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

// UpdateAuthPassword 更新登录密码并持久化到 config.yaml
func UpdateAuthPassword(newPassword string) error {
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		Load()
	}
	cfg.Auth.Password = newPassword
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