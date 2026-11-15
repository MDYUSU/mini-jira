export type Status = 'todo' | 'in_progress' | 'in_review' | 'done'
export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type Role = 'owner' | 'member' | 'viewer'

export interface User {
  id: number
  email: string
  name: string
  created_at: string
  updated_at: string
}

export interface UserProfile {
  user_id: number
  bio: string | null
  avatar_url: string | null
  updated_at: string
}

export interface UserWithProfile extends User {
  profile: UserProfile | null
}

export interface AuthResponse {
  token: string
  expires_at: string
  user: User
}

export interface UserRef {
  id: number
  name: string
}

export interface Organization {
  id: number
  name: string
  owner_id: number
  owner_name: string
  created_at: string
}

export interface Project {
  id: number
  org_id: number
  name: string
  key: string
  description: string | null
  created_at: string
}

export interface ProjectMember {
  user_id: number
  name: string
  email: string
  role: Role
  joined_at: string
}

export interface ProjectDetail extends Project {
  org_name: string
  members: ProjectMember[]
}

export interface Task {
  id: number
  project_id: number
  parent_task_id: number | null
  title: string
  description: string | null
  status: Status
  priority: Priority
  assignee: UserRef | null
  reporter: UserRef
  due_date: string | null
  created_at: string
  updated_at: string
}

export interface Label {
  id: number
  name: string
  color: string
}

export interface Comment {
  id: number
  author: UserRef | null
  body: string
  created_at: string
}

export interface TaskDetail extends Task {
  labels: Label[]
  subtasks: Task[]
  comments: Comment[]
}

export interface TaskTreeNode extends Task {
  depth: number
  children: TaskTreeNode[]
}

export interface TaskPage {
  items: Task[]
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface ActivityLog {
  id: number
  actor: UserRef | null
  action: string
  field: string | null
  old_value: string | null
  new_value: string | null
  created_at: string
}

export interface ProjectStats {
  project_id: number
  total: number
  by_status: Record<Status, number>
  by_priority: Record<Priority, number>
  overdue: number
  unassigned: number
  done_percent: number
}

export interface MemberWorkload {
  user_id: number
  name: string
  role: Role
  open_tasks: number
  done_tasks: number
  high_priority_open: number
  overdue: number
}

export type SortField = 'created_at' | 'due_date' | 'priority' | 'status' | 'title'

export interface TaskQuery {
  status?: Status | ''
  priority?: Priority | ''
  assignee?: string
  label?: string
  search?: string
  top_level?: boolean
  sort?: SortField
  order?: 'asc' | 'desc'
  page?: number
  limit?: number
}

export interface CreateTaskInput {
  title: string
  description?: string
  status?: Status
  priority?: Priority
  assignee_id?: number
  parent_task_id?: number
  due_date?: string
}

export interface UpdateTaskInput {
  title?: string
  description?: string
  status?: Status
  priority?: Priority
  assignee_id?: number
  due_date?: string
}
