CREATE TABLE tasks (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_task_id  BIGINT,
    title           TEXT NOT NULL,
    description     TEXT,
    status          TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'in_review', 'done')),
    priority        TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    assignee_id     BIGINT,
    reporter_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    due_date        DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, project_id),
    CHECK (parent_task_id <> id),
    FOREIGN KEY (parent_task_id, project_id) REFERENCES tasks(id, project_id) ON DELETE CASCADE,
    FOREIGN KEY (project_id, assignee_id) REFERENCES project_members(project_id, user_id) ON DELETE SET NULL (assignee_id)
);

CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_reporter ON tasks(reporter_id);
CREATE INDEX idx_tasks_parent ON tasks(parent_task_id);