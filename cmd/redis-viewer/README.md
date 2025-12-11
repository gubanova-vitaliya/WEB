# Redis Session Viewer

Утилита для просмотра активных сессий пользователей в Redis.

## Использование

### Запуск утилиты

```bash
go run cmd/redis-viewer/main.go
```

Или скомпилировать и запустить:

```bash
go build -o redis-viewer cmd/redis-viewer/main.go
./redis-viewer
```

### JSON вывод

Для получения вывода в формате JSON:

```bash
go run cmd/redis-viewer/main.go --json
```

## Что показывает утилита

1. **Список всех активных сессий** с информацией:
   - Session ID (JWT токен)
   - User UUID
   - User Login
   - User Name
   - User Email
   - User Role
   - Login Time
   - Last Access
   - TTL (время до истечения)

2. **Дополнительная информация**:
   - Общее количество ключей в Redis с префиксом сервиса
   - Количество JWT токенов в blacklist
   - Количество активных сессий

3. **Группировка по пользователям**:
   - Все сессии, сгруппированные по логину пользователя

## Требования

- Redis должен быть запущен и доступен
- Конфигурация Redis должна быть указана в `config/config.toml`:
  ```toml
  redis_host = "localhost"
  redis_port = 6379
  redis_password = "password"
  ```

## Как создаются сессии

Сессии автоматически создаются при успешном логине через API:
```
POST /api/auth/login
```

Сессии удаляются при логауте:
```
POST /api/auth/logout
```

## Формат хранения в Redis

Сессии хранятся в Redis с ключами вида:
```
gase_service.session.<JWT_TOKEN>
```

Значение - JSON с данными сессии:
```json
{
  "user_uuid": "...",
  "user_login": "...",
  "user_name": "...",
  "user_role": "...",
  "user_email": "...",
  "login_time": "...",
  "last_access": "..."
}
```

