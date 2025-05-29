// TODO: 确保密钥通过配置安全管理，而不是硬编码
package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定义了 JWT 中携带的自定义数据以及标准的 RegisteredClaims。
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager interface {
	GenerateToken(userID uint, username string, role string) (string, error)
	VerifyToken(tokenString string) (*CustomClaims, error)
}

type jwtManagerImpl struct {
	secretKey       []byte
	issuer          string
	expirationHours int64
}

func NewJWTManagerImpl(secretKey string, issuer string, expirationHours int64) (JWTManager, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("NewJWTManagerImpl: jwt secretKey cannot be empty")
	}
	return &jwtManagerImpl{
		secretKey:       []byte(secretKey),
		issuer:          issuer,
		expirationHours: expirationHours,
	}, nil
}

// GenerateToken 为指定的用户信息生成一个新的 JWT。
func (j *jwtManagerImpl) GenerateToken(userID uint, username string, role string) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID), // 通常是用户的唯一标识
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(j.expirationHours))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("jwtManagerImpl.GenerateToken: failed to gennerater token. %w", err)
	}
	return signedToken, nil
}

// VerifyToken 验证给定的 JWT 字符串并返回 CustomClaims。
func (j *jwtManagerImpl) VerifyToken(tokenString string) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		// 错误处理，区分是过期、签名无效还是其他格式问题
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired: %w", err)
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token is malformed: %w", err)
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, fmt.Errorf("token signature is invalid: %w", err)
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token is invalid")
}
