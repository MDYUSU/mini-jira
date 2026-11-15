import { useEffect, useState } from 'react'
import { Link, NavLink, useNavigate, useParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../api/client'
import { permissionsFor, useInvalidate, useProject } from '../../api/queries'
import type { ProjectDetail } from '../../api/types'
import { useCurrentUser } from '../../components/AuthGate'
import { useConfirm } from '../../components/Confirm'
import { Button, ErrorBox, Field, Input, Loading, Modal, RoleBadge, Textarea } from '../../components/ui'
import { rememberOrg } from '../../lib/orgs'
import { BoardTab } from './BoardTab'
import { ListTab } from './ListTab'
import { MembersTab } from './MembersTab'
import { LabelsTab } from './LabelsTab'
import { ReportsTab } from './ReportsTab'

const TABS = [
  { id: 'board', label: 'Board' },
  { id: 'list', label: 'List' },
  { id: 'members', label: 'Members' },
  { id: 'labels', label: 'Labels' },
  { id: 'reports', label: 'Reports' },
] as const

type TabId = (typeof TABS)[number]['id']

function EditProjectModal({ project, onClose }: { project: ProjectDetail; onClose: () => void }) {
  const invalidate = useInvalidate()
  const [name, setName] = useState(project.name)
  const [key, setKey] = useState(project.key)
  const [description, setDescription] = useState(project.description ?? '')

  const save = useMutation({
    mutationFn: () => {
      const body: { name?: string; key?: string; description?: string } = {}
      if (name.trim() !== project.name) body.name = name.trim()
      if (key.trim().toUpperCase() !== project.key) body.key = key.trim()
      if (description !== (project.description ?? '')) body.description = description
      if (Object.keys(body).length === 0) return Promise.resolve(null)
      return api.updateProject(project.id, body)
    },
    onSuccess: async () => {
      await Promise.all([invalidate.project(project.id), invalidate.org(project.org_id)])
      onClose()
    },
  })

  return (
    <Modal title="Edit project" onClose={onClose}>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          save.mutate()
        }}
        className="space-y-4"
      >
        <Field label="Name">
          <Input value={name} onChange={(e) => setName(e.target.value)} required />
        </Field>
        <Field label="Key" hint="2 to 10 letters or digits, starting with a letter">
          <Input value={key} onChange={(e) => setKey(e.target.value.toUpperCase())} maxLength={10} required />
        </Field>
        <Field label="Description">
          <Textarea rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
        </Field>
        <ErrorBox error={save.error} />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={save.isPending}>
            Save
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export function ProjectPage() {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const confirm = useConfirm()
  const invalidate = useInvalidate()
  const params = useParams()
  const projectId = Number(params.projectId)
  const tab: TabId = TABS.some((t) => t.id === params.tab) ? (params.tab as TabId) : 'board'
  const project = useProject(projectId)
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    if (project.data) rememberOrg(user.id, { id: project.data.org_id, name: project.data.org_name })
  }, [project.data, user.id])

  const remove = useMutation({
    mutationFn: () => api.deleteProject(projectId),
    onSuccess: async () => {
      const orgId = project.data!.org_id
      await invalidate.org(orgId)
      navigate(`/orgs/${orgId}`, { replace: true })
    },
  })

  if (project.isPending) return <Loading />
  if (project.isError) {
    return (
      <div className="p-6">
        <ErrorBox error={project.error} />
      </div>
    )
  }

  const p = project.data
  const perms = permissionsFor(p, user.id)

  const onDelete = async () => {
    const ok = await confirm({
      title: 'Delete project',
      message: `Delete "${p.name}" and all of its tasks, labels and comments? This cannot be undone.`,
    })
    if (ok) remove.mutate()
  }

  return (
    <div className="p-6">
      <div className="mb-1 text-sm text-slate-500">
        <Link to={`/orgs/${p.org_id}`} className="hover:underline">
          {p.org_name}
        </Link>{' '}
        / <span>{p.key}</span>
      </div>
      <div className="mb-4 flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-slate-800">{p.name}</h1>
            {perms.role ? <RoleBadge role={perms.role} /> : <span className="text-xs text-slate-400">Not a member</span>}
          </div>
          {p.description && <p className="mt-1 max-w-3xl text-sm text-slate-500">{p.description}</p>}
        </div>
        {perms.isOwner && (
          <div className="flex shrink-0 gap-2">
            <Button variant="secondary" onClick={() => setEditing(true)}>
              Edit project
            </Button>
            <Button variant="danger" onClick={onDelete} disabled={remove.isPending}>
              Delete project
            </Button>
          </div>
        )}
      </div>
      <ErrorBox error={remove.error} className="mb-4" />
      {!perms.isMember && (
        <div className="mb-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          You are not a member of this project, so you can look around but not make changes.
        </div>
      )}
      {perms.role === 'viewer' && (
        <div className="mb-4 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
          You are a viewer on this project. You can read everything and add comments.
        </div>
      )}

      <nav className="mb-5 flex gap-1 border-b border-slate-200">
        {TABS.map((t) => (
          <NavLink
            key={t.id}
            to={`/projects/${projectId}/${t.id}`}
            className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium ${tab === t.id ? 'border-blue-600 text-blue-700' : 'border-transparent text-slate-500 hover:text-slate-800'}`}
          >
            {t.label}
          </NavLink>
        ))}
      </nav>

      {tab === 'board' && <BoardTab project={p} perms={perms} />}
      {tab === 'list' && <ListTab project={p} />}
      {tab === 'members' && <MembersTab project={p} perms={perms} />}
      {tab === 'labels' && <LabelsTab project={p} perms={perms} />}
      {tab === 'reports' && <ReportsTab project={p} />}

      {editing && <EditProjectModal project={p} onClose={() => setEditing(false)} />}
    </div>
  )
}
