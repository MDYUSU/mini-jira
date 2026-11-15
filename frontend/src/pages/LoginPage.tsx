import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { setToken } from '../lib/auth'
import { DEMO_ACCOUNTS, DEMO_PASSWORD } from '../lib/constants'
import { Button, ErrorBox, Field, Input } from '../components/ui'
import { AuthShell } from './AuthShell'

export function LoginPage() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const login = useMutation({
    mutationFn: () => api.login(email, password),
    onSuccess: (res) => {
      setToken(res.token, res.expires_at)
      qc.clear()
      navigate('/', { replace: true })
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    login.mutate()
  }

  const useDemo = (address: string) => {
    setEmail(address)
    setPassword(DEMO_PASSWORD)
  }

  return (
    <AuthShell title="Log in to your account">
      <form onSubmit={submit} className="space-y-4">
        <Field label="Email">
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoFocus required />
        </Field>
        <Field label="Password">
          <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        </Field>
        <ErrorBox error={login.error} />
        <Button type="submit" className="w-full" disabled={login.isPending}>
          {login.isPending ? 'Logging in...' : 'Log in'}
        </Button>
      </form>
      <p className="mt-4 text-center text-sm text-slate-500">
        No account?{' '}
        <Link to="/register" className="font-medium text-blue-600 hover:underline">
          Create one
        </Link>
      </p>
      <div className="mt-5 rounded-md border border-blue-100 bg-blue-50 p-3 text-xs text-slate-600">
        <p className="mb-1.5 font-semibold text-slate-700">Demo accounts (password: {DEMO_PASSWORD})</p>
        <ul className="space-y-1">
          {DEMO_ACCOUNTS.map((address) => (
            <li key={address}>
              <button type="button" onClick={() => useDemo(address)} className="text-blue-700 hover:underline">
                {address}
              </button>
            </li>
          ))}
        </ul>
      </div>
    </AuthShell>
  )
}
