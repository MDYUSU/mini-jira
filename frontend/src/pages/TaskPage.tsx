import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import {
  keys,
  permissionsFor,
  useActivity,
  useInvalidate,
  useLabels,
  useProject,
  useTask,
  useTaskTree,
  useUsers,
  type Permissions,
} from '../api/queries'
import type { Comment, Priority, ProjectDetail, Status, TaskDetail, TaskTreeNode, UpdateTaskInput } from '../api/types'
import { useCurrentUser } from '../components/AuthGate'
import { useConfirm } from '../components/Confirm'
import { TaskFormModal } from '../components/TaskFormModal'
import {
  Avatar,
  Button,
  Card,
  ErrorBox,
  Input,
  LabelChip,
  Loading,
  PriorityBadge,
  Select,
  Spinner,
  StatusBadge,
  Textarea,
} from '../components/ui'
import { describeActivity } from '../lib/activity'
import { PRIORITIES, PRIORITY_LABEL, STATUSES, STATUS_LABEL } from '../lib/constants'
import { dueDateToInput, formatDateTime, formatDueDate, isOverdue, relativeTime } from '../lib/format'

function Section({ title, action, children }: { title: string; action?: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="mb-8">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">{title}</h2>
        {action}
      </div>
      {children}
    </section>
  )
}

function SideField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</p>
      {children}
    </div>
  )
}

function useTaskUpdate(task: TaskDetail) {
  const user = useCurrentUser()
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (body: UpdateTaskInput) => api.updateTask(task.id, user.id, body),
    onSettled: () => invalidate.projectAndTasks(task.project_id),
  })
}

function TitleEditor({ task, canEdit }: { task: TaskDetail; canEdit: boolean }) {
  const update = useTaskUpdate(task)
  const [value, setValue] = useState(task.title)
  useEffect(() => setValue(task.title), [task.title])

  const save = () => {
    const next = value.trim()
    if (!next) {
      setValue(task.title)
      return
    }
    if (next !== task.title) update.mutate({ title: next })
  }

  if (!canEdit) return <h1 className="text-2xl font-bold text-slate-800">{task.title}</h1>
  return (
    <div>
      <input
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onBlur={save}
        onKeyDown={(e) => {
          if (e.key === 'Enter') e.currentTarget.blur()
          if (e.key === 'Escape') {
            setValue(task.title)
            e.currentTarget.blur()
          }
        }}
        className="w-full rounded-md border border-transparent px-2 py-1 -ml-2 text-2xl font-bold text-slate-800 hover:border-slate-200 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
      />
      <ErrorBox error={update.error} className="mt-2" />
    </div>
  )
}

function DescriptionEditor({ task, canEdit }: { task: TaskDetail; canEdit: boolean }) {
  const update = useTaskUpdate(task)
  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState(task.description ?? '')

  if (!editing) {
    return (
      <div>
        <div
          onClick={() => {
            if (!canEdit) return
            setValue(task.description ?? '')
            setEditing(true)
          }}
          className={`whitespace-pre-wrap rounded-md p-2 -ml-2 text-sm ${task.description ? 'text-slate-700' : 'italic text-slate-400'} ${canEdit ? 'cursor-text hover:bg-slate-100' : ''}`}
        >
          {task.description || (canEdit ? 'Add a description...' : 'No description')}
        </div>
        <ErrorBox error={update.error} className="mt-2" />
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <Textarea rows={5} value={value} onChange={(e) => setValue(e.target.value)} autoFocus />
      <div className="flex gap-2">
        <Button
          size="sm"
          disabled={update.isPending}
          onClick={() => {
            if (value !== (task.description ?? '')) {
              update.mutate({ description: value }, { onSuccess: () => setEditing(false) })
            } else {
              setEditing(false)
            }
          }}
        >
          Save
        </Button>
        <Button size="sm" variant="secondary" onClick={() => setEditing(false)}>
          Cancel
        </Button>
      </div>
      <ErrorBox error={update.error} />
    </div>
  )
}

function DetailsPanel({ task, project, perms }: { task: TaskDetail; project: ProjectDetail; perms: Permissions }) {
  const update = useTaskUpdate(task)
  const disabled = !perms.canEdit || update.isPending
  const assigneeIsMember = !task.assignee || project.members.some((m) => m.user_id === task.assignee!.id)
  const overdue = isOverdue(task.due_date, task.status)

  return (
    <Card className="space-y-4 p-4">
      <SideField label="Status">
        <Select value={task.status} disabled={disabled} onChange={(e) => update.mutate({ status: e.target.value as Status })}>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {STATUS_LABEL[s]}
            </option>
          ))}
        </Select>
      </SideField>
      <SideField label="Priority">
        <Select value={task.priority} disabled={disabled} onChange={(e) => update.mutate({ priority: e.target.value as Priority })}>
          {PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {PRIORITY_LABEL[p]}
            </option>
          ))}
        </Select>
      </SideField>
      <SideField label="Assignee">
        <Select
          value={task.assignee?.id ?? ''}
          disabled={disabled}
          onChange={(e) => update.mutate({ assignee_id: e.target.value ? Number(e.target.value) : 0 })}
        >
          <option value="">Unassigned</option>
          {!assigneeIsMember && task.assignee && <option value={task.assignee.id}>{task.assignee.name}</option>}
          {project.members.map((m) => (
            <option key={m.user_id} value={m.user_id}>
              {m.name}
            </option>
          ))}
        </Select>
      </SideField>
      <SideField label="Due date">
        <div className="flex gap-2">
          <Input
            type="date"
            value={dueDateToInput(task.due_date)}
            disabled={disabled}
            className={overdue ? 'border-red-300 text-red-600' : ''}
            onChange={(e) => {
              if (e.target.value) update.mutate({ due_date: e.target.value })
            }}
          />
          {task.due_date && perms.canEdit && (
            <Button variant="secondary" disabled={disabled} onClick={() => update.mutate({ due_date: '' })}>
              Clear
            </Button>
          )}
        </div>
        {overdue && <p className="mt-1 text-xs font-medium text-red-600">Overdue</p>}
      </SideField>
      <SideField label="Reporter">
        <span className="flex items-center gap-2 text-sm text-slate-700">
          <Avatar name={task.reporter.name} id={task.reporter.id} size="sm" />
          {task.reporter.name}
        </span>
      </SideField>
      <div className="space-y-1 border-t border-slate-100 pt-3 text-xs text-slate-500">
        <p>Created {formatDateTime(task.created_at)}</p>
        <p>Updated {relativeTime(task.updated_at)}</p>
      </div>
      {update.isPending && <Spinner />}
      <ErrorBox error={update.error} />
    </Card>
  )
}

function LabelsPanel({ task, perms }: { task: TaskDetail; perms: Permissions }) {
  const user = useCurrentUser()
  const invalidate = useInvalidate()
  const labels = useLabels(task.project_id)
  const attachedIds = new Set(task.labels.map((l) => l.id))
  const available = (labels.data ?? []).filter((l) => !attachedIds.has(l.id))

  const attach = useMutation({
    mutationFn: (labelId: number) => api.attachLabel(task.id, labelId, user.id),
    onSettled: () => invalidate.projectAndTasks(task.project_id),
  })
  const detach = useMutation({
    mutationFn: (labelId: number) => api.detachLabel(task.id, labelId, user.id),
    onSettled: () => invalidate.projectAndTasks(task.project_id),
  })

  return (
    <Card className="p-4">
      <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">Labels</p>
      <div className="flex flex-wrap gap-1.5">
        {task.labels.length === 0 && <span className="text-sm text-slate-400">No labels</span>}
        {task.labels.map((l) => (
          <LabelChip key={l.id} label={l} onRemove={perms.canEdit ? () => detach.mutate(l.id) : undefined} />
        ))}
      </div>
      {perms.canEdit && (
        <Select
          className="mt-3"
          value=""
          disabled={attach.isPending || available.length === 0}
          onChange={(e) => e.target.value && attach.mutate(Number(e.target.value))}
        >
          <option value="">
            {labels.data?.length === 0 ? 'No labels in this project' : available.length === 0 ? 'All labels added' : 'Add a label...'}
          </option>
          {available.map((l) => (
            <option key={l.id} value={l.id}>
              {l.name}
            </option>
          ))}
        </Select>
      )}
      <ErrorBox error={attach.error || detach.error} className="mt-2" />
    </Card>
  )
}

function TreeNode({ node, currentId, projectKey }: { node: TaskTreeNode; currentId: number; projectKey: string }) {
  return (
    <li>
      <div
        className={`flex items-center gap-2 rounded px-2 py-1.5 text-sm ${node.id === currentId ? 'bg-blue-50' : 'hover:bg-slate-50'}`}
        style={{ paddingLeft: `${node.depth * 20 + 8}px` }}
      >
        {node.depth > 0 && <span className="text-slate-300">&#x2514;</span>}
        <span className="text-xs font-semibold text-slate-400">
          {projectKey}-{node.id}
        </span>
        {node.id === currentId ? (
          <span className="font-semibold text-slate-800">{node.title}</span>
        ) : (
          <Link to={`/tasks/${node.id}`} className="text-slate-700 hover:text-blue-700 hover:underline">
            {node.title}
          </Link>
        )}
        <span className="ml-auto">
          <StatusBadge status={node.status} />
        </span>
      </div>
      {node.children.length > 0 && (
        <ul>
          {node.children.map((c) => (
            <TreeNode key={c.id} node={c} currentId={currentId} projectKey={projectKey} />
          ))}
        </ul>
      )}
    </li>
  )
}

function SubtasksSection({ task, project, perms }: { task: TaskDetail; project: ProjectDetail; perms: Permissions }) {
  const [adding, setAdding] = useState(false)
  const [showTree, setShowTree] = useState(false)
  const tree = useTaskTree(task.id, showTree)

  return (
    <Section
      title={`Subtasks (${task.subtasks.length})`}
      action={
        <div className="flex gap-2">
          <Button size="sm" variant="secondary" onClick={() => setShowTree(!showTree)}>
            {showTree ? 'Hide full tree' : 'Show full tree'}
          </Button>
          {perms.canEdit && (
            <Button size="sm" onClick={() => setAdding(true)}>
              Add subtask
            </Button>
          )}
        </div>
      }
    >
      {showTree ? (
        <Card className="p-2">
          {tree.isPending ? (
            <Loading />
          ) : tree.isError ? (
            <ErrorBox error={tree.error} />
          ) : (
            <ul>
              <TreeNode node={tree.data} currentId={task.id} projectKey={project.key} />
            </ul>
          )}
        </Card>
      ) : task.subtasks.length === 0 ? (
        <p className="text-sm text-slate-400">No subtasks yet</p>
      ) : (
        <Card>
          <ul className="divide-y divide-slate-100">
            {task.subtasks.map((s) => (
              <li key={s.id} className="flex items-center gap-3 px-3 py-2 text-sm">
                <span className="text-xs font-semibold text-slate-400">
                  {project.key}-{s.id}
                </span>
                <Link to={`/tasks/${s.id}`} className="flex-1 text-slate-700 hover:text-blue-700 hover:underline">
                  {s.title}
                </Link>
                <PriorityBadge priority={s.priority} />
                {s.assignee && <Avatar name={s.assignee.name} id={s.assignee.id} size="sm" />}
                <StatusBadge status={s.status} />
              </li>
            ))}
          </ul>
        </Card>
      )}
      {adding && (
        <TaskFormModal
          projectId={project.id}
          members={project.members}
          parentTaskId={task.id}
          parentTitle={task.title}
          onClose={() => setAdding(false)}
        />
      )}
    </Section>
  )
}

function CommentItem({ comment, perms }: { comment: Comment; perms: Permissions }) {
  const user = useCurrentUser()
  const confirm = useConfirm()
  const invalidate = useInvalidate()
  const [editing, setEditing] = useState(false)
  const [body, setBody] = useState(comment.body)
  const isAuthor = comment.author?.id === user.id

  const save = useMutation({
    mutationFn: () => api.updateComment(comment.id, user.id, body.trim()),
    onSuccess: async () => {
      setEditing(false)
      await invalidate.tasks()
    },
  })
  const remove = useMutation({
    mutationFn: () => api.deleteComment(comment.id, user.id),
    onSuccess: () => invalidate.tasks(),
  })

  return (
    <li className="flex gap-3">
      {comment.author ? (
        <Avatar name={comment.author.name} id={comment.author.id} />
      ) : (
        <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-300 text-xs text-white">?</span>
      )}
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 text-sm">
          <span className="font-semibold text-slate-800">{comment.author?.name ?? 'Deleted user'}</span>
          <span className="text-xs text-slate-400" title={formatDateTime(comment.created_at)}>
            {relativeTime(comment.created_at)}
          </span>
        </div>
        {editing ? (
          <div className="mt-1 space-y-2">
            <Textarea rows={3} value={body} onChange={(e) => setBody(e.target.value)} autoFocus />
            <div className="flex gap-2">
              <Button size="sm" disabled={!body.trim() || save.isPending} onClick={() => save.mutate()}>
                Save
              </Button>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => {
                  setBody(comment.body)
                  setEditing(false)
                }}
              >
                Cancel
              </Button>
            </div>
          </div>
        ) : (
          <p className="mt-0.5 whitespace-pre-wrap text-sm text-slate-700">{comment.body}</p>
        )}
        {!editing && (isAuthor || perms.isOwner) && (
          <div className="mt-1 flex gap-3 text-xs">
            {isAuthor && (
              <button onClick={() => setEditing(true)} className="text-slate-500 hover:text-blue-700">
                Edit
              </button>
            )}
            <button
              disabled={remove.isPending}
              onClick={async () => {
                const ok = await confirm({ title: 'Delete comment', message: 'Delete this comment? This cannot be undone.' })
                if (ok) remove.mutate()
              }}
              className="text-slate-500 hover:text-red-600"
            >
              Delete
            </button>
          </div>
        )}
        <ErrorBox error={save.error || remove.error} className="mt-2" />
      </div>
    </li>
  )
}

function CommentsSection({ task, perms }: { task: TaskDetail; perms: Permissions }) {
  const user = useCurrentUser()
  const invalidate = useInvalidate()
  const [body, setBody] = useState('')

  const add = useMutation({
    mutationFn: () => api.createComment(task.id, user.id, body.trim()),
    onSuccess: async () => {
      setBody('')
      await invalidate.tasks()
    },
  })

  return (
    <Section title={`Comments (${task.comments.length})`}>
      {perms.isMember ? (
        <form
          onSubmit={(e) => {
            e.preventDefault()
            add.mutate()
          }}
          className="mb-5 flex gap-3"
        >
          <Avatar name={user.name} id={user.id} />
          <div className="flex-1 space-y-2">
            <Textarea rows={2} placeholder="Add a comment..." value={body} onChange={(e) => setBody(e.target.value)} />
            {body.trim() && (
              <Button size="sm" type="submit" disabled={add.isPending}>
                Comment
              </Button>
            )}
            <ErrorBox error={add.error} />
          </div>
        </form>
      ) : (
        <p className="mb-4 text-sm text-slate-400">Only project members can comment.</p>
      )}
      {task.comments.length === 0 ? (
        <p className="text-sm text-slate-400">No comments yet</p>
      ) : (
        <ul className="space-y-5">
          {task.comments.map((c) => (
            <CommentItem key={c.id} comment={c} perms={perms} />
          ))}
        </ul>
      )}
    </Section>
  )
}

function ActivitySection({ task, project }: { task: TaskDetail; project: ProjectDetail }) {
  const activity = useActivity(task.id)
  const users = useUsers()

  const nameOf = (id: string | null): string => {
    if (!id) return 'nobody'
    const n = Number(id)
    return (
      project.members.find((m) => m.user_id === n)?.name ??
      users.data?.find((u) => u.id === n)?.name ??
      `user #${id}`
    )
  }

  return (
    <Section title="Activity">
      {activity.isPending ? (
        <Loading />
      ) : activity.isError ? (
        <ErrorBox error={activity.error} />
      ) : activity.data.length === 0 ? (
        <p className="text-sm text-slate-400">No activity yet</p>
      ) : (
        <ol className="relative ml-3 border-l border-slate-200">
          {[...activity.data].reverse().map((log) => (
            <li key={log.id} className="mb-4 ml-5">
              <span className="absolute -left-[5px] mt-1.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-blue-500" />
              <p className="text-sm text-slate-700">{describeActivity(log, nameOf)}</p>
              <p className="text-xs text-slate-400" title={formatDateTime(log.created_at)}>
                {relativeTime(log.created_at)}
              </p>
            </li>
          ))}
        </ol>
      )}
    </Section>
  )
}

export function TaskPage() {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const confirm = useConfirm()
  const qc = useQueryClient()
  const invalidate = useInvalidate()
  const taskId = Number(useParams().taskId)
  const task = useTask(taskId)
  const project = useProject(task.data?.project_id ?? null)

  const remove = useMutation({
    mutationFn: () => api.deleteTask(taskId),
    onSuccess: async () => {
      const projectId = task.data!.project_id
      const parentId = task.data!.parent_task_id
      qc.removeQueries({ queryKey: keys.task(taskId) })
      await invalidate.projectAndTasks(projectId)
      navigate(parentId ? `/tasks/${parentId}` : `/projects/${projectId}`, { replace: true })
    },
  })

  if (task.isPending || (task.data && project.isPending)) return <Loading />
  if (task.isError) {
    return (
      <div className="p-6">
        <ErrorBox error={task.error} />
      </div>
    )
  }
  if (project.isError) {
    return (
      <div className="p-6">
        <ErrorBox error={project.error} />
      </div>
    )
  }

  const t = task.data
  const p = project.data!
  const perms = permissionsFor(p, user.id)

  const onDelete = async () => {
    const extra = t.subtasks.length > 0 ? ' All of its subtasks will be deleted too.' : ''
    const ok = await confirm({ title: 'Delete task', message: `Delete "${t.title}"?${extra} This cannot be undone.` })
    if (ok) remove.mutate()
  }

  return (
    <div className="mx-auto max-w-6xl p-6">
      <div className="mb-3 flex items-center justify-between gap-4">
        <div className="text-sm text-slate-500">
          <Link to={`/orgs/${p.org_id}`} className="hover:underline">
            {p.org_name}
          </Link>{' '}
          /{' '}
          <Link to={`/projects/${p.id}`} className="hover:underline">
            {p.name}
          </Link>{' '}
          {t.parent_task_id && (
            <>
              /{' '}
              <Link to={`/tasks/${t.parent_task_id}`} className="hover:underline">
                {p.key}-{t.parent_task_id}
              </Link>{' '}
            </>
          )}
          / <span className="font-medium text-slate-700">{p.key}-{t.id}</span>
        </div>
        {perms.canEdit && (
          <Button variant="danger" size="sm" onClick={onDelete} disabled={remove.isPending}>
            Delete task
          </Button>
        )}
      </div>
      <ErrorBox error={remove.error} className="mb-3" />
      {perms.role === 'viewer' && (
        <div className="mb-4 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
          You are a viewer on this project. You can read and comment, but not edit the task.
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[1fr_300px]">
        <div className="min-w-0">
          <div className="mb-2 flex items-center gap-2">
            <StatusBadge status={t.status} />
            <PriorityBadge priority={t.priority} />
            {t.due_date && (
              <span className={`text-xs ${isOverdue(t.due_date, t.status) ? 'font-semibold text-red-600' : 'text-slate-500'}`}>
                Due {formatDueDate(t.due_date)}
              </span>
            )}
          </div>
          <TitleEditor task={t} canEdit={perms.canEdit} />
          <div className="mb-8 mt-4">
            <h2 className="mb-1 text-sm font-semibold uppercase tracking-wide text-slate-500">Description</h2>
            <DescriptionEditor key={t.updated_at} task={t} canEdit={perms.canEdit} />
          </div>
          <SubtasksSection task={t} project={p} perms={perms} />
          <CommentsSection task={t} perms={perms} />
          <ActivitySection task={t} project={p} />
        </div>
        <div className="space-y-4">
          <DetailsPanel task={t} project={p} perms={perms} />
          <LabelsPanel task={t} perms={perms} />
        </div>
      </div>
    </div>
  )
}
