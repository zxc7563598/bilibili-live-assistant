package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zxc7563598/GoAdminKit/internal/config"
)

var (
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
)

type Claims struct {
	ID       uint64 `json:"id"`
	Type     string `json:"type"`     // access / refresh
	Identity string `json:"identity"` // admin / user
	RoleID   uint64 `json:"role_id"`  // 0 and role_id
	RoleCode string `json:"role_code"`
	jwt.RegisteredClaims
}

// Init 初始化 JWT
func Init(c config.JWTConfig) {
	secret = []byte(c.Secret)
	accessTokenTTL = time.Duration(c.AccessTTL) * time.Second
	refreshTokenTTL = time.Duration(c.RefreshTTL) * time.Second
}

// GenerateAccessToken 生成 AccessToken
func GenerateAccessToken(id uint64, identity string, role uint64, roleCode string) (string, error) {
	now := time.Now()
	claims := Claims{
		ID:       id,
		Type:     "access",
		Identity: identity,
		RoleID:   role,
		RoleCode: roleCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// GenerateRefreshToken 生成 RefreshToken
func GenerateRefreshToken(id uint64, identity string, role uint64, roleCode string) (string, error) {
	now := time.Now()
	claims := Claims{
		ID:       id,
		Type:     "refresh",
		Identity: identity,
		RoleID:   role,
		RoleCode: roleCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无法验证的签名方式")
			}
			return secret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 token")
	}
	return claims, nil
}

func AccessTTL() time.Duration {
	return accessTokenTTL
}

func RefreshTTL() time.Duration {
	return refreshTokenTTL
}

func AdminTokenKey(id uint64) string {
	return fmt.Sprintf("admin:token:%d", id)
}

func AdminRefreshKey(id uint64) string {
	return fmt.Sprintf("admin:refresh:%d", id)
}
