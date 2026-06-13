-- +goose Up
create table users (
    id uuid primary key,
    email text not null unique,
    name text not null,
    avatar_url text,
    provider text not null,
    password_hash text,
    created_at timestamptz not null default now()
);

create table sessions (
    token text primary key,
    user_id uuid not null references users(id) on delete cascade,
    created_at timestamptz not null default now()
);

create table user_ai_settings (
    user_id uuid primary key references users(id) on delete cascade,
    provider text not null,
    encrypted_api_key text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table user_profiles (
    user_id uuid primary key references users(id) on delete cascade,
    age integer not null check (age between 14 and 120),
    occupation text not null,
    free_hours_per_week numeric not null check (free_hours_per_week > 0 and free_hours_per_week <= 168),
    available_budget numeric not null default 0 check (available_budget >= 0),
    constraints text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table goals (
    id uuid primary key,
    user_id uuid not null references users(id) on delete cascade,
    title text not null,
    description text not null,
    category text not null,
    priority text not null,
    status text not null,
    deadline date,
    available_hours_per_week numeric not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table goal_contexts (
    id uuid primary key,
    goal_id uuid not null unique references goals(id) on delete cascade,
    current_level text not null,
    constraints_text text not null,
    preferences jsonb not null default '[]',
    expected_result text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table tasks (
    id uuid primary key,
    goal_id uuid not null references goals(id) on delete cascade,
    parent_task_id uuid references tasks(id) on delete cascade,
    title text not null,
    description text not null,
    status text not null,
    priority text not null,
    estimated_hours numeric not null default 0,
    deadline date,
    order_index integer not null,
    source text not null,
    edited_by_user boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table task_dependencies (
    id uuid primary key,
    task_id uuid not null references tasks(id) on delete cascade,
    depends_on_task_id uuid not null references tasks(id) on delete cascade,
    created_at timestamptz not null default now(),
    unique (task_id, depends_on_task_id),
    check (task_id <> depends_on_task_id)
);

create table planning_requests (
    id uuid primary key,
    user_id uuid not null references users(id) on delete cascade,
    original_text text not null,
    provider text not null default 'mock',
    interpreted_goal jsonb not null,
    feasibility jsonb not null,
    clarifying_questions jsonb not null,
    answers jsonb not null default '[]',
    activity_tracker jsonb not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table generations (
    id uuid primary key,
    goal_id uuid not null references goals(id) on delete cascade,
    request_id uuid references planning_requests(id) on delete set null,
    status text not null,
    model_name text not null,
    error_message text,
    prompt text not null,
    raw_ai_response text not null,
    created_at timestamptz not null default now()
);

create table plan_versions (
    id uuid primary key,
    goal_id uuid not null references goals(id) on delete cascade,
    generation_id uuid not null references generations(id) on delete cascade,
    version_number integer not null,
    is_active boolean not null default false,
    created_at timestamptz not null default now(),
    unique (goal_id, version_number)
);

create unique index plan_versions_one_active_per_goal
    on plan_versions(goal_id)
    where is_active;

create table roadmap_history (
    id uuid primary key,
    goal_id uuid not null references goals(id) on delete cascade,
    request_id uuid references planning_requests(id) on delete set null,
    generation_id uuid references generations(id) on delete set null,
    plan_version_id uuid references plan_versions(id) on delete set null,
    feasibility jsonb not null,
    snapshot jsonb not null,
    created_at timestamptz not null default now()
);

create table feedback (
    id uuid primary key,
    generation_id uuid not null references generations(id) on delete cascade,
    rating integer not null check (rating between 1 and 5),
    comment text not null,
    created_at timestamptz not null default now()
);

create index goals_user_id_idx on goals(user_id);
create index tasks_goal_id_idx on tasks(goal_id);
create index tasks_parent_task_id_idx on tasks(parent_task_id);
create index task_dependencies_task_id_idx on task_dependencies(task_id);
create index task_dependencies_depends_on_task_id_idx on task_dependencies(depends_on_task_id);
create index generations_goal_id_idx on generations(goal_id);
create index generations_request_id_idx on generations(request_id);
create index planning_requests_user_id_idx on planning_requests(user_id);
create index planning_requests_user_updated_idx on planning_requests(user_id, updated_at desc);
create index roadmap_history_goal_id_idx on roadmap_history(goal_id);
create index roadmap_history_request_id_idx on roadmap_history(request_id);
create index sessions_user_id_idx on sessions(user_id);
