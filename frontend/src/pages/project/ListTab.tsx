import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useLabels, useTasks } from '../../api/queries'
import type { ProjectDetail, SortField } from '../../api/types'
import { EMPTY_FILTERS, TaskFilters, type FilterState } from '../../components/TaskFilters'
import { Avatar, Button, Card, EmptyState, ErrorBox, Loading, PriorityBadge, Select, Spinner, StatusBadge } from '../../components/ui'
import { formatDate, formatDueDate, isOverdue } from '../../lib/format'
import { useDebounced } from '../../hooks/useDebounced'
import { filtersToQuery } from './BoardTab'

const COLUMNS: { field: SortField | null; label: string }[] = [
  { field: 'title', label: 'Title' },
  { field: 'status', label: 'Status' },
  { field: 'priority', label: 'Priority' },
  { field: null, label: 'Assignee' },
  { field: 'due_date', label: 'Due date' },
  { field: 'created_at', label: 'Created' },
]

export function ListTab({ project }: { project: ProjectDetail }) {
  const [filters, setFilters] = useState<FilterState>(EMPTY_FILTERS)
  const [sort, setSort] = useState<SortField>('created_at')
  const [order, setOrder] = useState<'asc' | 'desc'>('desc')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const search = useDebounced(filters.search)
  const labels = useLabels(project.id)

  const tasks = useTasks(project.id, { ...filtersToQuery(filters, search), sort, order, page, limit })

  const toggleSort = (field: SortField) => {
    if (sort === field) {
      setOrder(order === 'asc' ? 'desc' : 'asc')
    } else {
      setSort(field)
      setOrder('asc')
    }
    setPage(1)
  }

  const data = tasks.data

  return (
    <div>
      <div className="mb-4">
        <TaskFilters
          value={filters}
          onChange={(next) => {
            setFilters(next)
            setPage(1)
          }}
          members={project.members}
          labels={labels.data ?? []}
        />
      </div>
      <ErrorBox error={tasks.error} className="mb-3" />
      {tasks.isPending ? (
        <Loading />
      ) : data && data.total === 0 ? (
        <EmptyState title="No tasks yet" />
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-4 py-2.5 font-semibold">Key</th>
                  {COLUMNS.map((c) => (
                    <th key={c.label} className="px-4 py-2.5 font-semibold">
                      {c.field ? (
                        <button onClick={() => toggleSort(c.field!)} className="flex items-center gap-1 uppercase hover:text-slate-800">
                          {c.label}
                          {sort === c.field && <span className="text-blue-600">{order === 'asc' ? '↑' : '↓'}</span>}
                        </button>
                      ) : (
                        c.label
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {data?.items.map((t) => (
                  <tr key={t.id} className="hover:bg-slate-50">
                    <td className="whitespace-nowrap px-4 py-2.5 text-xs font-semibold text-slate-400">
                      {project.key}-{t.id}
                    </td>
                    <td className="px-4 py-2.5">
                      <Link to={`/tasks/${t.id}`} className="font-medium text-slate-800 hover:text-blue-700 hover:underline">
                        {t.title}
                      </Link>
                      {t.parent_task_id && <span className="ml-2 text-xs text-slate-400">subtask</span>}
                    </td>
                    <td className="px-4 py-2.5">
                      <StatusBadge status={t.status} />
                    </td>
                    <td className="px-4 py-2.5">
                      <PriorityBadge priority={t.priority} />
                    </td>
                    <td className="px-4 py-2.5">
                      {t.assignee ? (
                        <span className="flex items-center gap-2">
                          <Avatar name={t.assignee.name} id={t.assignee.id} size="sm" />
                          {t.assignee.name}
                        </span>
                      ) : (
                        <span className="text-slate-400">Unassigned</span>
                      )}
                    </td>
                    <td className={`whitespace-nowrap px-4 py-2.5 ${isOverdue(t.due_date, t.status) ? 'font-semibold text-red-600' : 'text-slate-600'}`}>
                      {t.due_date ? formatDueDate(t.due_date) : <span className="text-slate-300">None</span>}
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5 text-slate-500">{formatDate(t.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {data && (
            <div className="flex items-center justify-between border-t border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-600">
              <div className="flex items-center gap-3">
                <span>
                  {data.total} task{data.total === 1 ? '' : 's'}
                </span>
                {tasks.isFetching && <Spinner className="h-3 w-3" />}
              </div>
              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2">
                  Per page
                  <Select
                    className="w-20 py-1"
                    value={limit}
                    onChange={(e) => {
                      setLimit(Number(e.target.value))
                      setPage(1)
                    }}
                  >
                    {[10, 20, 50, 100].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </Select>
                </label>
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  Previous
                </Button>
                <span>
                  Page {data.page} of {Math.max(data.total_pages, 1)}
                </span>
                <Button size="sm" variant="secondary" disabled={page >= data.total_pages} onClick={() => setPage(page + 1)}>
                  Next
                </Button>
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}
