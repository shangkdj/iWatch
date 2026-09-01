package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWT(t *testing.T) {
	userID := int64(12345)
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("生成 JWT 失败: %v", err)
	}
	if token == "" {
		t.Error("生成的 token 为空")
	}
	t.Logf("生成的 Token: %s", token)
}

func TestValidateJWT(t *testing.T) {
	userID := int64(12345)
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("生成 JWT 失败: %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("验证 JWT 失败: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("期望 UserID=%d, 实际得到 %d", userID, claims.UserID)
	}
	if claims.Issuer != "watch-api" {
		t.Errorf("期望 Issuer=watch-api, 实际得到 %s", claims.Issuer)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	// 测试无效 token
	invalidTokens := []string{
		"invalid-token",
		"",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
	}

	for _, token := range invalidTokens {
		_, err := ValidateJWT(token)
		if err == nil {
			t.Errorf("无效 token '%s' 应该返回错误", token)
		}
	}
}

func TestExpiredJWT(t *testing.T) {
	// 创建一个已经过期的 JWT
	claims := Claims{
		UserID: 12345,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1小时前过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "watch-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("生成过期 token 失败: %v", err)
	}

	_, err = ValidateJWT(tokenString)
	if err == nil {
		t.Error("过期的 token 应该返回错误")
	}
	t.Logf("过期 token 错误: %v", err)
}

func TestRefreshJWT(t *testing.T) {
	userID := int64(12345)
	oldToken, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("生成 JWT 失败: %v", err)
	}

	newToken, err := RefreshJWT(oldToken)
	if err != nil {
		t.Fatalf("刷新 JWT 失败: %v", err)
	}
	if newToken == oldToken {
		t.Error("刷新后的 token 应该不同")
	}

	// 验证新 token 是否有效
	claims, err := ValidateJWT(newToken)
	if err != nil {
		t.Fatalf("验证新 token 失败: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("期望 UserID=%d, 实际得到 %d", userID, claims.UserID)
	}
}

func TestJWTSecretLength(t *testing.T) {
	if len(jwtSecret) < 32 {
		t.Errorf("JWT_SECRET 长度不足，当前 %d 字节，需要 >= 32", len(jwtSecret))
	}
	t.Logf("✅ JWT_SECRET 长度: %d 字节", len(jwtSecret))
}
