package services

import (
	"errors"
	"time"

	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ 加载 .env 失败:", err)
	} else {
		log.Println("✅ .env 加载成功")
	}

	if secret := os.Getenv("JWT_SECRET"); secret == "" {
		log.Println("⚠️ JWT_SECRET 未设置，使用默认密钥（仅开发测试用）")
		secret = "dev-secret-min-32-characters-long!!"
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	log.Printf("✅ JWT_SECRET 长度: %d 字节\n", len(jwtSecret))
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Claims 自定义 JWT 声明
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT 令牌
// 参数 userID: 用户ID
// 返回: token字符串, error
func GenerateJWT(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     //签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                     //生效时间
			Issuer:    "watch-api",
			Subject:   "user-auth",
		},
	}

	// 创建 token
	// claims 声明载荷 JWT的核心数据部分（用户信息+过期时间等）
	// SigningMethodHS256 是 HMAC SHA256 签名方法 对称加密 速度快
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名并获取完整的编码后的字符串token
	return token.SignedString(jwtSecret)
}

// ValidateJWT 验证 JWT 令牌
// 参数 tokenString: 待验证的token
// 返回: Claims, error
func ValidateJWT(tokenString string) (*Claims, error) {
	// 1. 调用 ParseWithClaims 解析并验证
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		//2. 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		// 3. 返回密钥用于验证签名 返回给jwt库 重新计算签名并与token中的签名进行比较
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}
	// 4. 类型断言 + 有效性检查 ok只证明token一致 token.Valid检查token是否有效（过期、签名错误等）
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// RefreshJWT 刷新 JWT（延长过期时间）
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新 token，延长过期时间
	return GenerateJWT(claims.UserID)
}

// GetUserIDFromToken 从 token 中提取用户 ID（便捷方法）
func GetUserIDFromToken(tokenString string) (int64, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
