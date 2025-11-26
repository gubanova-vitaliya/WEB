# Настройка MinIO для отображения изображений

## Проблема с ошибкой 500

Ошибка `ERR_ABORTED 500` при запросе к `/api/gases` теперь обрабатывается корректно:
- При ошибке используются mock данные
- Изображения загружаются из локальных файлов (`/gas-images/*.svg`)
- Приложение продолжает работать без бэкенда

## Настройка MinIO

### 1. Запуск MinIO

```bash
# В корневой директории проекта
docker-compose up -d
```

MinIO будет доступен на:
- API: `http://localhost:9000`
- Console: `http://localhost:9001`

### 2. Создание bucket и настройка доступа

1. Откройте MinIO Console: `http://localhost:9001`
2. Войдите:
   - Access Key: `minio`
   - Secret Key: `minio124`
3. Создайте bucket `gases`:
   - Нажмите "Create Bucket"
   - Имя: `gases`
   - Нажмите "Create Bucket"
4. Настройте публичный доступ:
   - Выберите bucket `gases`
   - Перейдите в "Access Policy"
   - Установите "Public" или "Download" для чтения

### 3. Загрузка изображений

**Через веб-интерфейс:**
1. Откройте bucket `gases`
2. Нажмите "Upload"
3. Загрузите изображения газов
4. Скопируйте URL (например: `http://localhost:9000/gases/hydrogen-1234567890.jpg`)

**Через API:**
```bash
curl -X POST http://localhost:8080/api/gases/1/image \
  -F "file=@path/to/hydrogen.jpg"
```

### 4. Обновление URL в БД

После загрузки обновите `image_url` в базе данных:

```sql
UPDATE gas 
SET image_url = 'http://localhost:9000/gases/hydrogen-1234567890.jpg' 
WHERE id = 1;
```

### 5. Проксирование через бэкенд (опционально)

Изображения из MinIO теперь можно проксировать через бэкенд:
- URL в БД: `http://localhost:9000/gases/hydrogen-1234567890.jpg`
- Прокси URL: `http://localhost:3000/api/minio/gases/hydrogen-1234567890.jpg`

Это решает проблемы с CORS.

## Отображение изображений

### Без запущенного сервиса (mock данные)
- Изображения загружаются из `public/gas-images/*.svg`
- Пути указаны в `src/modules/mock.ts`

### С запущенным сервисом
- Изображения загружаются из MinIO: `http://localhost:9000/gases/{filename}`
- Или через прокси: `http://localhost:3000/api/minio/gases/{filename}`
- При ошибке загрузки используется дефолтное изображение

## Проверка работы

1. **Без бэкенда:**
   - Запустите только фронтенд: `npm run dev`
   - Откройте `http://localhost:3000/gases`
   - Изображения должны загружаться из локальных файлов

2. **С бэкендом:**
   - Запустите бэкенд: `go run cmd/GaseProject/main.go`
   - Запустите MinIO: `docker-compose up -d`
   - Загрузите изображения в MinIO
   - Обновите URL в БД
   - Изображения должны загружаться из MinIO

## Troubleshooting

### Изображения не загружаются из MinIO

1. Проверьте, что MinIO запущен: `docker ps`
2. Проверьте доступность: `http://localhost:9000`
3. Проверьте CORS настройки в MinIO
4. Проверьте, что bucket публичный
5. Используйте прокси через бэкенд: `/api/minio/gases/{filename}`

### Ошибка 500 при запросе к API

- Это нормально, если бэкенд не запущен
- Приложение автоматически использует mock данные
- Изображения загружаются из локальных файлов




