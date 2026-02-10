package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OptionalAuth 可选 JWT 认证中间件（不强制要求登录，但会尝试解析 token）
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// 安全地提取用户信息（带类型检查）
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			// 类型不匹配，可选认证失败时继续但不设置用户信息
			c.Next()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			// 类型不匹配，可选认证失败时继续但不设置用户信息
			c.Next()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", int64(userIDFloat))
		c.Set("username", username)

		c.Next()
	}
}

// Auth JWT 认证中间件
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// 安全地提取用户信息（带类型检查）
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			// 必需认证：类型不匹配时返回错误
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			// 必需认证：类型不匹配时返回错误
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username in token"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", int64(userIDFloat))
		c.Set("username", username)

		c.Next()
	}
}
