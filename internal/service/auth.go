package service

import (
	"errors"
	"sync"
	"time"

	"smtp-lite/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// loginAttempt 登录尝试记录（暴力破解防护）
type loginAttempt struct {
	count    int
	lockedAt time.Time
}

type AuthService struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

const (
	maxLoginAttempts  = 5                // 最大连续失败次数
	loginLockDuration = 15 * time.Minute // 锁定时长
)

func NewAuthService() *AuthService {
	return &AuthService{
		attempts: make(map[string]*loginAttempt),
	}
}

type Claims struct {
	Username     string `json:"username"`
	TokenVersion int64  `json:"token_version"`
	jwt.RegisteredClaims
}

// checkLoginLock 检查是否因多次失败而锁定
func (s *AuthService) checkLoginLock(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	attempt, exists := s.attempts[username]
	if !exists {
		return nil
	}

	// 如果已锁定且未过期
	if attempt.count >= maxLoginAttempts {
		if time.Since(attempt.lockedAt) < loginLockDuration {
			return errors.New("登录失败次数过多，请稍后再试")
		}
		// 锁定已过期，重置
		delete(s.attempts, username)
	}
	return nil
}

// recordLoginFailure 记录登录失败
func (s *AuthService) recordLoginFailure(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attempt, exists := s.attempts[username]
	if !exists {
		attempt = &loginAttempt{}
		s.attempts[username] = attempt
	}
	attempt.count++
	if attempt.count >= maxLoginAttempts {
		attempt.lockedAt = time.Now()
	}
}

// clearLoginAttempts 登录成功后清除记录
func (s *AuthService) clearLoginAttempts(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.attempts, username)
}

func (s *AuthService) Login(username, password string) (string, error) {
	// 检查是否被锁定
	if err := s.checkLoginLock(username); err != nil {
		return "", err
	}

	cfg := config.Get()

	// 验证用户名
	if username != cfg.Auth.Username {
		s.recordLoginFailure(username)
		return "", errors.New("invalid credentials")
	}

	// 使用 bcrypt 验证密码
	if !config.VerifyPassword(password) {
		s.recordLoginFailure(username)
		return "", errors.New("invalid credentials")
	}

	// 登录成功，清除失败记录
	s.clearLoginAttempts(username)

	// 生成 JWT
	expireHours := cfg.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}

	now := time.Now()
	claims := &Claims{
		Username:     username,
		TokenVersion: config.GetTokenVersion(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "smtp-lite",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ChangePassword 修改登录密码
func (s *AuthService) ChangePassword(oldPassword, newPassword string) error {
	// 使用 bcrypt 验证旧密码
	if !config.VerifyPassword(oldPassword) {
		return errors.New("旧密码不正确")
	}
	if len(newPassword) < 8 {
		return errors.New("新密码至少需要 8 位字符")
	}
	return config.UpdateAuthPassword(newPassword)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保签名算法正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查 token 版本号（密码修改后旧 token 失效）
		if claims.TokenVersion < config.GetTokenVersion() {
			return nil, errors.New("token has been revoked")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
