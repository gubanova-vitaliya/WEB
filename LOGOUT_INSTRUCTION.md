# Инструкция: Правильный logout и добавление токена в blacklist

## Проблема
При logout токен не добавлялся в blacklist, потому что токен не передавался в заголовке Authorization.

## Решение
Теперь токен сохраняется в контекст при прохождении через middleware и используется при logout.

## Как правильно выполнить logout через Swagger:

### Вариант 1: Использование кнопки "Authorize" (рекомендуется)

1. **После логина**, в Swagger UI нажмите кнопку **"Authorize"** (🔒) в верхней части страницы
2. В поле **Value** вставьте ваш токен (без слова "Bearer", только сам токен)
3. Нажмите **"Authorize"**, затем **"Close"**
4. Теперь все защищенные эндпоинты будут автоматически использовать этот токен
5. Найдите эндпоинт **POST `/api/auth/logout`**
6. Нажмите **"Try it out"**
7. Нажмите **"Execute"** (токен уже будет в заголовках)
8. Проверьте логи приложения - должны увидеть:
   ```
   === ApiLogout called ===
   Token extracted from Authorization header...
   ✓ Session deleted successfully
   ✓ Token added to blacklist successfully
   ```

### Вариант 2: Ручное добавление токена в заголовок

1. Найдите эндпоинт **POST `/api/auth/logout`**
2. Нажмите **"Try it out"**
3. В разделе **Parameters** найдите поле **Authorization**
4. Введите: `Bearer <ваш_токен>` (с пробелом после Bearer)
5. Нажмите **"Execute"**

### Вариант 3: Через curl

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer <ваш_токен>"
```

## Проверка результата

После logout проверьте blacklist:

```bash
go run cmd/redis-viewer/main.go
```

Вы должны увидеть токен в разделе:
```
================================================================================
JWT Blacklist (Revoked Tokens):
================================================================================
Found 1 blacklisted token(s):

[1] Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
     TTL: 23h59m45s (expires in ...)
```

## Что происходит при logout:

1. ✅ Сессия удаляется из Redis
2. ✅ Токен добавляется в blacklist (TTL = 24 часа)
3. ✅ Токен больше нельзя использовать для аутентификации

## Логи приложения

При успешном logout вы увидите в логах:
```
=== ApiLogout called ===
Authorization header present: true
Redis client is nil: false
Token extracted from Authorization header (first 30 chars): eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Attempting to delete session from Redis...
✓ Session deleted successfully
Attempting to add token to blacklist...
✓ Token added to blacklist successfully
✓ Verified: Token is now in blacklist
=== ApiLogout completed ===
```

## Если токен не добавляется в blacklist:

1. **Проверьте логи приложения** - там будет указана причина
2. **Убедитесь, что Redis запущен**:
   ```bash
   docker ps | grep redis
   ```
3. **Проверьте, что приложение подключено к Redis**:
   - В логах при запуске должно быть: `Redis client initialized successfully`
4. **Убедитесь, что токен передается в заголовке Authorization**

## Важно:

- Токен должен быть передан в формате: `Bearer <токен>`
- После logout токен нельзя использовать повторно
- Blacklist автоматически очищается через 24 часа (TTL токена)

