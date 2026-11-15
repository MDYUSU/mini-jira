import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../api/client'
import { useInvalidate, useUsers, type Permissions } from '../../api/queries'
import type { ProjectDetail, ProjectMember, Role } from '../../api/types'
import { useCurrentUser } from '../../components/AuthGate'
import { useConfirm } from '../../components/Confirm'
import { Avatar, Button, Card, ErrorBox, RoleBadge, Select } from '../../components/ui'
import { ROLES } from '../../lib/constants'
import { ApiError } from '../../lib/http'
import { formatDate } from '../../lib/format'

function friendlyError(err: unknown): unknown {
  if (err instanceof ApiError && err.status === 409 && err.message.includes('owner')) {
    return 'A project must always keep at least one owner. Make someone else an owner first.'
  }
  return err
}

function AddMemberForm({ project }: { project: ProjectDetail }) {
  const invalidate = useInvalidate()
  const users = useUsers()
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState<Role>('member')
  const memberIds = new Set(project.members.map((m) => m.user_id))
  const candidates = (users.data ?? []).filter((u) => !memberIds.has(u.id))

  const add = useMutation({
    mutationFn: () => api.addMember(project.id, Number(userId), role),
    onSuccess: async () => {
      setUserId('')
      setRole('member')
      await invalidate.project(project.id)
    },
  })

  return (
    <Card className="mb-4 p-4">
      <h3 className="mb-3 text-sm font-semibold text-slate-700">Add a member</h3>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          add.mutate()
        }}
        className="flex flex-wrap gap-2"
      >
        <Select className="w-72" value={userId} onChange={(e) => setUserId(e.target.value)} disabled={users.isPending}>
          <option value="">{users.isPending ? 'Loading users...' : candidates.length ? 'Choose a user' : 'Everyone is already a member'}</option>
          {candidates.map((u) => (
            <option key={u.id} value={u.id}>
              {u.name} ({u.email})
            </option>
          ))}
        </Select>
        <Select className="w-32" value={role} onChange={(e) => setRole(e.target.value as Role)}>
          {ROLES.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </Select>
        <Button type="submit" disabled={!userId || add.isPending}>
          Add member
        </Button>
      </form>
      <ErrorBox error={users.error || add.error} className="mt-3" />
    </Card>
  )
}

function MemberRow({ project, member, isOwner }: { project: ProjectDetail; member: ProjectMember; isOwner: boolean }) {
  const user = useCurrentUser()
  const confirm = useConfirm()
  const invalidate = useInvalidate()

  const changeRole = useMutation({
    mutationFn: (role: Role) => api.updateMemberRole(project.id, member.user_id, role),
    onSettled: () => invalidate.project(project.id),
  })

  const remove = useMutation({
    mutationFn: () => api.removeMember(project.id, member.user_id),
    onSuccess: () => invalidate.projectAndTasks(project.id),
  })

  const onRemove = async () => {
    const self = member.user_id === user.id
    const ok = await confirm({
      title: 'Remove member',
      message: `Remove ${self ? 'yourself' : member.name} from ${project.name}? Their tasks in this project will become unassigned.`,
      confirmText: 'Remove',
    })
    if (ok) remove.mutate()
  }

  const error = changeRole.error || remove.error

  return (
    <li className="px-4 py-3">
      <div className="flex items-center gap-3">
        <Avatar name={member.name} id={member.user_id} />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-slate-800">
            {member.name}
            {member.user_id === user.id && <span className="ml-2 text-xs text-slate-400">(you)</span>}
          </p>
          <p className="text-xs text-slate-500">
            {member.email} &middot; joined {formatDate(member.joined_at)}
          </p>
        </div>
        {isOwner ? (
          <div className="flex items-center gap-2">
            <Select
              className="w-28 py-1"
              value={member.role}
              disabled={changeRole.isPending}
              onChange={(e) => changeRole.mutate(e.target.value as Role)}
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </Select>
            <Button size="sm" variant="ghost" className="text-red-600 hover:bg-red-50" onClick={onRemove} disabled={remove.isPending}>
              Remove
            </Button>
          </div>
        ) : (
          <RoleBadge role={member.role} />
        )}
      </div>
      {error && <ErrorBox error={friendlyError(error)} className="mt-2" />}
    </li>
  )
}

export function MembersTab({ project, perms }: { project: ProjectDetail; perms: Permissions }) {
  return (
    <div className="max-w-3xl">
      {perms.isOwner && <AddMemberForm project={project} />}
      <Card>
        <div className="border-b border-slate-200 px-4 py-3">
          <h3 className="text-sm font-semibold text-slate-700">
            {project.members.length} member{project.members.length === 1 ? '' : 's'}
          </h3>
        </div>
        <ul className="divide-y divide-slate-100">
          {project.members.map((m) => (
            <MemberRow key={m.user_id} project={project} member={m} isOwner={perms.isOwner} />
          ))}
        </ul>
      </Card>
    </div>
  )
}
