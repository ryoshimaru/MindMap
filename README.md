# mindmap

Дипломный проект для декомпозиции пользовательских целей с помощью Gemini или DeepSeek.

## Возможности

- регистрация по почте и вход через Google OAuth 2.0;
- обязательный профиль контекста пользователя;
- проверка и зашифрованное хранение персонального AI API-ключа;
- последовательные уточняющие вопросы на русском языке;
- оценка реалистичности и предложение безопасной корректировки;
- двухуровневый план: этапы и шаги;
- дедлайны, зависимости, блокировка шагов и автоматический прогресс;
- редактирование, добавление, удаление и перестановка шагов;
- адаптивный тёмный интерфейс.

## Запуск через Docker

1. Создайте файл окружения:

   ```bash
   cp .env.example .env
   ```

2. Замените `AI_KEY_ENCRYPTION_KEY` в `.env` длинной случайной строкой.

3. Для Google OAuth создайте OAuth 2.0 Web Client в Google Cloud Console и добавьте redirect URI:

   ```text
   http://localhost:8080/api/auth/oauth/google/callback
   ```

   Затем заполните `GOOGLE_CLIENT_ID` и `GOOGLE_CLIENT_SECRET`. Обычный вход по почте работает и без них.

4. Запустите приложение:

   ```bash
   docker compose up --build
   ```

5. Откройте [http://localhost:3000](http://localhost:3000).

PostgreSQL доступен внутри Compose, API работает на `http://localhost:8080`. При первом входе интерфейс последовательно попросит AI-ключ и данные профиля.

## Локальная разработка

Backend:

```bash
export DB_URL='postgres://mindmap:mindmap@localhost:5432/mindmap?sslmode=disable'
export AI_KEY_ENCRYPTION_KEY='development-secret'
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm ci
npm run dev
```

Актуальный контракт API находится в `docs/openapi.yaml`.
