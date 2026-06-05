# GoalMind

GoalMind is a diploma project for AI-assisted goal decomposition and progress analysis.

## Backend

Run the API in local development mode:

```powershell
$env:GOCACHE='F:\MindMap\.gocache'
go run .\cmd\api
```

The default backend port is `8080`. Override it with:

```powershell
$env:PORT='18080'
```

AI provider selection:

```powershell
$env:AI_PROVIDER='mock'      # optional dev/demo mode, no external API calls
$env:AI_PROVIDER='gemini'
$env:GEMINI_API_KEY='...'
$env:GEMINI_MODEL='gemini-1.5-flash'

$env:AI_PROVIDER='deepseek'
$env:DEEPSEEK_API_KEY='...'
$env:DEEPSEEK_MODEL='deepseek-chat'
```

Production configuration expected by the backend contract:

```powershell
$env:DB_URL='postgres://user:password@localhost:5432/goalmind?sslmode=disable'
$env:JWT_SECRET='replace-me'
$env:GOOGLE_CLIENT_ID='...'
$env:GOOGLE_CLIENT_SECRET='...'
$env:GOOGLE_REDIRECT_URL='http://localhost:8080/api/auth/oauth/google/callback'
$env:FRONTEND_URL='http://localhost:5173'
$env:GEMINI_API_KEY='...'       # fallback when a user has not saved a Gemini key
$env:DEEPSEEK_API_KEY='...'     # fallback when a user has not saved a DeepSeek key
```

Primary product flow:

1. `POST /api/requests/analyze`
2. `POST /api/requests/{requestId}/answers`
3. `POST /api/requests/{requestId}/evaluate`
4. `POST /api/requests/{requestId}/generate-plan`

The backend evaluates feasibility before building the task tree. Plans are generated only when `feasibility.canGeneratePlan` is `true`. Otherwise the API returns an error envelope and does not create goals/tasks.

User AI keys:

- `GET /api/ai-settings`
- `PUT /api/ai-settings`

API keys are not returned by the API; responses contain only `provider` and `maskedKey`.

## Database

PostgreSQL schema migrations are in `migrations/` and use Goose comments.

Run PostgreSQL with Docker:

```powershell
docker run --name goalmind-postgres -e POSTGRES_USER=goalmind -e POSTGRES_PASSWORD=goalmind -e POSTGRES_DB=goalmind -p 5432:5432 -d postgres:16
```

Apply migrations:

```powershell
$env:DB_URL='postgres://goalmind:goalmind@localhost:5432/goalmind?sslmode=disable'
goose -dir migrations postgres $env:DB_URL up
```

When `DB_URL` or `DATABASE_URL` is set, the backend uses PostgreSQL for auth, AI settings, request sessions, answers, feasibility, goals, task trees, generations, versions, and feedback. If neither variable is set, the backend logs a warning and falls back to the in-memory store for local development only.
