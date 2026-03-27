package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Auth      AuthConfig      `yaml:"auth"`
	JWT       JWTConfig       `yaml:"jwt"`
	Encryption EncryptionConfig `yaml:"encryption"`
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

var (
	cfg  *Config
	once sync.Once
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