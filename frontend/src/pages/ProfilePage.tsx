import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../api/client'
import { useInvalidate } from '../api/queries'
import { useCurrentUser } from '../components/AuthGate'
import { Avatar, Button, Card, ErrorBox, Field, Input, Textarea } from '../components/ui'
import { formatDate } from '../lib/format'

export function ProfilePage() {
  const user = useCurrentUser()
  const invalidate = useInvalidate()
  const [name, setName] = useState(user.name)
  const [bio, setBio] = useState(user.profile?.bio ?? '')
  const [saved, setSaved] = useState(false)

  const save = useMutation({
    mutationFn: async () => {
      const tasks: Promise<unknown>[] = []
      if (name.trim() !== user.name) tasks.push(api.updateUser(user.id, { name: name.trim() }))
      if (bio !== (user.profile?.bio ?? '')) tasks.push(api.updateProfile(user.id, { bio }))
      await Promise.all(tasks)
    },
    onSuccess: async () => {
      await Promise.all([invalidate.me(), invalidate.users()])
      setSaved(true)
    },
  })

  return (
    <div className="mx-auto max-w-2xl p-6">
      <h1 className="mb-6 text-xl font-bold text-slate-800">Profile</h1>
      <Card className="p-6">
        <div className="mb-6 flex items-center gap-4">
          <Avatar name={user.name} id={user.id} size="lg" />
          <div>
            <p className="text-lg font-semibold text-slate-800">{user.name}</p>
            <p className="text-sm text-slate-500">{user.email}</p>
            <p className="text-xs text-slate-400">Member since {formatDate(user.created_at)}</p>
          </div>
        </div>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            setSaved(false)
            save.mutate()
          }}
          className="space-y-4"
        >
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} required />
          </Field>
          <Field label="Email" hint="Email cannot be changed here">
            <Input value={user.email} disabled />
          </Field>
          <Field label="Bio">
            <Textarea rows={4} value={bio} onChange={(e) => setBio(e.target.value)} placeholder="Tell your team about yourself" />
          </Field>
          <ErrorBox error={save.error} />
          <div className="flex items-center gap-3">
            <Button type="submit" disabled={save.isPending || !name.trim()}>
              {save.isPending ? 'Saving...' : 'Save changes'}
            </Button>
            {saved && !save.isPending && <span className="text-sm text-emerald-600">Saved</span>}
          </div>
        </form>
      </Card>
    </div>
  )
}
