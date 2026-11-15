# Mini Jira API

A backend-only task tracker, built to learn Go and PostgreSQL with a focus on relational data modeling. The API can be explored through Swagger UI or Postman, or through the React frontend in `frontend/`.

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.22+ (standard library `net/http`, no framework) |
| Database | PostgreSQL 16, running in Docker |
| DB access | `database/sql` with the `pgx` driver, hand-written SQL |
| Migrations | `golang-migrate` with plain `.sql` files |
| API docs | OpenAPI 3 spec served through Swagger UI at `/docs` |

## Getting started

### Prerequisites

- Go 1.22 or newer
- Docker with Docker Compose
- `golang-migrate` CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 1. Configure

```bash
cp .env.example .env
```

If port 5432 is already in use on your machine, change both `DB_PORT` and the port inside `DATABASE_URL` in `.env` (for example to 5433).

### 2. Start Postgres

```bash
docker compose up -d
```

### 3. Create the schema

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/minijira?sslmode=disable" up
```

### 4. Load sample data (optional)

```bash
cat seeds/seed.sql | docker exec -i minijira-db psql -U postgres -d minijira -v ON_ERROR_STOP=1
```

This wipes all data and loads a fixed set of users, organizations, projects, tasks, labels and comments. Run it any time to reset.

### 5. Run the server

```bash
go run ./cmd/api
```

- Health check: http://localhost:8080/health
- API docs: http://localhost:8080/docs

## Frontend

A React + TypeScript web app lives in `frontend/` (Vite, Tailwind CSS, React Router, TanStack Query). It has a kanban board with drag and drop, a sortable list, task details with subtasks, labels, comments and an activity timeline, member management and project reports.

### Run it

Requires Node.js 20 or newer.

1. Start the backend (Postgres must be running, see above):

   ```bash
   go run ./cmd/api
   ```

2. In a second terminal, start the frontend:

   ```bash
   cd frontend && npm install && npm run dev
   ```

3. Open the URL Vite prints (usually http://localhost:5173).

Log in with a seeded account such as `mohammed@aiotrix.com`, `ravi@aiotrix.com` or `priya@aiotrix.com` (password `password123`).

### How it talks to the API

The backend sends no CORS headers, so the browser only talks to the Vite dev server. Vite proxies every `/api/*` request to `http://localhost:8080/*` with the `/api` prefix removed (see `frontend/vite.config.ts`).

The JWT is kept in `localStorage`. All token reads and writes are in `frontend/src/lib/auth.ts`, so moving to httpOnly cookies later only touches that file. Any 401 response clears the token and sends you back to the login page.

Because there is no "list my organizations" endpoint yet, the home page opens an organization by ID and remembers every organization you open or create in `localStorage`, so they show up in the sidebar.

`npm run build` type-checks the code and writes a production build to `frontend/dist/`.

## Data model

Ten tables, chosen to cover every common relation type.

| Table | Relation | Delete behavior |
|---|---|---|
| `users` | base entity | |
| `user_profiles` | 1:1 with users | cascade |
| `organizations` | many per user (owner) | restrict owner delete |
| `projects` | many per organization | cascade |
| `project_members` | M:N users and projects, with a role | cascade |
| `tasks` | many per project; self-referencing subtasks | cascade |
| `tasks.assignee_id` | composite FK to `project_members` | set null |
| `tasks.reporter_id` | many per user | restrict |
| `labels` | many per project | cascade |
| `task_labels` | M:N tasks and labels | cascade |
| `comments` | many per task | author set null |
| `activity_logs` | change history per task | actor set null |

Rules enforced by the database:

- A subtask must be in the same project as its parent (composite foreign key).
- A task can only be assigned to a member of its project (composite foreign key).
- Project keys are unique per organization; label names are unique per project.
- Status, priority and role are limited to fixed values (check constraints).

## Project structure

```
cmd/api/            entry point: config, wiring, routes
internal/config/    reads .env
internal/database/  connection pool
internal/model/     Go structs returned by the API
internal/repository/ all SQL lives here
internal/service/   validation and business rules
internal/handler/   HTTP: parse request, call service, write JSON
migrations/         versioned schema changes
seeds/              sample data
docs/               OpenAPI spec and Swagger UI
```

A request flows `handler -> service -> repository -> database`.

## API overview

| Area | Endpoints |
|---|---|
| Users | list, get with profile, create, update, delete, upsert profile |
| Organizations | create, get, list projects |
| Projects | create (owner added in the same transaction), get with members, update, delete |
| Members | add, change role, remove |
| Tasks | list with filters/sort/pagination, create, get with labels/subtasks/comments, update, delete, activity history, full subtask tree |
| Labels | list, create, delete, attach to task, detach from task |
| Comments | list, add, edit, delete |
| Reports | project stats, member workload |

Full request and response details are in Swagger at `/docs`.

### Task list query parameters

`GET /projects/{id}/tasks?status=todo&assignee=2&label=security&search=token&top_level=true&sort=due_date&order=desc&page=1&limit=20`

`assignee=none` returns unassigned tasks. `sort` accepts `created_at`, `due_date`, `priority`, `status` or `title`.

## Business rules

- A project must always keep at least one owner.
- Viewers can read tasks and comment, but cannot create or change tasks or labels.
- Only a comment's author can edit it; the author or a project owner can delete it.
- Every task change (status, priority, assignee, title, description, due date, labels) is written to `activity_logs` in the same transaction as the change.

## Error responses

All errors are JSON: `{"error": "message"}`.

| Status | Meaning |
|---|---|
| 400 | Invalid input |
| 403 | Not allowed (for example, a viewer editing a task) |
| 404 | Not found |
| 409 | Conflicts with existing data (duplicate, last owner, delete blocked) |
| 500 | Unexpected server error (details are logged, not returned) |

## Not yet implemented

Authentication. Until it is added, the acting user is passed explicitly as `owner_id`, `reporter_id`, `actor_id` or `author_id`.