# Инструкция: Аутентификация через Swagger и просмотр Redis

## Шаг 1: Запуск Redis

Убедитесь, что Redis запущен. Если используете Docker Compose:

```bash
docker-compose up -d redis
```

Или запустите Redis отдельно с паролем `password` (как указано в `config/config.toml`).

## Шаг 2: Запуск приложения

Запустите Go приложение:

```bash
go run cmd/GaseProject/main.go
```

Приложение запустится на `http://localhost:8080`

Вы должны увидеть в логах:
```
Redis client initialized successfully
Server start up
```

## Шаг 3: Открытие Swagger UI

Откройте в браузере:

```
http://localhost:8080/swagger/index.html
```

Вы увидите интерфейс Swagger с документацией API.

## Шаг 4: Аутентификация через Swagger

### Вариант A: Использовать `/api/auth/login` (рекомендуется - сохраняет сессию в Redis)

#### 4.1. Найдите эндпоинт `/api/auth/login`

1. В Swagger UI найдите раздел **Users**
2. Найдите эндпоинт **POST `/api/auth/login`**
3. Нажмите на него, чтобы развернуть

#### 4.2. Создайте пользователя (если еще нет)

**Способ 1: Через Swagger регистрацию**
1. Найдите эндпоинт **POST `/api/auth/register`** в разделе **Users**
2. Нажмите **"Try it out"**
3. Введите данные:
```json
{
  "login": "demo",
  "password": "demo123",
  "email": "demo@example.com",
  "name": "Demo User"
}
```
4. Нажмите **"Execute"**
5. Дождитесь ответа `201 Created`

**Способ 2: Использовать существующего пользователя**
- Если есть админ: `login: "admin"`, `password: "password"`

#### 4.3. Выполните запрос на логин

1. В эндпоинте **POST `/api/auth/login`** нажмите **"Try it out"**
2. В поле **Request body** введите:
```json
{
  "login": "demo",
  "password": "demo123"
}
```
3. Нажмите **"Execute"**

#### 4.4. Получите токен

В ответе вы увидите:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "uuid": "...",
    "name": "Demo User",
    "login": "demo",
    "email": "demo@example.com",
    "role": "buyer"
  }
}
```

**Скопируйте значение `access_token`** - это ваш JWT токен.

---

### Вариант B: Использовать `/login` (альтернативный эндпоинт)

Если `/api/auth/login` не работает, попробуйте альтернативный эндпоинт:

#### 4.1. Найдите эндпоинт `/login`

1. В Swagger UI найдите раздел **Auth**
2. Найдите эндпоинт **POST `/login`**
3. Нажмите на него, чтобы развернуть

#### 4.2. Создайте пользователя через `/sign_up`

1. Найдите эндпоинт **POST `/sign_up`** в разделе **Auth**
2. Нажмите **"Try it out"**
3. Введите данные:
```json
{
  "name": "Test User",
  "pass": "test123",
  "email": "test@example.com"
}
```
**Важно:** В этом эндпоинте используется поле `pass`, а не `password`!

4. Нажмите **"Execute"**
5. Дождитесь ответа `200 OK` с `{"ok": true}`

#### 4.3. Выполните запрос на логин

1. В эндпоинте **POST `/login`** нажмите **"Try it out"**
2. В поле **Request body** введите:
```json
{
  "login": "Test User",
  "password": "test123"
}
```
**Примечание:** В `/sign_up` логин создается из поля `name`, поэтому используйте имя пользователя как логин.

3. Нажмите **"Execute"**

#### 4.4. Получите токен

В ответе вы увидите:
```json
{
  "expires_in": 86400,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

**Скопируйте значение `access_token`** - это ваш JWT токен.

**Важно:** Этот эндпоинт `/login` НЕ сохраняет сессию в Redis. Для сохранения сессии используйте `/api/auth/login`.

### 4.5. Авторизация в Swagger (опционально)

Для использования защищенных эндпоинтов:

1. В верхней части Swagger UI нажмите кнопку **"Authorize"** (🔒)
2. В поле **Value** вставьте ваш токен (без слова "Bearer")
3. Нажмите **"Authorize"**, затем **"Close"**

Теперь вы можете использовать защищенные эндпоинты без ручного ввода токена.

## Шаг 5: Просмотр содержимого Redis

### 5.1. В новом терминале запустите утилиту просмотра Redis

```bash
go run cmd/redis-viewer/main.go
```

### 5.2. Что вы увидите

Вывод будет содержать:

```
=== Redis Session Viewer ===

Found 1 active session(s):
================================================================================

Session ID: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
--------------------------------------------------------------------------------
  User UUID:    <uuid-пользователя>
  User Login:   testuser
  User Name:    Test User
  User Email:   test@example.com
  User Role:    buyer
  Login Time:   2024-01-15T10:30:00Z
  Last Access:  2024-01-15T10:30:00Z
  TTL:          24h0m0s (expires in 2024-01-16T10:30:00Z)

================================================================================

Additional Redis Information:
--------------------------------------------------------------------------------
Total keys with service prefix: 1
  - JWT blacklist entries: 0
  - Active sessions: 1

Sessions grouped by user:
--------------------------------------------------------------------------------

User: testuser (1 session(s))
  [1] eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 5.3. JSON вывод (опционально)

Для получения данных в формате JSON:

```bash
go run cmd/redis-viewer/main.go --json
```

## Шаг 6: Проверка логаута

### 6.1. Выполните логаут через Swagger

1. Найдите эндпоинт **POST `/api/auth/logout`**
2. Нажмите **"Try it out"**
3. Если вы авторизованы (через кнопку Authorize), просто нажмите **"Execute"**
4. Если нет - добавьте заголовок:
   - В разделе **Parameters** найдите **Authorization**
   - Введите: `Bearer <ваш_токен>`
5. Нажмите **"Execute"**

### 6.2. Проверьте Redis снова

```bash
go run cmd/redis-viewer/main.go
```

Сессия должна быть удалена. Вы увидите:
```
No active sessions found in Redis.
```

## Дополнительные команды Redis

Если хотите проверить Redis напрямую через redis-cli:

```bash
# Подключение к Redis (если в Docker)
docker exec -it <redis-container-name> redis-cli -a password

# Или если Redis запущен локально
redis-cli -a password

# Просмотр всех ключей с префиксом сервиса
KEYS gase_service.*

# Просмотр конкретной сессии
GET gase_service.session.<ваш_токен>

# Просмотр TTL
TTL gase_service.session.<ваш_токен>
```

## Устранение проблем

### Redis не подключается

1. Проверьте, что Redis запущен:
   ```bash
   docker ps | grep redis
   ```

2. Проверьте пароль в `config/config.toml` (должен быть `password`)

3. Проверьте логи приложения - должно быть:
   ```
   Redis client initialized successfully
   ```

### Нет сессий в Redis после логина

1. Проверьте логи приложения при логине - должно быть:
   ```
   Session saved for user testuser (UUID: ...)
   ```

2. Убедитесь, что Redis клиент не nil (проверьте логи запуска)

### Swagger не открывается

1. Убедитесь, что приложение запущено на `localhost:8080`
2. Проверьте URL: `http://localhost:8080/swagger/index.html`
3. Убедитесь, что порт 8080 не занят другим приложением

## Пример полного цикла

```bash
# Терминал 1: Запуск приложения
go run cmd/GaseProject/main.go

# Терминал 2: Просмотр Redis (до логина)
go run cmd/redis-viewer/main.go
# Результат: No active sessions found

# Браузер: Swagger UI
# 1. Открыть http://localhost:8080/swagger/index.html
# 2. POST /api/auth/login с данными пользователя
# 3. Скопировать access_token

# Терминал 2: Просмотр Redis (после логина)
go run cmd/redis-viewer/main.go
# Результат: Показывает сессию пользователя

# Браузер: Swagger UI
# 4. POST /api/auth/logout (с авторизацией)

# Терминал 2: Просмотр Redis (после логаута)
go run cmd/redis-viewer/main.go
# Результат: No active sessions found
```

