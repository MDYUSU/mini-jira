import type { Label, Priority, ProjectMember } from '../api/types'
import { PRIORITIES, PRIORITY_LABEL } from '../lib/constants'
import { Input, Select } from './ui'

export interface FilterState {
  priority: Priority | ''
  assignee: string
  label: string
  search: string
  hideSubtasks: boolean
}

export const EMPTY_FILTERS: FilterState = { priority: '', assignee: '', label: '', search: '', hideSubtasks: false }

interface Props {
  value: FilterState
  onChange: (next: FilterState) => void
  members: ProjectMember[]
  labels: Label[]
}

export function TaskFilters({ value, onChange, members, labels }: Props) {
  const set = <K extends keyof FilterState>(key: K, v: FilterState[K]) => onChange({ ...value, [key]: v })
  const active =
    value.priority !== '' || value.assignee !== '' || value.label !== '' || value.search !== '' || value.hideSubtasks

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        className="w-56"
        placeholder="Search tasks"
        value={value.search}
        onChange={(e) => set('search', e.target.value)}
      />
      <Select className="w-36" value={value.priority} onChange={(e) => set('priority', e.target.value as Priority | '')}>
        <option value="">All priorities</option>
        {PRIORITIES.map((p) => (
          <option key={p} value={p}>
            {PRIORITY_LABEL[p]}
          </option>
        ))}
      </Select>
      <Select className="w-40" value={value.assignee} onChange={(e) => set('assignee', e.target.value)}>
        <option value="">All assignees</option>
        <option value="none">Unassigned</option>
        {members.map((m) => (
          <option key={m.user_id} value={m.user_id}>
            {m.name}
          </option>
        ))}
      </Select>
      <Select className="w-36" value={value.label} onChange={(e) => set('label', e.target.value)}>
        <option value="">All labels</option>
        {labels.map((l) => (
          <option key={l.id} value={l.name}>
            {l.name}
          </option>
        ))}
      </Select>
      <label className="flex items-center gap-2 px-2 text-sm text-slate-600">
        <input
          type="checkbox"
          checked={value.hideSubtasks}
          onChange={(e) => set('hideSubtasks', e.target.checked)}
          className="h-4 w-4 rounded border-slate-300"
        />
        Hide subtasks
      </label>
      {active && (
        <button onClick={() => onChange(EMPTY_FILTERS)} className="text-sm text-blue-600 hover:underline">
          Clear filters
        </button>
      )}
    </div>
  )
}
