# Инструкция: Запросы в Postman

## Важно: Все заголовки указываются ВРУЧНУЮ (белые поля)

В Postman:
1. Откройте вкладку **Headers**
2. Удалите все автоматически добавленные заголовки (если есть)
3. Добавляйте заголовки вручную, нажимая **"Add Header"**

---

## 1. POST Регистрация пользователя

### Настройки запроса:
- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/register`

### Headers (вручную):
```
Content-Type: application/json
```

### Body (raw JSON):
```json
{
  "login": "user123",
  "password": "password123",
  "email": "user123@example.com",
  "name": "Иван Иванов"
}
```

### Ожидаемый ответ:
```json
{
  "message": "User registered successfully",
  "user": {
    "uuid": "b84af5ba-6bb0-4592-bd7d-55006c07a5ad",
    "name": "Иван Иванов",
    "login": "user123",
    "email": "user123@example.com",
    "role": "buyer"
  }
}
```

---

## 2. POST Аутентификация (логин)

### Настройки запроса:
- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/login`

### Headers (вручную):
```
Content-Type: application/json
```

### Body (raw JSON):
```json
{
  "login": "user123",
  "password": "password123"
}
```

### Ожидаемый ответ:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "uuid": "b84af5ba-6bb0-4592-bd7d-55006c07a5ad",
    "name": "Иван Иванов",
    "login": "user123",
    "email": "user123@example.com",
    "role": "buyer"
  }
}
```

**ВАЖНО:** Скопируйте значение `access_token` из ответа - оно понадобится для следующих запросов!

---

## 3. GET Получение данных пользователя (личный кабинет)

### Настройки запроса:
- **Method:** `GET`
- **URL:** `http://localhost:8080/api/users/me`

### Headers (вручную):
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc
Content-Type: application/json
```

**ВАЖНО:** 
- Замените `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` на ваш реальный токен из шага 2
- После слова `Bearer` должен быть **один пробел**, затем токен

### Body:
Не требуется (GET запрос)

### Ожидаемый ответ:
```json
{
  "uuid": "b84af5ba-6bb0-4592-bd7d-55006c07a5ad",
  "name": "Иван Иванов",
  "login": "user123",
  "email": "user123@example.com",
  "role": "buyer"
}
```

---

## 4. PUT Обновление данных пользователя (личный кабинет)

### Настройки запроса:
- **Method:** `PUT`
- **URL:** `http://localhost:8080/api/users/me`

### Headers (вручную):
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc
Content-Type: application/json
```

**ВАЖНО:** Замените токен на ваш реальный токен из шага 2

### Body (raw JSON):
```json
{
  "login": "newlogin123"
}
```

**Примечание:** Можно обновить только `login`. Другие поля пока не поддерживаются.

### Ожидаемый ответ:
- **Status Code:** `204 No Content` (без тела ответа)

---

## 5. POST Деавторизация (logout)

### Настройки запроса:
- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/logout`

### Headers (вручную):
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc
Content-Type: application/json
```

**ВАЖНО:** Замените токен на ваш реальный токен из шага 2

### Body:
Не требуется (POST запрос без тела)

### Ожидаемый ответ:
- **Status Code:** `204 No Content` (без тела ответа)

**После logout:**
- Сессия удаляется из Redis
- Токен добавляется в blacklist
- Токен больше нельзя использовать для аутентификации

---

## Примеры значений для копирования

### Пример токена (замените на ваш):
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc
```

### Пример заголовка Authorization (замените токен):
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjU1MzU0NDYsImlhdCI6MTc2NTQ0OTA0NiwidXNlcl91dWlkIjoiYjg0YWY1YmEtNmJiMC00NTkyLWJkN2QtNTUwMDZjMDdhNWFkIiwicm9sZSI6MH0.08_eXpb1y4E5iBgjrKKQc5sQT2U_moe3qdFtQKBb_Bc
```

---

## Порядок выполнения запросов:

1. **POST /api/auth/register** - создать пользователя
2. **POST /api/auth/login** - получить токен (скопировать `access_token`)
3. **GET /api/users/me** - получить данные пользователя (использовать токен)
4. **PUT /api/users/me** - обновить данные (использовать токен)
5. **POST /api/auth/logout** - выйти из системы (использовать токен)

---

## Проверка работы Redis:

После логина и logout можно проверить Redis:

```bash
go run cmd/redis-viewer/main.go
```

Вы увидите:
- Активные сессии (после логина)
- Токены в blacklist (после logout)

---

## Частые ошибки:

### 401 Unauthorized
- Проверьте, что токен скопирован полностью (без пробелов в начале/конце)
- Убедитесь, что после `Bearer` есть пробел
- Проверьте, что токен не истек (TTL 24 часа)

### 400 Bad Request
- Проверьте формат JSON в Body
- Убедитесь, что все обязательные поля заполнены

### 404 Not Found
- Проверьте URL (должен быть `http://localhost:8080/api/...`)
- Убедитесь, что приложение запущено

---

## Настройка Postman для удобства:

1. **Создайте Environment:**
   - Variable: `base_url` = `http://localhost:8080`
   - Variable: `token` = (ваш токен после логина)

2. **Используйте переменные в URL:**
   - `{{base_url}}/api/auth/login`

3. **Используйте переменные в заголовках:**
   - `Authorization: Bearer {{token}}`

