package redis

import (
	"WEB/internal/app/config"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "gase_service." // наш префикс сервиса

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := &Client{}
	client.cfg = cfg

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfg.Password,
		Username:    cfg.User,
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient возвращает внутренний redis клиент для прямого доступа
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// JWT методы
const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.client.Set(ctx, getJWTKey(jwtStr), true, jwtTTL).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Get(ctx, getJWTKey(jwtStr)).Err()
	// если токена нет, то вернется ошибка not exists
}

// GetAllBlacklistedTokens получает все токены из blacklist
func (c *Client) GetAllBlacklistedTokens(ctx context.Context) ([]string, error) {
	pattern := getJWTKey("*")
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get blacklist keys: %w", err)
	}

	// Извлекаем токены из ключей (убираем префикс)
	tokens := make([]string, 0, len(keys))
	for _, key := range keys {
		// key format: "gase_service.jwt.<token>"
		// извлекаем токен
		if strings.HasPrefix(key, servicePrefix+jwtPrefix) {
			token := key[len(servicePrefix+jwtPrefix):]
			tokens = append(tokens, token)
		}
	}

	return tokens, nil
}

// GetBlacklistTokenInfo получает информацию о токене в blacklist (TTL)
func (c *Client) GetBlacklistTokenInfo(ctx context.Context, token string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, getJWTKey(token)).Result()
	if err != nil {
		return 0, err
	}
	return ttl, nil
}

// Session методы для работы с сессиями пользователей
const sessionPrefix = "session."

// SessionData структура данных сессии
type SessionData struct {
	UserUUID   string    `json:"user_uuid"`
	UserLogin  string    `json:"user_login"`
	UserName   string    `json:"user_name"`
	UserRole   string    `json:"user_role"`
	UserEmail  string    `json:"user_email"`
	LoginTime  time.Time `json:"login_time"`
	LastAccess time.Time `json:"last_access"`
}

func getSessionKey(sessionID string) string {
	return servicePrefix + sessionPrefix + sessionID
}

// SaveSession сохраняет сессию пользователя в Redis
func (c *Client) SaveSession(ctx context.Context, sessionID string, sessionData SessionData, ttl time.Duration) error {
	sessionData.LastAccess = time.Now()
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	return c.client.Set(ctx, getSessionKey(sessionID), data, ttl).Err()
}

// GetSession получает данные сессии из Redis
func (c *Client) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	data, err := c.client.Get(ctx, getSessionKey(sessionID)).Result()
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Обновляем время последнего доступа
	sessionData.LastAccess = time.Now()
	ttl, _ := c.client.TTL(ctx, getSessionKey(sessionID)).Result()
	if ttl > 0 {
		c.SaveSession(ctx, sessionID, sessionData, ttl)
	}

	return &sessionData, nil
}

// DeleteSession удаляет сессию из Redis
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	return c.client.Del(ctx, getSessionKey(sessionID)).Err()
}

// GetAllSessions получает все сессии пользователей
func (c *Client) GetAllSessions(ctx context.Context) (map[string]*SessionData, error) {
	pattern := getSessionKey("*")
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	sessions := make(map[string]*SessionData)
	for _, key := range keys {
		// Извлекаем sessionID из ключа
		sessionID := key[len(servicePrefix+sessionPrefix):]
		data, err := c.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var sessionData SessionData
		if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
			continue
		}

		sessions[sessionID] = &sessionData
	}

	return sessions, nil
}

// GetSessionsByUserUUID получает все сессии конкретного пользователя
func (c *Client) GetSessionsByUserUUID(ctx context.Context, userUUID string) (map[string]*SessionData, error) {
	allSessions, err := c.GetAllSessions(ctx)
	if err != nil {
		return nil, err
	}

	userSessions := make(map[string]*SessionData)
	for sessionID, sessionData := range allSessions {
		if sessionData.UserUUID == userUUID {
			userSessions[sessionID] = sessionData
		}
	}

	return userSessions, nil
}
