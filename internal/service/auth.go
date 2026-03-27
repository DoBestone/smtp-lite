package service

import (
	"errors"
	"time"

	"smtp-lite/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(username, password string) (string, error) {
	cfg := config.Get()

	// 验证用户名密码
	if username != cfg.Auth.Username || password != cfg.Auth.Password {
		return "", errors.New("invalid credentials")
	}

	// 生成 JWT
	expireHours := cfg.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ChangePassword 修改登录密码
func (s *AuthService) ChangePassword(oldPassword, newPassword string) error {
	cfg := config.Get()
	if oldPassword != cfg.Auth.Password {
		return errors.New("旧密码不正确")
	}
	if len(newPassword) < 6 {
		return errors.New("新密码至少需要 6 位字符")
	}
	return config.UpdateAuthPassword(newPassword)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
