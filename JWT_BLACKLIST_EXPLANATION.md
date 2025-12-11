# JWT Blacklist - Объяснение

## Что такое JWT Blacklist?

**JWT Blacklist** (черный список токенов) - это список JWT токенов, которые были **отозваны** (revoked) и больше не могут использоваться для аутентификации, даже если они еще не истекли по времени.

## Зачем нужен Blacklist?

### Проблема без Blacklist:
1. Пользователь логинится и получает JWT токен (срок действия 24 часа)
2. Пользователь выходит из системы (logout)
3. **НО токен все еще валиден до истечения 24 часов!**
4. Если кто-то украдет этот токен, он может использовать его для доступа к API

### Решение с Blacklist:
1. При logout токен добавляется в blacklist
2. При каждой аутентификации проверяется, не в blacklist ли токен
3. Токены из blacklist отклоняются, даже если они еще не истекли

## Как работает в нашем проекте?

### 1. При логине (`POST /api/auth/login`):
- Создается JWT токен
- Токен сохраняется как **сессия** в Redis
- Токен **НЕ** попадает в blacklist

### 2. При логауте (`POST /api/auth/logout`):
- Сессия удаляется из Redis
- **Токен добавляется в blacklist** (с TTL = 24 часа)
- Теперь этот токен нельзя использовать для аутентификации

### 3. При проверке токена:
- Проверяется, не истек ли токен (по времени)
- Проверяется, не находится ли токен в blacklist
- Если токен в blacklist → доступ запрещен

## Где хранится Blacklist?

Blacklist хранится в Redis с ключами вида:
```
gase_service.jwt.<JWT_TOKEN>
```

Значение: `true` (просто флаг, что токен отозван)
TTL: 24 часа (время жизни токена)

## Как посмотреть Blacklist?

### Через утилиту redis-viewer:

```bash
go run cmd/redis-viewer/main.go
```

Вы увидите раздел:
```
================================================================================
JWT Blacklist (Revoked Tokens):
================================================================================
Found 2 blacklisted token(s):

[1] Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
     TTL: 23h45m12s (expires in 2025-12-12T14:00:00+03:00)

[2] Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
     TTL: 12h30m5s (expires in 2025-12-12T02:30:00+03:00)

Note: These tokens are revoked and cannot be used for authentication.
```

### Через Redis CLI:

```bash
# Подключиться к Redis
docker exec -it rip-redis-1 redis-cli -a password

# Посмотреть все токены в blacklist
KEYS gase_service.jwt.*

# Посмотреть конкретный токен
GET gase_service.jwt.<ваш_токен>

# Посмотреть TTL токена
TTL gase_service.jwt.<ваш_токен>
```

## Пример использования

### 1. Пользователь логинится:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login": "demo", "password": "demo123"}'
```

**Результат:**
- Токен создан
- Сессия сохранена в Redis
- Blacklist: пуст

### 2. Пользователь выходит:
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer <токен>"
```

**Результат:**
- Сессия удалена из Redis
- Токен добавлен в blacklist
- Blacklist: содержит 1 токен

### 3. Попытка использовать отозванный токен:
```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer <отозванный_токен>"
```

**Результат:**
- ❌ Доступ запрещен (токен в blacklist)

## Важные моменты

1. **TTL blacklist = TTL токена**: Токен хранится в blacklist столько же, сколько он был бы валиден. После истечения TTL токен автоматически удаляется из Redis.

2. **Blacklist не заменяет проверку времени**: Токен проверяется и по времени, и по blacklist. Если токен истек, он не будет принят, даже если его нет в blacklist.

3. **Производительность**: Проверка blacklist происходит очень быстро, так как Redis - это in-memory база данных.

4. **Масштабируемость**: Если у вас несколько серверов, все они используют один и тот же Redis, поэтому blacklist работает для всех серверов.

## Код реализации

### Добавление в blacklist (при logout):
```go
// internal/app/handler/handler.go
func (h *Handler) ApiLogout(ctx *gin.Context) {
    // ...
    // Добавляем токен в blacklist
    h.Redis.WriteJWTToBlacklist(context.Background(), token, 24*time.Hour)
    // ...
}
```

### Проверка blacklist (в middleware):
```go
// internal/pkg/middleware.go
func (a *Application) SimpleAuthMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // ...
        // Проверяем, не в blacklist ли токен
        if err := a.RedisClient.CheckJWTInBlacklist(ctx, tokenString); err == nil {
            // Токен в blacklist - доступ запрещен
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
            ctx.Abort()
            return
        }
        // ...
    }
}
```

## Статистика

В утилите `redis-viewer` вы видите:
```
Additional Redis Information:
--------------------------------------------------------------------------------
Total keys with service prefix: 5
  - JWT blacklist entries: 2
  - Active sessions: 3
```

Это означает:
- **2 токена** в blacklist (отозваны)
- **3 активные сессии** (пользователи залогинены)

