import { useState } from 'react'
import { Link, NavLink, Outlet, useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useOrgProjects } from '../api/queries'
import { clearToken } from '../lib/auth'
import type { RememberedOrg } from '../lib/orgs'
import { useRememberedOrgs } from '../hooks/useRememberedOrgs'
import { useCurrentUser } from './AuthGate'
import { Logo } from './Logo'
import { Avatar, Spinner } from './ui'

function SidebarOrg({ org }: { org: RememberedOrg }) {
  const [open, setOpen] = useState(true)
  const projects = useOrgProjects(org.id)

  return (
    <div>
      <div className="flex items-center">
        <button
          onClick={() => setOpen(!open)}
          className="flex h-6 w-6 items-center justify-center rounded text-slate-400 hover:bg-slate-200"
          aria-label={open ? 'Collapse' : 'Expand'}
        >
          <svg viewBox="0 0 20 20" className={`h-3 w-3 transition-transform ${open ? 'rotate-90' : ''}`} fill="currentColor">
            <path d="M7 4l6 6-6 6V4Z" />
          </svg>
        </button>
        <NavLink
          to={`/orgs/${org.id}`}
          className={({ isActive }) =>
            `flex-1 truncate rounded px-2 py-1 text-sm font-semibold ${isActive ? 'bg-blue-50 text-blue-700' : 'text-slate-700 hover:bg-slate-200'}`
          }
        >
          {org.name}
        </NavLink>
      </div>
      {open && (
        <div className="ml-6 mt-0.5 space-y-0.5">
          {projects.isPending && <Spinner className="ml-2 h-3 w-3" />}
          {projects.isError && <p className="px-2 text-xs text-red-500">Could not load</p>}
          {projects.data?.length === 0 && <p className="px-2 text-xs text-slate-400">No projects</p>}
          {projects.data?.map((p) => (
            <NavLink
              key={p.id}
              to={`/projects/${p.id}`}
              className={({ isActive }) =>
                `flex items-center gap-2 truncate rounded px-2 py-1 text-sm ${isActive ? 'bg-blue-50 font-medium text-blue-700' : 'text-slate-600 hover:bg-slate-200'}`
              }
            >
              <span className="rounded bg-slate-200 px-1 text-[10px] font-bold text-slate-600">{p.key}</span>
              <span className="truncate">{p.name}</span>
            </NavLink>
          ))}
        </div>
      )}
    </div>
  )
}

export function AppLayout() {
  const user = useCurrentUser()
  const orgs = useRememberedOrgs(user.id)
  const navigate = useNavigate()
  const qc = useQueryClient()
  const params = useParams()

  const logout = () => {
    clearToken()
    qc.clear()
    navigate('/login', { replace: true })
  }

  return (
    <div className="flex h-screen flex-col">
      <header className="flex h-14 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-4">
        <Link to="/" className="flex items-center gap-2">
          <Logo />
          <span className="text-lg font-bold text-slate-800">Mini Jira</span>
        </Link>
        <div className="flex items-center gap-3">
          <Link to="/profile" className="flex items-center gap-2 rounded-md px-2 py-1 hover:bg-slate-100">
            <Avatar name={user.name} id={user.id} size="sm" />
            <span className="text-sm font-medium text-slate-700">{user.name}</span>
          </Link>
          <button onClick={logout} className="rounded-md px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100">
            Logout
          </button>
        </div>
      </header>
      <div className="flex min-h-0 flex-1">
        <aside className="w-64 shrink-0 overflow-y-auto border-r border-slate-200 bg-slate-100/70 p-3">
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `mb-3 block rounded px-2 py-1.5 text-sm font-medium ${isActive ? 'bg-blue-50 text-blue-700' : 'text-slate-700 hover:bg-slate-200'}`
            }
          >
            Home
          </NavLink>
          <p className="mb-1 px-2 text-[11px] font-semibold uppercase tracking-wider text-slate-400">Organizations</p>
          {orgs.length === 0 ? (
            <p className="px-2 text-xs text-slate-500">Open or create an organization from Home.</p>
          ) : (
            <div className="space-y-1">
              {orgs.map((org) => (
                <SidebarOrg key={org.id} org={org} />
              ))}
            </div>
          )}
        </aside>
        <main className="min-w-0 flex-1 overflow-y-auto" key={params.projectId ?? params.taskId ?? 'main'}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
