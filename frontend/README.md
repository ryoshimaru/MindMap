# GoalMind Frontend

## Run locally

1. Open the frontend directory:

   ```powershell
   cd frontend
   ```

2. Install dependencies:

   ```powershell
   npm install
   ```

3. Create an environment file:

   ```powershell
   copy .env.example .env
   ```

4. Start the development server:

   ```powershell
   npm run dev
   ```

## Environment variables

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_USE_MOCKS=true
```

### Real backend mode

- Set `VITE_USE_MOCKS=false`
- Set `VITE_API_BASE_URL` to the backend base URL
- The app will call the backend using the centralized client in `src/api/client.ts`

### Demo mode

- Set `VITE_USE_MOCKS=true`
- The app uses in-memory sample data from `src/api/mockData.ts`
- No backend server is required

## Validation

```powershell
npm run typecheck
npm run build
```
