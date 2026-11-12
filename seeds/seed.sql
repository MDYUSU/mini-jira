TRUNCATE users, user_profiles, organizations, projects, project_members, tasks, labels, task_labels, comments, activity_logs RESTART IDENTITY CASCADE;

INSERT INTO users (email, name) VALUES
('mohammed@aiotrix.com', 'Mohammed'),
('ravi@aiotrix.com', 'Ravi'),
('priya@aiotrix.com', 'Priya'),
('arjun@aiotrix.com', 'Arjun');

INSERT INTO user_profiles (user_id, bio) VALUES
(1, 'Backend learner'),
(2, 'Full stack dev');

INSERT INTO organizations (name, owner_id) VALUES
('Aiotrix', 1),
('Bix Bytes', 3);

INSERT INTO projects (org_id, name, key) VALUES
(1, 'ART Backend', 'ART'),
(1, 'Website', 'WEB'),
(2, 'Internal Tools', 'INT');
DELETE /users/5
INSERT INTO project_members (project_id, user_id, role) VALUES
(1, 1, 'owner'),
(1, 2, 'member'),
(1, 3, 'viewer'),
(2, 1, 'owner'),
(2, 2, 'member');

INSERT INTO tasks (project_id, parent_task_id, title, status, priority, assignee_id, reporter_id) VALUES
(1, NULL, 'Build login API', 'in_progress', 'high', 2, 1),
(1, NULL, 'Design database schema', 'done', 'medium', 1, 1),
(1, 1, 'Write JWT middleware', 'todo', 'high', 2, 1),
(1, 1, 'Add refresh tokens', 'todo', 'medium', NULL, 2),
(1, 3, 'Test token expiry', 'todo', 'low', NULL, 2),
(1, NULL, 'Setup CI pipeline', 'todo', 'urgent', 1, 3),
(2, NULL, 'Redesign homepage', 'in_review', 'medium', 2, 1);

INSERT INTO labels (project_id, name, color) VALUES
(1, 'backend', '#2563eb'),
(1, 'urgent', '#dc2626'),
(1, 'security', '#7c3aed'),
(2, 'frontend', '#16a34a');

INSERT INTO task_labels (task_id, label_id) VALUES
(1, 1), (1, 3), (3, 1), (3, 3), (6, 2), (7, 4);

INSERT INTO comments (task_id, author_id, body) VALUES
(1, 3, 'Any ETA?'),
(1, 2, 'Should be done by Friday'),
(3, 1, 'Use short-lived access tokens');

INSERT INTO activity_logs (task_id, actor_id, action, field, old_value, new_value) VALUES
(1, 1, 'assigned', 'assignee_id', NULL, '2'),
(1, 2, 'status_changed', 'status', 'todo', 'in_progress');