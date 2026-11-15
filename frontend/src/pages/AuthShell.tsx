import type { ReactNode } from 'react'
import { Logo } from '../components/Logo'

export function AuthShell({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-50 px-4 py-10">
      <div className="mb-6 flex items-center gap-2">
        <Logo />
        <span className="text-xl font-bold text-slate-800">Mini Jira</span>
      </div>
      <div className="w-full max-w-sm rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        <h1 className="mb-5 text-lg font-semibold text-slate-800">{title}</h1>
        {children}
      </div>
    </div>
  )
}
