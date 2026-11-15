import type { DragEvent } from 'react'
import type { Label, Task } from '../api/types'
import { formatDueDate, isOverdue } from '../lib/format'
import { Avatar, LabelChip, PriorityBadge } from './ui'

interface Props {
  task: Task
  projectKey: string
  labels: Label[]
  draggable: boolean
  onOpen: () => void
  onDragStart: (e: DragEvent) => void
}

export function TaskCard({ task, projectKey, labels, draggable, onOpen, onDragStart }: Props) {
  const overdue = isOverdue(task.due_date, task.status)
  return (
    <div
      draggable={draggable}
      onDragStart={onDragStart}
      onClick={onOpen}
      className={`group cursor-pointer rounded-md border border-slate-200 bg-white p-3 shadow-sm transition hover:border-blue-300 hover:shadow ${draggable ? 'active:cursor-grabbing' : ''}`}
    >
      {task.parent_task_id && (
        <div className="mb-1 flex items-center gap-1 text-[11px] font-medium text-slate-400">
          <svg viewBox="0 0 20 20" className="h-3 w-3" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M5 3v8a3 3 0 0 0 3 3h8m-3-3 3 3-3 3" />
          </svg>
          Subtask of {projectKey}-{task.parent_task_id}
        </div>
      )}
      <p className="text-sm font-medium text-slate-800 group-hover:text-blue-700">{task.title}</p>
      {labels.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {labels.map((l) => (
            <LabelChip key={l.id} label={l} />
          ))}
        </div>
      )}
      <div className="mt-3 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className="text-[11px] font-semibold text-slate-400">
            {projectKey}-{task.id}
          </span>
          <PriorityBadge priority={task.priority} />
        </div>
        <div className="flex items-center gap-2">
          {task.due_date && (
            <span className={`text-xs ${overdue ? 'font-semibold text-red-600' : 'text-slate-500'}`}>
              {formatDueDate(task.due_date)}
            </span>
          )}
          {task.assignee ? (
            <Avatar name={task.assignee.name} id={task.assignee.id} size="sm" />
          ) : (
            <span className="inline-flex h-6 w-6 items-center justify-center rounded-full border border-dashed border-slate-300 text-[10px] text-slate-400" title="Unassigned">
              ?
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
