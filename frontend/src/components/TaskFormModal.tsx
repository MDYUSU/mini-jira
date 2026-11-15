import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../api/client'
import { useInvalidate } from '../api/queries'
import type { CreateTaskInput, Priority, ProjectMember, Status } from '../api/types'
import { PRIORITIES, PRIORITY_LABEL, STATUSES, STATUS_LABEL } from '../lib/constants'
import { useCurrentUser } from './AuthGate'
import { Button, ErrorBox, Field, Input, Modal, Select, Textarea } from './ui'

interface Props {
  projectId: number
  members: ProjectMember[]
  parentTaskId?: number
  parentTitle?: string
  defaultStatus?: Status
  onClose: () => void
  onCreated?: (taskId: number) => void
}

export function TaskFormModal({ projectId, members, parentTaskId, parentTitle, defaultStatus, onClose, onCreated }: Props) {
  const user = useCurrentUser()
  const invalidate = useInvalidate()
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<Status>(defaultStatus ?? 'todo')
  const [priority, setPriority] = useState<Priority>('medium')
  const [assignee, setAssignee] = useState('')
  const [dueDate, setDueDate] = useState('')

  const create = useMutation({
    mutationFn: () => {
      const body: CreateTaskInput = { title: title.trim(), status, priority }
      if (description.trim()) body.description = description.trim()
      if (assignee) body.assignee_id = Number(assignee)
      if (dueDate) body.due_date = dueDate
      if (parentTaskId) body.parent_task_id = parentTaskId
      return api.createTask(projectId, user.id, body)
    },
    onSuccess: async (task) => {
      await invalidate.projectAndTasks(projectId)
      onCreated?.(task.id)
      onClose()
    },
  })

  return (
    <Modal title={parentTaskId ? 'Add subtask' : 'New task'} onClose={onClose} wide>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          create.mutate()
        }}
        className="space-y-4"
      >
        {parentTitle && (
          <p className="rounded-md bg-slate-50 px-3 py-2 text-sm text-slate-600">
            Subtask of <span className="font-medium text-slate-800">{parentTitle}</span>
          </p>
        )}
        <Field label="Title">
          <Input value={title} onChange={(e) => setTitle(e.target.value)} autoFocus required />
        </Field>
        <Field label="Description">
          <Textarea rows={4} value={description} onChange={(e) => setDescription(e.target.value)} />
        </Field>
        <div className="grid grid-cols-2 gap-4">
          <Field label="Status">
            <Select value={status} onChange={(e) => setStatus(e.target.value as Status)}>
              {STATUSES.map((s) => (
                <option key={s} value={s}>
                  {STATUS_LABEL[s]}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Priority">
            <Select value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
              {PRIORITIES.map((p) => (
                <option key={p} value={p}>
                  {PRIORITY_LABEL[p]}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Assignee">
            <Select value={assignee} onChange={(e) => setAssignee(e.target.value)}>
              <option value="">Unassigned</option>
              {members.map((m) => (
                <option key={m.user_id} value={m.user_id}>
                  {m.name}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Due date">
            <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
          </Field>
        </div>
        <ErrorBox error={create.error} />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={!title.trim() || create.isPending}>
            {create.isPending ? 'Creating...' : 'Create task'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
