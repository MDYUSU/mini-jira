import { request } from '../lib/http'
import type {
  ActivityLog,
  AuthResponse,
  Comment,
  CreateTaskInput,
  Label,
  MemberWorkload,
  Organization,
  Project,
  ProjectDetail,
  ProjectMember,
  ProjectStats,
  Role,
  Task,
  TaskDetail,
  TaskPage,
  TaskQuery,
  TaskTreeNode,
  UpdateTaskInput,
  User,
  UserProfile,
  UserWithProfile,
} from './types'

export const api = {
  login: (email: string, password: string) =>
    request<AuthResponse>('POST', '/auth/login', { body: { email, password }, skipAuthRedirect: true }),
  register: (email: string, name: string, password: string) =>
    request<AuthResponse>('POST', '/auth/register', { body: { email, name, password }, skipAuthRedirect: true }),
  me: () => request<UserWithProfile>('GET', '/auth/me'),

  listUsers: () => request<User[]>('GET', '/users'),
  updateUser: (id: number, body: { name?: string; email?: string }) => request<User>('PATCH', `/users/${id}`, { body }),
  updateProfile: (id: number, body: { bio?: string; avatar_url?: string }) =>
    request<UserProfile>('PUT', `/users/${id}/profile`, { body }),

  createOrg: (name: string, ownerId: number) =>
    request<Organization>('POST', '/orgs', { body: { name, owner_id: ownerId } }),
  getOrg: (id: number) => request<Organization>('GET', `/orgs/${id}`),
  listOrgProjects: (id: number) => request<Project[]>('GET', `/orgs/${id}/projects`),

  createProject: (orgId: number, ownerId: number, body: { name: string; key: string; description?: string }) =>
    request<Project>('POST', `/orgs/${orgId}/projects`, { body: { ...body, owner_id: ownerId } }),
  getProject: (id: number) => request<ProjectDetail>('GET', `/projects/${id}`),
  updateProject: (id: number, body: { name?: string; key?: string; description?: string }) =>
    request<Project>('PATCH', `/projects/${id}`, { body }),
  deleteProject: (id: number) => request<void>('DELETE', `/projects/${id}`),
  projectStats: (id: number) => request<ProjectStats>('GET', `/projects/${id}/stats`),
  projectWorkload: (id: number) => request<MemberWorkload[]>('GET', `/projects/${id}/workload`),

  addMember: (projectId: number, userId: number, role: Role) =>
    request<ProjectMember>('POST', `/projects/${projectId}/members`, { body: { user_id: userId, role } }),
  updateMemberRole: (projectId: number, userId: number, role: Role) =>
    request<ProjectMember>('PATCH', `/projects/${projectId}/members/${userId}`, { body: { role } }),
  removeMember: (projectId: number, userId: number) =>
    request<void>('DELETE', `/projects/${projectId}/members/${userId}`),

  listTasks: (projectId: number, query: TaskQuery) =>
    request<TaskPage>('GET', `/projects/${projectId}/tasks`, { query: { ...query } }),
  createTask: (projectId: number, reporterId: number, body: CreateTaskInput) =>
    request<Task>('POST', `/projects/${projectId}/tasks`, { body: { ...body, reporter_id: reporterId } }),
  getTask: (id: number) => request<TaskDetail>('GET', `/tasks/${id}`),
  updateTask: (id: number, actorId: number, body: UpdateTaskInput) =>
    request<Task>('PATCH', `/tasks/${id}`, { body: { ...body, actor_id: actorId } }),
  deleteTask: (id: number) => request<void>('DELETE', `/tasks/${id}`),
  taskActivity: (id: number) => request<ActivityLog[]>('GET', `/tasks/${id}/activity`),
  taskTree: (id: number) => request<TaskTreeNode>('GET', `/tasks/${id}/tree`),

  listLabels: (projectId: number) => request<Label[]>('GET', `/projects/${projectId}/labels`),
  createLabel: (projectId: number, name: string, color: string) =>
    request<Label>('POST', `/projects/${projectId}/labels`, { body: { name, color } }),
  deleteLabel: (id: number) => request<void>('DELETE', `/labels/${id}`),
  attachLabel: (taskId: number, labelId: number, actorId: number) =>
    request<Label[]>('POST', `/tasks/${taskId}/labels`, { body: { label_id: labelId, actor_id: actorId } }),
  detachLabel: (taskId: number, labelId: number, actorId: number) =>
    request<void>('DELETE', `/tasks/${taskId}/labels/${labelId}`, { query: { actor_id: actorId } }),

  createComment: (taskId: number, authorId: number, body: string) =>
    request<Comment>('POST', `/tasks/${taskId}/comments`, { body: { author_id: authorId, body } }),
  updateComment: (id: number, authorId: number, body: string) =>
    request<Comment>('PATCH', `/comments/${id}`, { body: { author_id: authorId, body } }),
  deleteComment: (id: number, actorId: number) =>
    request<void>('DELETE', `/comments/${id}`, { query: { actor_id: actorId } }),
}
