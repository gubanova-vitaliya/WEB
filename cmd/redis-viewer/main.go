package main

import (
	"WEB/internal/app/config"
	"WEB/internal/app/redis"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Инициализируем Redis клиент
	ctx := context.Background()
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Fatalf("error connecting to Redis: %v", err)
	}
	defer redisClient.Close()

	fmt.Println("=== Redis Session Viewer ===")
	fmt.Println()

	// Получаем все сессии
	sessions, err := redisClient.GetAllSessions(ctx)
	if err != nil {
		logrus.Fatalf("error getting sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No active sessions found in Redis.")
		fmt.Println()
		fmt.Println("To create a session, login through the API:")
		fmt.Println("  POST /api/auth/login")
		return
	}

	fmt.Printf("Found %d active session(s):\n", len(sessions))
	fmt.Println(strings.Repeat("=", 80))

	// Выводим информацию о каждой сессии
	for sessionID, sessionData := range sessions {
		// Показываем первые 50 символов sessionID
		sessionIDDisplay := sessionID
		if len(sessionID) > 50 {
			sessionIDDisplay = sessionID[:50] + "..."
		}
		fmt.Printf("\nSession ID: %s\n", sessionIDDisplay)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("  User UUID:    %s\n", sessionData.UserUUID)
		fmt.Printf("  User Login:   %s\n", sessionData.UserLogin)
		fmt.Printf("  User Name:    %s\n", sessionData.UserName)
		fmt.Printf("  User Email:   %s\n", sessionData.UserEmail)
		fmt.Printf("  User Role:    %s\n", sessionData.UserRole)
		fmt.Printf("  Login Time:   %s\n", sessionData.LoginTime.Format(time.RFC3339))
		fmt.Printf("  Last Access:  %s\n", sessionData.LastAccess.Format(time.RFC3339))

		// Вычисляем TTL
		sessionKey := fmt.Sprintf("gase_service.session.%s", sessionID)
		ttl, err := redisClient.GetClient().TTL(ctx, sessionKey).Result()
		if err == nil && ttl > 0 {
			fmt.Printf("  TTL:          %s (expires in %s)\n", ttl.String(), time.Now().Add(ttl).Format(time.RFC3339))
		}
		fmt.Println()
	}

	// Дополнительная информация
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\nAdditional Redis Information:")
	fmt.Println(strings.Repeat("-", 80))

	// Получаем все ключи с префиксом сервиса
	allKeys, err := redisClient.GetClient().Keys(ctx, "gase_service.*").Result()
	if err == nil {
		fmt.Printf("Total keys with service prefix: %d\n", len(allKeys))

		// Группируем по типам
		jwtKeys := 0
		sessionKeys := 0
		for _, key := range allKeys {
			if strings.Contains(key, "jwt.") {
				jwtKeys++
			} else if strings.Contains(key, "session.") {
				sessionKeys++
			}
		}
		fmt.Printf("  - JWT blacklist entries: %d\n", jwtKeys)
		fmt.Printf("  - Active sessions: %d\n", sessionKeys)
	}

	// Показываем JWT blacklist
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("JWT Blacklist (Revoked Tokens):")
	fmt.Println(strings.Repeat("=", 80))

	blacklistedTokens, err := redisClient.GetAllBlacklistedTokens(ctx)
	if err != nil {
		logrus.Warnf("Failed to get blacklisted tokens: %v", err)
	} else if len(blacklistedTokens) == 0 {
		fmt.Println("No tokens in blacklist.")
		fmt.Println("\nBlacklist explanation:")
		fmt.Println("  JWT blacklist contains tokens that have been revoked (e.g., after logout).")
		fmt.Println("  These tokens cannot be used for authentication even if they haven't expired yet.")
		fmt.Println("  Tokens are added to blacklist when user logs out via POST /api/auth/logout")
	} else {
		fmt.Printf("Found %d blacklisted token(s):\n", len(blacklistedTokens))
		fmt.Println()

		for i, token := range blacklistedTokens {
			tokenDisplay := token
			if len(token) > 50 {
				tokenDisplay = token[:50] + "..."
			}

			// Получаем TTL токена
			ttl, err := redisClient.GetBlacklistTokenInfo(ctx, token)
			if err == nil && ttl > 0 {
				fmt.Printf("[%d] Token: %s\n", i+1, tokenDisplay)
				fmt.Printf("     TTL: %s (expires in %s)\n", ttl.String(), time.Now().Add(ttl).Format(time.RFC3339))
			} else {
				fmt.Printf("[%d] Token: %s\n", i+1, tokenDisplay)
				fmt.Printf("     Status: Expired or invalid\n")
			}
			fmt.Println()
		}

		fmt.Println("Note: These tokens are revoked and cannot be used for authentication.")
	}

	// Показываем информацию о пользователях
	fmt.Println("\nSessions grouped by user:")
	fmt.Println(strings.Repeat("-", 80))
	userSessions := make(map[string][]string)
	for sessionID, sessionData := range sessions {
		userSessions[sessionData.UserLogin] = append(userSessions[sessionData.UserLogin], sessionID)
	}

	for userLogin, sessionIDs := range userSessions {
		fmt.Printf("\nUser: %s (%d session(s))\n", userLogin, len(sessionIDs))
		for i, sessionID := range sessionIDs {
			sessionIDDisplay := sessionID
			if len(sessionID) > 30 {
				sessionIDDisplay = sessionID[:30] + "..."
			}
			fmt.Printf("  [%d] %s\n", i+1, sessionIDDisplay)
		}
	}

	// JSON вывод (опционально)
	if len(os.Args) > 1 && os.Args[1] == "--json" {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("JSON Output:")
		fmt.Println(strings.Repeat("=", 80))
		jsonData, err := json.MarshalIndent(sessions, "", "  ")
		if err == nil {
			fmt.Println(string(jsonData))
		}
	}
}
