CREATE TABLE labels (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    color       TEXT NOT NULL DEFAULT '#888888',
    UNIQUE (project_id, name)
);

CREATE TABLE task_labels (
    task_id   BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id  BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX idx_task_labels_label ON task_labels(label_id);