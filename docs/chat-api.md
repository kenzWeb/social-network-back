# Chat API (WebSocket + REST)

Версия: v1  
Базовый путь: `/api/v1`

Аутентификация: все WS/REST эндпоинты требуют Bearer JWT (как и остальной бэкенд).

## WebSocket

Endpoint: `GET /api/v1/ws`

Заголовки:

- `Authorization: Bearer <access_token>`

Сообщения передаются в формате JSON: `{ "type": string, "data": any }`

Поддерживаемые события (client -> server):

- `message`
  - data: `{ "conversation_id": "<id>", "body": "<text>" }`
- `typing`
  - data: `{ "conversation_id": "<id>", "is_typing": true|false }`
- `read`
  - data: `{ "conversation_id": "<id>" }`

События от сервера (server -> client):

- `presence` — изменение статуса пользователя
  - `{ "type": "presence", "data": { "user_id": "<id>", "online": true|false, "last_seen": 1695040000 } }`
- `typing` — индикатор набора собеседника
  - `{ "type": "typing", "data": { "conversation_id": "<id>", "user_id": "<id>", "is_typing": true|false } }`
- `message` — новое сообщение в беседе
  - `{ "type": "message", "data": { "conversation_id": "<id>", "sender_id": "<id>", "body": "<text>", "created_at": 1695040000 } }`
- `error` — ошибка обработки
  - `{ "type": "error", "data": { "error": "<string>" } }`

Примечания:

- Presence: при подключении приходит `online=true`, при отключении отправляется `online=false` + `last_seen`.
- `typing` рассылается всем участникам беседы (кроме отправителя, если front это не требуется).

## REST

Все запросы требуют заголовок `Authorization: Bearer <access_token>`.

### Список диалогов

GET `/api/v1/chat/conversations`

Query:

- `limit` (по умолчанию 20)
- `offset` (по умолчанию 0)

Ответ:

```
{
  "conversations": [
    {"id": "c1", "created_at": "...", "updated_at": "..."}
  ]
}
```

### История сообщений

GET `/api/v1/chat/conversations/:id/messages`

Query:

- `limit` (по умолчанию 50)
- `offset` (по умолчанию 0)

Ответ (в хронологическом порядке):

```
{
  "messages": [
    {"id": "m1", "conversation_id": "c1", "sender_id": "u1", "body": "hi", "created_at": "..."}
  ]
}
```

### Отправка сообщения в 1:1

POST `/api/v1/chat/direct/:user_id/send`

Body:

```
{
  "body": "Привет!"
}
```

Ответ:

```
{
  "message": {"id": "m1", "conversation_id": "c1", "sender_id": "me", "body": "Привет!", "created_at": "..."}
}
```

Сайд-эффект: событие `message` по WS уходит обоим участникам.

### Пометить беседу прочитанной

POST `/api/v1/chat/conversations/:id/read`

Ответ: `204 No Content`

### Presence (онлайн/последний визит)

GET `/api/v1/chat/presence/:user_id`

Ответ:

```
{
  "user_id": "u2",
  "online": true,
  "last_seen": 1695040000
}
```

## Примеры

### Подключение WS (browser)

```js
const socket = new WebSocket('wss://example.com/api/v1/ws', [])
socket.onopen = () => console.log('ws open')
socket.onmessage = (e) => console.log('event:', e.data)
// Важно: передать JWT в заголовках при апгрейде (в браузере нельзя) — используйте fetch+upgrade с прокси на бэке или подключайтесь из нативного клиента.
```

В браузере нельзя установить произвольные заголовки для WebSocket. Обычно делают:

- подключение через бекенд-прокси, который подставляет заголовок;
- или JWT помещают в query `?token=...` и валидируют на сервере.

Текущая реализация ждёт Header `Authorization`. Если нужен query-параметр — могу добавить.

### Отправка события message

```json
{
	"type": "message",
	"data": {"conversation_id": "c1", "body": "hello"}
}
```

### Отправка typing

```json
{
	"type": "typing",
	"data": {"conversation_id": "c1", "is_typing": true}
}
```

## Ошибки

- 401 Unauthorized — отсутствует или неверный JWT
- 400 Bad Request — некорректный payload
- 500 Internal Server Error — внутренняя ошибка

## Заметки по масштабированию

- Для горизонтального масштабирования WS-хаба потребуется шарить presence/события через Redis Pub/Sub.
- Для истории сообщений — использовать пагинацию по курсору и индексы уже добавлены.
