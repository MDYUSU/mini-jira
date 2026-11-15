import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../api/client'
import { useInvalidate, useLabels, type Permissions } from '../../api/queries'
import type { Label, ProjectDetail } from '../../api/types'
import { useConfirm } from '../../components/Confirm'
import { Button, Card, EmptyState, ErrorBox, Input, LabelChip, Loading } from '../../components/ui'

const PRESETS = ['#2563eb', '#16a34a', '#dc2626', '#d97706', '#7c3aed', '#0891b2', '#db2777', '#475569']

function LabelRow({ label, projectId, canEdit }: { label: Label; projectId: number; canEdit: boolean }) {
  const confirm = useConfirm()
  const invalidate = useInvalidate()
  const remove = useMutation({
    mutationFn: () => api.deleteLabel(label.id),
    onSuccess: () => invalidate.projectAndTasks(projectId),
  })

  return (
    <li className="flex items-center gap-3 px-4 py-3">
      <span className="h-4 w-4 rounded" style={{ backgroundColor: label.color }} />
      <LabelChip label={label} />
      <span className="flex-1 font-mono text-xs text-slate-400">{label.color}</span>
      {remove.error && <ErrorBox error={remove.error} />}
      {canEdit && (
        <Button
          size="sm"
          variant="ghost"
          className="text-red-600 hover:bg-red-50"
          disabled={remove.isPending}
          onClick={async () => {
            const ok = await confirm({
              title: 'Delete label',
              message: `Delete the label "${label.name}"? It will be removed from every task that uses it.`,
            })
            if (ok) remove.mutate()
          }}
        >
          Delete
        </Button>
      )}
    </li>
  )
}

export function LabelsTab({ project, perms }: { project: ProjectDetail; perms: Permissions }) {
  const invalidate = useInvalidate()
  const labels = useLabels(project.id)
  const [name, setName] = useState('')
  const [color, setColor] = useState(PRESETS[0])

  const create = useMutation({
    mutationFn: () => api.createLabel(project.id, name.trim(), color),
    onSuccess: async () => {
      setName('')
      await invalidate.project(project.id)
    },
  })

  return (
    <div className="max-w-3xl">
      {perms.canEdit && (
        <Card className="mb-4 p-4">
          <h3 className="mb-3 text-sm font-semibold text-slate-700">Create a label</h3>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              create.mutate()
            }}
            className="flex flex-wrap items-center gap-3"
          >
            <Input className="w-56" placeholder="Label name" value={name} onChange={(e) => setName(e.target.value)} />
            <input
              type="color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              className="h-9 w-12 cursor-pointer rounded border border-slate-300 bg-white p-1"
              aria-label="Label color"
            />
            <div className="flex gap-1">
              {PRESETS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setColor(c)}
                  className={`h-6 w-6 rounded-full ${color === c ? 'ring-2 ring-slate-800 ring-offset-1' : ''}`}
                  style={{ backgroundColor: c }}
                  aria-label={c}
                />
              ))}
            </div>
            {name.trim() && <LabelChip label={{ id: 0, name: name.trim().toLowerCase(), color }} />}
            <Button type="submit" disabled={!name.trim() || create.isPending}>
              Create
            </Button>
          </form>
          <ErrorBox error={create.error} className="mt-3" />
        </Card>
      )}

      {labels.isPending ? (
        <Loading />
      ) : labels.isError ? (
        <ErrorBox error={labels.error} />
      ) : labels.data.length === 0 ? (
        <EmptyState title="No labels yet" />
      ) : (
        <Card>
          <ul className="divide-y divide-slate-100">
            {labels.data.map((l) => (
              <LabelRow key={l.id} label={l} projectId={project.id} canEdit={perms.canEdit} />
            ))}
          </ul>
        </Card>
      )}
    </div>
  )
}
