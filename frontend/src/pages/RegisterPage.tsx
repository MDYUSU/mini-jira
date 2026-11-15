import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { setToken } from '../lib/auth'
import { Button, ErrorBox, Field, Input } from '../components/ui'
import { AuthShell } from './AuthShell'

export function RegisterPage() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [localError, setLocalError] = useState('')

  const register = useMutation({
    mutationFn: () => api.register(email, name, password),
    onSuccess: (res) => {
      setToken(res.token, res.expires_at)
      qc.clear()
      navigate('/', { replace: true })
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (password.length < 8 || password.length > 72) {
      setLocalError('Password must be 8 to 72 characters')
      return
    }
    setLocalError('')
    register.mutate()
  }

  return (
    <AuthShell title="Create your account">
      <form onSubmit={submit} className="space-y-4">
        <Field label="Name">
          <Input value={name} onChange={(e) => setName(e.target.value)} autoFocus required />
        </Field>
        <Field label="Email">
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </Field>
        <Field label="Password" hint="8 to 72 characters">
          <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        </Field>
        <ErrorBox error={localError || register.error} />
        <Button type="submit" className="w-full" disabled={register.isPending}>
          {register.isPending ? 'Creating account...' : 'Create account'}
        </Button>
      </form>
      <p className="mt-4 text-center text-sm text-slate-500">
        Already have an account?{' '}
        <Link to="/login" className="font-medium text-blue-600 hover:underline">
          Log in
        </Link>
      </p>
    </AuthShell>
  )
}
