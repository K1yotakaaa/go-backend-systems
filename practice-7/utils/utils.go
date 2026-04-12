package utils

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "secret"))

var RDB = redis.NewClient(&redis.Options{
	Addr: getEnv("REDIS_ADDR", "localhost:6379"),
})

var Ctx = context.Background()

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

func GenerateTokens(id uuid.UUID, role string) (string, string, error) {
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id.String(),
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id.String(),
		"role":    role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	a, err := access.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	r, err := refresh.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return a, r, nil
}

func GenerateAccessTokenFromRefresh(refreshToken string) (string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id")
	}

	role, ok := claims["role"].(string)
	if !ok {
		role = "user"
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user_id format")
	}

	_, err = RDB.Get(Ctx, "refresh:"+userIDStr).Result()
	if err != nil {
		return "", fmt.Errorf("refresh token not found")
	}

	access, _, err := GenerateTokens(userID, role)
	return access, err
}

func GenerateCode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func SendEmail(to, code string) error {
	fmt.Printf("\n========================================\n")
	fmt.Printf("📧 Email to: %s\n", to)
	fmt.Printf("🔐 Verification code: %s\n", code)
	fmt.Printf("========================================\n\n")
	return nil
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		if tokenStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token required"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid claims"})
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid user_id"})
			return
		}

		val, err := RDB.Get(Ctx, "auth:"+userID).Result()
		if err != nil || val != tokenStr {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token revoked or expired"})
			return
		}

		c.Set("userID", userID)
		c.Set("role", claims["role"])

		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		if role != requiredRole {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: admin access required"})
			return
		}
		c.Next()
	}
}

type rateLimiter struct {
	visitors map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

var limiter = &rateLimiter{
	visitors: make(map[string][]time.Time),
	limit:    10,
	window:   time.Minute,
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			key = userID.(string)
		}

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		now := time.Now()
		cutoff := now.Add(-limiter.window)

		requests := limiter.visitors[key]
		valid := make([]time.Time, 0)
		for _, t := range requests {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}

		if len(valid) >= limiter.limit {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		limiter.visitors[key] = append(valid, now)
		c.Next()
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}