# 🏦 Bank Queue System

Система электронной очереди для банковских отделений с real-time обновлениями через WebSocket и ТВ-табло.

## 📋 Содержание
- [Архитектура](#архитектура)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Установка и запуск](#установка-и-запуск)
- [API Документация](#api-документация)
- [WebSocket Real-time обновления](#websocket-real-time-обновления)
- [ТВ-табло](#тв-табло)
- [Тестирование](#тестирование)
- [Интеграция с другими системами](#интеграция-с-другими-системами)

## 🏗 Архитектура

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Flutter App   │     │   ТВ-табло       │     │   Внешние       │
│   (Клиенты)     │     │   (Монитор)      │     │   Системы       │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         └───────────────┬───────┴───────────────┬───────┘
                        ▼                       ▼
              ┌─────────────────────────────────────┐
              │         REST API / WebSocket         │
              │         (порт 8080)                   │
              └─────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              │         Go Backend              │
              │    (Gorilla Mux + WebSocket)    │
              └───────────────┬───────────────┘
                              │
              ┌───────────────┴───────────────┐
              │         PostgreSQL              │
              │         (Банк данных)           │
              └─────────────────────────────────┘
```

## 🛠 Технологии

- **Backend**: Go 1.21+
- **База данных**: PostgreSQL 15
- **API**: REST + WebSocket (gorilla/websocket)
- **Контейнеризация**: Docker + Docker Compose
- **Фронтенд (ТВ-табло)**: HTML, CSS, JavaScript

## 📁 Структура проекта

```
bank_go/
├── cmd/
│   └── main.go                 # Точка входа
├── internal/
│   ├── geoGet/                  # Геокодинг (OpenStreetMap)
│   │   ├── geocoder/
│   │   └── overpass/
│   ├── model/                    # Модели данных
│   │   ├── bank/
│   │   │   ├── branch.go
│   │   │   └── ticket.go
│   │   └── queue/
│   ├── queue/                     # Логика очередей
│   │   └── manager.go
│   ├── router/                     # Маршрутизация
│   │   └── router.go
│   ├── database/                    # Работа с БД
│   │   └── postgres/
│   │       ├── client.go
│   │       ├── queue_repo.go
│   │       ├── branch_repo.go
│   │       └── schema.sql
│   ├── ticket-controller/          # HTTP хендлеры
│   │   └── handlers.go
│   └── websocketStart/              # WebSocket менеджер
│       └── manager.go
├── static/                          # Статические файлы
│   └── tv_display.html              # ТВ-табло
├── docker-compose.yml               # Docker Compose
├── Dockerfile                       # Docker образ
├── .env                             # Переменные окружения
├── go.mod                           # Зависимости Go
└── README.md                        # Этот файл
```

## 🚀 Установка и запуск

### Предварительные требования
- Go 1.21+
- Docker и Docker Compose
- Git

### 1. Клонирование репозитория
```bash
git clone https://github.com/DrollltedUp/bank_go.git
cd bank_go
```

### 2. Настройка окружения
Создайте файл `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=bankuser
DB_PASSWORD=bankpassword123
DB_NAME=bank_queue
DB_SSLMODE=disable
SERVER_PORT=8080
```

### 3. Запуск PostgreSQL через Docker
```bash
docker-compose up -d postgres
```

### 4. Установка зависимостей Go
```bash
go mod download
```

### 5. Запуск сервера
```bash
go run cmd/main.go
```

### 6. Проверка работы
```bash
# Получить список отделений Сбербанка в Москве
curl -X POST http://localhost:8080/api/bank/branches \
  -H "Content-Type: application/json" \
  -d '{"bank": "Сбербанк", "city": "Москва"}'
```

## 📚 API Документация

### Банковские отделения

#### Получить список отделений
```http
POST /api/bank/branches
Content-Type: application/json

{
    "bank": "Сбербанк",
    "city": "Москва"
}
```

**Ответ:**
```json
[
  {
    "bank_name": "Сбербанк",
    "address": "ул. Арбат, 1, Москва",
    "lat": 55.751244,
    "lng": 37.618423,
    "type": "branch",
    "load_score": 2,
    "load_color": "#8BC34A",
    "load_label": "Нормально",
    "tickets": 5,
    "windows": 2,
    "wait_time": 12
  }
]
```

### Талоны и очередь

#### Создать талон
```http
POST /tickets/{branch_id}
Content-Type: application/json

{
    "service_code": "CASH"
}
```

**Параметры:**
- `branch_id` - ID отделения (например, `Сбербанк-55.7512-37.6184`)
- `service_code` - код услуги: `CASH`, `PENSION`, `DEBIT`, `CREDIT`, `VIP`, `BUSINESS`

**Ответ:**
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "ticket_number": "005",
    "service_code": "CASH",
    "service_name": "Кассовое обслуживание",
    "branch_id": "Сбербанк-55.7512-37.6184",
    "branch_name": "Сбербанк",
    "position": 5,
    "wait_time": 12,
    "created_at": "15:30:45",
    "status": "waiting"
}
```

#### Получить статус очереди
```http
GET /tickets/{branch_id}/status
```

**Ответ:**
```json
{
    "branch_id": "Сбербанк-55.7512-37.6184",
    "tickets": 5,
    "windows": 2,
    "wait_time": 12,
    "load_score": 2,
    "load_color": "#8BC34A",
    "load_label": "Нормально",
    "distribution": {
        "CASH": 3,
        "PENSION": 1,
        "DEBIT": 1
    }
}
```

#### Вызвать следующего клиента
```http
POST /tickets/{branch_id}/call
```

**Ответ:**
```json
{
    "success": true,
    "ticket": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "ticket_number": "005",
        "service_code": "CASH",
        "service_name": "Кассовое обслуживание",
        "status": "called"
    }
}
```

### Услуги

#### Получить список всех услуг
```http
GET /api/services
```

**Ответ:**
```json
[
    {"code": "CASH", "name": "Кассовое обслуживание", "color": "#FF6B6B"},
    {"code": "PENSION", "name": "Пенсии и пособия", "color": "#4ECDC4"},
    {"code": "DEBIT", "name": "Дебетовые карты", "color": "#45B7D1"},
    {"code": "CREDIT", "name": "Кредитные карты", "color": "#96CEB4"},
    {"code": "VIP", "name": "Премиум-обслуживание", "color": "#DDA0DD"},
    {"code": "BUSINESS", "name": "Юридическим лицам", "color": "#98D8C8"}
]
```

## 🔌 WebSocket Real-time обновления

### Подключение
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

### Сообщения от сервера

**Новый талон:**
```json
{
    "type": "ticket_created",
    "branch_id": "Сбербанк-55.7512-37.6184",
    "ticket": "005",
    "service": "Кассовое обслуживание",
    "position": 5,
    "wait_time": 12,
    "timestamp": "15:30:45",
    "action": "new_ticket"
}
```

**Вызов талона:**
```json
{
    "type": "ticket_called",
    "branch_id": "Сбербанк-55.7512-37.6184",
    "ticket": "005",
    "service": "Кассовое обслуживание",
    "timestamp": "15:35:22",
    "action": "call_ticket"
}
```

## 📺 ТВ-табло

Встроенное ТВ-табло для отображения очереди в реальном времени.

### Доступ
```
http://localhost:8080/
```

### Функциональность
- Отображение нескольких отделений одновременно
- Цветовая индикация загруженности (зеленый → красный)
- Анимация при создании новых талонов
- Автоматическое обновление через WebSocket
- Адаптивный дизайн

## 🧪 Тестирование

### Скрипт для массового создания талонов
```bash
#!/bin/bash
for i in {1..10}; do
    curl -X POST http://localhost:8080/tickets/Сбербанк-55.7512-37.6184 \
        -H "Content-Type: application/json" \
        -d '{"service_code": "CASH"}'
    echo ""
    sleep 1
done
```

### Проверка WebSocket
```bash
# Установи wscat
npm install -g wscat

# Подключись к WebSocket
wscat -c ws://localhost:8080/ws
```

## 🔗 Интеграция с другими системами

### REST API (любой язык)
```python
import requests

# Создание талона
response = requests.post(
    'http://localhost:8080/tickets/Сбербанк-55.7512-37.6184',
    json={'service_code': 'CASH'}
)
ticket = response.json()
print(f"Талон: {ticket['ticket_number']}")
```

### WebSocket (JavaScript)
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'ticket_created') {
        console.log(`Новый талон ${data.ticket} для услуги ${data.service}`);
        // Обновить UI
    }
};
```

## 🐳 Docker

### Запуск всех сервисов
```bash
docker-compose up -d
```

### Просмотр логов
```bash
docker-compose logs -f
```

### Остановка
```bash
docker-compose down
```

### Полная перезагрузка (с удалением данных)
```bash
docker-compose down -v
docker-compose up -d
```

## 📊 Оценка загруженности

| Оценка | Цвет | Название | Описание |
|--------|------|----------|----------|
| 1 | 🟢 Зеленый | Свободно | < 2 чел/окно |
| 2 | 🟢 Светло-зеленый | Нормально | 2-4 чел/окно |
| 3 | 🟡 Желтый | Загружено | 4-7 чел/окно |
| 4 | 🟠 Оранжевый | Многолюдно | 7-10 чел/окно |
| 5 | 🔴 Красный | Переполнено | > 10 чел/окно |

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

MIT License

## 👥 Авторы

- [DrollltedUp](https://github.com/DrollltedUp)

## 🙏 Благодарности

- Gorilla WebSocket за отличную библиотеку
- OpenStreetMap за геоданные
- PostgreSQL за надежную базу данных

---

**⭐ Если проект полезен, поставьте звезду на GitHub!**
