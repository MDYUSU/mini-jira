import type { ActivityLog, Priority, Status } from '../api/types'
import { PRIORITY_LABEL, STATUS_LABEL } from './constants'
import { formatDueDate } from './format'

function statusText(v: string | null): string {
  return v && v in STATUS_LABEL ? STATUS_LABEL[v as Status] : (v ?? 'none')
}

function priorityText(v: string | null): string {
  return v && v in PRIORITY_LABEL ? PRIORITY_LABEL[v as Priority] : (v ?? 'none')
}

function truncate(v: string | null, max = 60): string {
  if (!v) return ''
  return v.length > max ? `${v.slice(0, max)}...` : v
}

export function describeActivity(log: ActivityLog, nameOf: (id: string | null) => string): string {
  const actor = log.actor?.name ?? 'Deleted user'
  switch (log.action) {
    case 'created':
      return `${actor} created the task`
    case 'title_changed':
      return `${actor} renamed the task from "${truncate(log.old_value)}" to "${truncate(log.new_value)}"`
    case 'description_changed':
      return log.new_value ? `${actor} updated the description` : `${actor} cleared the description`
    case 'status_changed':
      return `${actor} changed status from ${statusText(log.old_value)} to ${statusText(log.new_value)}`
    case 'priority_changed':
      return `${actor} changed priority from ${priorityText(log.old_value)} to ${priorityText(log.new_value)}`
    case 'assigned':
      return log.old_value
        ? `${actor} reassigned the task from ${nameOf(log.old_value)} to ${nameOf(log.new_value)}`
        : `${actor} assigned the task to ${nameOf(log.new_value)}`
    case 'unassigned':
      return `${actor} unassigned ${nameOf(log.old_value)}`
    case 'due_date_changed':
      if (!log.new_value) return `${actor} cleared the due date`
      if (!log.old_value) return `${actor} set the due date to ${formatDueDate(log.new_value)}`
      return `${actor} changed the due date from ${formatDueDate(log.old_value)} to ${formatDueDate(log.new_value)}`
    case 'label_added':
      return `${actor} added label ${log.new_value}`
    case 'label_removed':
      return `${actor} removed label ${log.old_value}`
    default:
      return `${actor} ${log.action.replace(/_/g, ' ')}`
  }
}
