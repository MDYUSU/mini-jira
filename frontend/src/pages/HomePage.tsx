import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { api } from '../api/client'
import { useInvalidate, useOrg, useOrgProjects } from '../api/queries'
import type { Organization } from '../api/types'
import { useCurrentUser } from '../components/AuthGate'
import { Button, Card, EmptyState, ErrorBox, Field, Input, Loading, Modal, Textarea } from '../components/ui'
import { forgetOrg, rememberOrg } from '../lib/orgs'
import { useRememberedOrgs } from '../hooks/useRememberedOrgs'
import { formatDate } from '../lib/format'

function OpenOrgForm() {
  const navigate = useNavigate()
  const [value, setValue] = useState('')
  const submit = (e: FormEvent) => {
    e.preventDefault()
    const id = Number(value)
    if (Number.isInteger(id) && id > 0) navigate(`/orgs/${id}`)
  }
  return (
    <form onSubmit={submit} className="flex gap-2">
      <Input type="number" min={1} placeholder="Organization ID, e.g. 1" value={value} onChange={(e) => setValue(e.target.value)} />
      <Button type="submit" variant="secondary" disabled={!value}>
        Open
      </Button>
    </form>
  )
}

function CreateOrgForm() {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const create = useMutation({
    mutationFn: () => api.createOrg(name.trim(), user.id),
    onSuccess: (org) => {
      rememberOrg(user.id, { id: org.id, name: org.name })
      setName('')
      navigate(`/orgs/${org.id}`)
    },
  })
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        create.mutate()
      }}
      className="space-y-2"
    >
      <div className="flex gap-2">
        <Input placeholder="New organization name" value={name} onChange={(e) => setName(e.target.value)} />
        <Button type="submit" disabled={!name.trim() || create.isPending}>
          Create
        </Button>
      </div>
      <ErrorBox error={create.error} />
    </form>
  )
}

function NewProjectModal({ org, onClose }: { org: Organization; onClose: () => void }) {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const invalidate = useInvalidate()
  const [name, setName] = useState('')
  const [key, setKey] = useState('')
  const [description, setDescription] = useState('')

  const create = useMutation({
    mutationFn: () =>
      api.createProject(org.id, user.id, {
        name: name.trim(),
        key: key.trim(),
        ...(description.trim() ? { description: description.trim() } : {}),
      }),
    onSuccess: async (project) => {
      await invalidate.org(org.id)
      onClose()
      navigate(`/projects/${project.id}`)
    },
  })

  return (
    <Modal title={`New project in ${org.name}`} onClose={onClose}>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          create.mutate()
        }}
        className="space-y-4"
      >
        <Field label="Name">
          <Input value={name} onChange={(e) => setName(e.target.value)} autoFocus required />
        </Field>
        <Field label="Key" hint="2 to 10 letters or digits, starting with a letter">
          <Input value={key} onChange={(e) => setKey(e.target.value.toUpperCase())} maxLength={10} required />
        </Field>
        <Field label="Description">
          <Textarea rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
        </Field>
        <ErrorBox error={create.error} />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={create.isPending}>
            Create project
          </Button>
        </div>
      </form>
    </Modal>
  )
}

function OrgView({ orgId }: { orgId: number }) {
  const user = useCurrentUser()
  const navigate = useNavigate()
  const org = useOrg(orgId)
  const projects = useOrgProjects(orgId)
  const [showNew, setShowNew] = useState(false)

  useEffect(() => {
    if (org.data) rememberOrg(user.id, { id: org.data.id, name: org.data.name })
  }, [org.data, user.id])

  if (org.isPending) return <Loading />
  if (org.isError) return <ErrorBox error={org.error} />

  return (
    <div>
      <div className="mb-5 flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Organization #{org.data.id}</p>
          <h2 className="text-2xl font-bold text-slate-800">{org.data.name}</h2>
          <p className="mt-1 text-sm text-slate-500">
            Owner: <span className="font-medium text-slate-700">{org.data.owner_name}</span> &middot; Created{' '}
            {formatDate(org.data.created_at)}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="ghost"
            onClick={() => {
              forgetOrg(user.id, orgId)
              navigate('/')
            }}
          >
            Remove from sidebar
          </Button>
          <Button onClick={() => setShowNew(true)}>New project</Button>
        </div>
      </div>

      {projects.isPending && <Loading />}
      <ErrorBox error={projects.error} />
      {projects.data?.length === 0 && (
        <EmptyState title="No projects yet">Create the first project in this organization.</EmptyState>
      )}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {projects.data?.map((p) => (
          <Link key={p.id} to={`/projects/${p.id}`}>
            <Card className="h-full p-4 transition-shadow hover:border-blue-300 hover:shadow-md">
              <div className="mb-2 flex items-center gap-2">
                <span className="rounded bg-blue-100 px-1.5 py-0.5 text-xs font-bold text-blue-700">{p.key}</span>
                <h3 className="truncate font-semibold text-slate-800">{p.name}</h3>
              </div>
              <p className="line-clamp-2 text-sm text-slate-500">{p.description || 'No description'}</p>
              <p className="mt-3 text-xs text-slate-400">Created {formatDate(p.created_at)}</p>
            </Card>
          </Link>
        ))}
      </div>

      {showNew && <NewProjectModal org={org.data} onClose={() => setShowNew(false)} />}
    </div>
  )
}

export function HomePage() {
  const user = useCurrentUser()
  const { orgId } = useParams()
  const remembered = useRememberedOrgs(user.id)
  const selected = orgId ? Number(orgId) : null

  return (
    <div className="mx-auto max-w-6xl p-6">
      <h1 className="mb-1 text-xl font-bold text-slate-800">Welcome, {user.name}</h1>
      <p className="mb-6 text-sm text-slate-500">Open an organization by ID or create a new one.</p>

      <div className="mb-8 grid gap-4 md:grid-cols-2">
        <Card className="p-4">
          <h2 className="mb-3 text-sm font-semibold text-slate-700">Open organization</h2>
          <OpenOrgForm />
          {remembered.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-2">
              {remembered.map((o) => (
                <Link
                  key={o.id}
                  to={`/orgs/${o.id}`}
                  className={`rounded-full border px-3 py-1 text-xs ${selected === o.id ? 'border-blue-300 bg-blue-50 text-blue-700' : 'border-slate-200 text-slate-600 hover:bg-slate-50'}`}
                >
                  {o.name} <span className="text-slate-400">#{o.id}</span>
                </Link>
              ))}
            </div>
          )}
        </Card>
        <Card className="p-4">
          <h2 className="mb-3 text-sm font-semibold text-slate-700">Create organization</h2>
          <CreateOrgForm />
        </Card>
      </div>

      {selected !== null && Number.isInteger(selected) && selected > 0 ? (
        <OrgView key={selected} orgId={selected} />
      ) : (
        <EmptyState title="No organization selected">Pick one from the list or the sidebar.</EmptyState>
      )}
    </div>
  )
}
