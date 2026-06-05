# GoalMind API Overview

## Purpose

`docs/openapi.yaml` is the source-of-truth backend contract for the GoalMind frontend.

The frontend expects:

- Bearer JWT authentication
- JSON responses wrapped as `{ "data": ... }`
- JSON errors wrapped as `{ "error": { "code", "message", "details" } }`
- UUID identifiers for goals, tasks, generations, and versions

## Main endpoint groups

### Requests: primary product flow

- `POST /api/requests/analyze`
- `POST /api/requests/{requestId}/answers`
- `POST /api/requests/{requestId}/evaluate`
- `POST /api/requests/{requestId}/generate-plan`

This is the main scenario for the application. The user enters a freeform request such as:

`изучить Golang за 3 дня`

The backend analyzes it with an AI provider interface. In the MVP implementation this is a mock AI provider, but the boundary is ready for a later Gemini integration.

The analysis response returns:

- `requestId`
- `interpretedGoal`
- `clarifyingQuestions`
- `activityTracker`

Clarifying questions are structured for UI rendering, not chat messages. Supported question types are:

- `text`
- `number`
- `date`
- `single_select`
- `multi_select`

After answers are saved, `evaluate` re-checks feasibility. `generate-plan` creates a normal `Goal` plus a `TaskTreeNode[]` only when feasibility is `FEASIBLE` or `RISKY`. For `INFEASIBLE`, `UNSAFE`, or `NEEDS_CLARIFICATION`, the backend returns an error envelope and does not create goals or tasks.

Users can configure AI keys through:

- `GET /api/ai-settings`
- `PUT /api/ai-settings`

The API never returns the raw key, only `provider` and `maskedKey`.

### Auth

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET /api/auth/oauth/google/start`
- `GET /api/auth/oauth/google/callback`

The frontend stores the issued JWT in `localStorage` for MVP purposes and sends it as:

`Authorization: Bearer <token>`

If the backend returns `401`, the frontend clears the token and redirects the user to `/login`.

### Goals and context

- `GET /api/goals`
- `POST /api/goals`
- `GET /api/goals/{goalId}`
- `PUT /api/goals/{goalId}`
- `DELETE /api/goals/{goalId}`
- `GET/POST/PUT /api/goals/{goalId}/context`

The goal context is used to improve AI decomposition quality. The frontend expects a single context record per goal.

### Tasks and dependencies

- `GET /api/goals/{goalId}/tasks`
- `POST /api/goals/{goalId}/tasks`
- `GET/PUT/DELETE /api/tasks/{taskId}`
- `PATCH /api/tasks/{taskId}/status`
- `GET/POST /api/tasks/{taskId}/dependencies`
- `DELETE /api/tasks/{taskId}/dependencies/{dependencyId}`

`GET /api/goals/{goalId}/tasks` should return a tree of `TaskTreeNode[]`, not only a flat task list.

### AI decomposition

- `POST /api/goals/{goalId}/decompose`
- `POST /api/goals/{goalId}/regenerate`
- `POST /api/tasks/{taskId}/decompose`

The frontend uses these endpoints to:

- create the first AI plan for a goal
- regenerate a refined plan version
- expand a single task into more detailed subtasks

### Analytics

- `GET /api/goals/{goalId}/progress`
- `GET /api/goals/{goalId}/feasibility`
- `GET /api/goals/{goalId}/quality`
- `GET /api/goals/{goalId}/metrics`

These endpoints power the right-side insight cards on the goal details page.

### Versions, generations, and feedback

- `GET /api/goals/{goalId}/versions`
- `GET /api/goals/{goalId}/versions/{versionId}`
- `POST /api/goals/{goalId}/versions/{versionId}/activate`
- `GET /api/goals/{goalId}/generations`
- `GET /api/generations/{generationId}`
- `POST /api/generations/{generationId}/feedback`

The frontend assumes a goal can have multiple plan versions and multiple AI generation records, with exactly one active version at a time.

## Backend implementation notes

- Return ISO 8601 `date-time` strings for timestamps.
- Return `date` strings for deadlines.
- Support CORS for the frontend origin during development.
- Keep enum values exactly as documented in `openapi.yaml`.
- Preserve stable UUIDs for all entity identifiers.
- For OAuth, redirect the user back to the frontend route `/auth/callback?token=...`.

## Demo mode vs real backend

The frontend supports a frontend-only demo mode controlled by:

`VITE_USE_MOCKS=true`

In demo mode:

- no backend server is required
- the UI uses in-memory mock data from `frontend/src/api/mockData.ts`
- the same typed endpoint layer is used by the pages

When `VITE_USE_MOCKS=false`, the frontend sends real HTTP requests to `VITE_API_BASE_URL`.
