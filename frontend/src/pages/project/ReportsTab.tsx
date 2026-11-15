import { useProjectStats, useWorkload } from '../../api/queries'
import type { ProjectDetail } from '../../api/types'
import { Avatar, Card, EmptyState, ErrorBox, Loading, RoleBadge } from '../../components/ui'
import { PRIORITIES, PRIORITY_LABEL, PRIORITY_STYLE, STATUSES, STATUS_BAR, STATUS_LABEL } from '../../lib/constants'

function Tile({ label, value, tone = 'text-slate-800' }: { label: string; value: string | number; tone?: string }) {
  return (
    <Card className="p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">{label}</p>
      <p className={`mt-1 text-3xl font-bold ${tone}`}>{value}</p>
    </Card>
  )
}

export function ReportsTab({ project }: { project: ProjectDetail }) {
  const stats = useProjectStats(project.id)
  const workload = useWorkload(project.id)

  if (stats.isPending) return <Loading />
  if (stats.isError) return <ErrorBox error={stats.error} />
  const s = stats.data
  const busiest = [...(workload.data ?? [])].sort((a, b) => b.open_tasks - a.open_tasks || a.name.localeCompare(b.name))

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Tile label="Total tasks" value={s.total} />
        <Tile label="Done" value={`${Math.round(s.done_percent)}%`} tone="text-emerald-600" />
        <Tile label="Overdue" value={s.overdue} tone={s.overdue > 0 ? 'text-red-600' : 'text-slate-800'} />
        <Tile label="Unassigned" value={s.unassigned} tone={s.unassigned > 0 ? 'text-amber-600' : 'text-slate-800'} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="p-4">
          <h3 className="mb-3 text-sm font-semibold text-slate-700">By status</h3>
          {s.total === 0 ? (
            <p className="text-sm text-slate-400">No tasks yet</p>
          ) : (
            <div className="flex h-4 overflow-hidden rounded-full bg-slate-100">
              {STATUSES.map((st) =>
                s.by_status[st] > 0 ? (
                  <div
                    key={st}
                    className={STATUS_BAR[st]}
                    style={{ width: `${(s.by_status[st] / s.total) * 100}%` }}
                    title={`${STATUS_LABEL[st]}: ${s.by_status[st]}`}
                  />
                ) : null,
              )}
            </div>
          )}
          <ul className="mt-4 grid grid-cols-2 gap-2 text-sm">
            {STATUSES.map((st) => (
              <li key={st} className="flex items-center gap-2">
                <span className={`h-2.5 w-2.5 rounded-full ${STATUS_BAR[st]}`} />
                <span className="text-slate-600">{STATUS_LABEL[st]}</span>
                <span className="ml-auto font-semibold text-slate-800">{s.by_status[st]}</span>
              </li>
            ))}
          </ul>
        </Card>
        <Card className="p-4">
          <h3 className="mb-3 text-sm font-semibold text-slate-700">By priority</h3>
          <ul className="space-y-2.5">
            {PRIORITIES.map((p) => {
              const n = s.by_priority[p]
              const pct = s.total ? (n / s.total) * 100 : 0
              return (
                <li key={p} className="flex items-center gap-3 text-sm">
                  <span className={`w-16 rounded px-1.5 py-0.5 text-center text-xs font-medium ${PRIORITY_STYLE[p]}`}>{PRIORITY_LABEL[p]}</span>
                  <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                    <div className="h-full rounded-full bg-slate-400" style={{ width: `${pct}%` }} />
                  </div>
                  <span className="w-6 text-right font-semibold text-slate-800">{n}</span>
                </li>
              )
            })}
          </ul>
        </Card>
      </div>

      <Card>
        <div className="border-b border-slate-200 px-4 py-3">
          <h3 className="text-sm font-semibold text-slate-700">Workload</h3>
        </div>
        {workload.isPending ? (
          <Loading />
        ) : workload.isError ? (
          <ErrorBox error={workload.error} className="m-4" />
        ) : busiest.length === 0 ? (
          <div className="p-4">
            <EmptyState title="No members" />
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-2.5 font-semibold">Member</th>
                <th className="px-4 py-2.5 font-semibold">Role</th>
                <th className="px-4 py-2.5 text-right font-semibold">Open</th>
                <th className="px-4 py-2.5 text-right font-semibold">High priority open</th>
                <th className="px-4 py-2.5 text-right font-semibold">Overdue</th>
                <th className="px-4 py-2.5 text-right font-semibold">Done</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {busiest.map((w) => (
                <tr key={w.user_id}>
                  <td className="px-4 py-2.5">
                    <span className="flex items-center gap-2">
                      <Avatar name={w.name} id={w.user_id} size="sm" />
                      {w.name}
                    </span>
                  </td>
                  <td className="px-4 py-2.5">
                    <RoleBadge role={w.role} />
                  </td>
                  <td className="px-4 py-2.5 text-right font-semibold">{w.open_tasks}</td>
                  <td className="px-4 py-2.5 text-right">{w.high_priority_open}</td>
                  <td className={`px-4 py-2.5 text-right ${w.overdue > 0 ? 'font-semibold text-red-600' : ''}`}>{w.overdue}</td>
                  <td className="px-4 py-2.5 text-right text-emerald-700">{w.done_tasks}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  )
}
