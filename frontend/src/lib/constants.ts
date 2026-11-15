import type { Priority, Role, Status } from '../api/types'

export const STATUSES: Status[] = ['todo', 'in_progress', 'in_review', 'done']
export const PRIORITIES: Priority[] = ['low', 'medium', 'high', 'urgent']
export const ROLES: Role[] = ['owner', 'member', 'viewer']

export const STATUS_LABEL: Record<Status, string> = {
  todo: 'To do',
  in_progress: 'In progress',
  in_review: 'In review',
  done: 'Done',
}

export const STATUS_STYLE: Record<Status, string> = {
  todo: 'bg-slate-100 text-slate-700',
  in_progress: 'bg-blue-100 text-blue-800',
  in_review: 'bg-amber-100 text-amber-800',
  done: 'bg-emerald-100 text-emerald-800',
}

export const STATUS_BAR: Record<Status, string> = {
  todo: 'bg-slate-400',
  in_progress: 'bg-blue-500',
  in_review: 'bg-amber-500',
  done: 'bg-emerald-500',
}

export const PRIORITY_LABEL: Record<Priority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
}

export const PRIORITY_STYLE: Record<Priority, string> = {
  low: 'bg-slate-100 text-slate-600',
  medium: 'bg-sky-100 text-sky-800',
  high: 'bg-orange-100 text-orange-800',
  urgent: 'bg-red-100 text-red-800',
}

export const ROLE_STYLE: Record<Role, string> = {
  owner: 'bg-violet-100 text-violet-800',
  member: 'bg-blue-100 text-blue-800',
  viewer: 'bg-slate-100 text-slate-600',
}

export const DEMO_ACCOUNTS = ['mohammed@aiotrix.com', 'ravi@aiotrix.com', 'priya@aiotrix.com']
export const DEMO_PASSWORD = 'password123'
