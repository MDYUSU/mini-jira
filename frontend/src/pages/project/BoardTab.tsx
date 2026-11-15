import { useState, type DragEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import { keys, useInvalidate, useLabels, useTaskLabelMap, useTasks, type Permissions } from '../../api/queries'
import type { ProjectDetail, Status, TaskPage, TaskQuery } from '../../api/types'
import { useCurrentUser } from '../../components/AuthGate'
import { EMPTY_FILTERS, TaskFilters, type FilterState } from '../../components/TaskFilters'
import { TaskCard } from '../../components/TaskCard'
import { TaskFormModal } from '../../components/TaskFormModal'
import { Button, ErrorBox, Loading, Spinner } from '../../components/ui'
import { STATUSES, STATUS_BAR, STATUS_LABEL } from '../../lib/constants'
import { useDebounced } from '../../hooks/useDebounced'

export function filtersToQuery(filters: FilterState, search: string): TaskQuery {
  return {
    priority: filters.priority,
    assignee: filters.assignee,
    label: filters.label,
    search: search.trim(),
    top_level: filters.hideSubtasks,
  }
}

export function BoardTab({ project, perms }: { project: ProjectDetail; perms: Permissions }) {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const invalidate = useInvalidate()
  const [filters, setFilters] = useState<FilterState>(EMPTY_FILTERS)
  const [creating, setCreating] = useState<Status | null>(null)
  const [dragOver, setDragOver] = useState<Status | null>(null)
  const search = useDebounced(filters.search)

  const query: TaskQuery = { ...filtersToQuery(filters, search), limit: 100 }
  const tasks = useTasks(project.id, query)
  const labels = useLabels(project.id)
  const labelMap = useTaskLabelMap(project.id, labels.data)

  const move = useMutation({
    mutationFn: ({ taskId, status }: { taskId: number; status: Status }) =>
      api.updateTask(taskId, user.id, { status }),
    onMutate: async ({ taskId, status }) => {
      const key = keys.tasks(project.id, query)
      await qc.cancelQueries({ queryKey: key })
      const previous = qc.getQueryData<TaskPage>(key)
      if (previous) {
        qc.setQueryData<TaskPage>(key, {
          ...previous,
          items: previous.items.map((t) => (t.id === taskId ? { ...t, status } : t)),
        })
      }
      return { previous, key }
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(ctx.key, ctx.previous)
    },
    onSettled: () => invalidate.projectAndTasks(project.id),
  })

  const onDragStart = (e: DragEvent, taskId: number, status: Status) => {
    e.dataTransfer.setData('text/plain', JSON.stringify({ taskId, status }))
    e.dataTransfer.effectAllowed = 'move'
  }

  const onDrop = (e: DragEvent, status: Status) => {
    e.preventDefault()
    setDragOver(null)
    try {
      const data = JSON.parse(e.dataTransfer.getData('text/plain')) as { taskId: number; status: Status }
      if (data.status !== status) move.mutate({ taskId: data.taskId, status })
    } catch {
      return
    }
  }

  const items = tasks.data?.items ?? []

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <TaskFilters value={filters} onChange={setFilters} members={project.members} labels={labels.data ?? []} />
        {perms.canEdit && <Button onClick={() => setCreating('todo')}>New task</Button>}
      </div>

      <ErrorBox error={tasks.error} className="mb-3" />
      <ErrorBox error={move.error} className="mb-3" />
      {tasks.data && tasks.data.total > tasks.data.items.length && (
        <p className="mb-3 text-xs text-slate-500">
          Showing the first {tasks.data.items.length} of {tasks.data.total} tasks. Use filters or the List tab to see the rest.
        </p>
      )}

      {tasks.isPending ? (
        <Loading text="Loading tasks..." />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          {STATUSES.map((status) => {
            const column = items.filter((t) => t.status === status)
            return (
              <div
                key={status}
                onDragOver={(e) => {
                  if (!perms.canEdit) return
                  e.preventDefault()
                  setDragOver(status)
                }}
                onDragLeave={() => setDragOver((s) => (s === status ? null : s))}
                onDrop={(e) => perms.canEdit && onDrop(e, status)}
                className={`flex min-h-[420px] flex-col rounded-lg bg-slate-100 p-2 transition ${dragOver === status ? 'ring-2 ring-blue-400' : ''}`}
              >
                <div className="mb-2 flex items-center justify-between px-1">
                  <div className="flex items-center gap-2">
                    <span className={`h-2 w-2 rounded-full ${STATUS_BAR[status]}`} />
                    <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-600">{STATUS_LABEL[status]}</h3>
                    <span className="rounded-full bg-slate-200 px-2 text-xs text-slate-600">{column.length}</span>
                    {tasks.isFetching && <Spinner className="h-3 w-3" />}
                  </div>
                  {perms.canEdit && (
                    <button
                      onClick={() => setCreating(status)}
                      className="rounded px-1.5 text-lg leading-none text-slate-400 hover:bg-slate-200 hover:text-slate-700"
                      aria-label={`Add task to ${STATUS_LABEL[status]}`}
                    >
                      +
                    </button>
                  )}
                </div>
                <div className="flex flex-1 flex-col gap-2">
                  {column.length === 0 && <p className="px-1 py-4 text-center text-xs text-slate-400">No tasks yet</p>}
                  {column.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      projectKey={project.key}
                      labels={labelMap.data?.get(task.id) ?? []}
                      draggable={perms.canEdit}
                      onOpen={() => navigate(`/tasks/${task.id}`)}
                      onDragStart={(e) => onDragStart(e, task.id, task.status)}
                    />
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {creating && (
        <TaskFormModal
          projectId={project.id}
          members={project.members}
          defaultStatus={creating}
          onClose={() => setCreating(null)}
        />
      )}
    </div>
  )
}
