import { createContext, useContext, type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useMe } from '../api/queries'
import type { UserWithProfile } from '../api/types'
import { isLoggedIn } from '../lib/auth'
import { ErrorBox, Loading } from './ui'

const CurrentUserContext = createContext<UserWithProfile | null>(null)

export function useCurrentUser(): UserWithProfile {
  const user = useContext(CurrentUserContext)
  if (!user) throw new Error('useCurrentUser must be used inside AuthGate')
  return user
}

function Gate({ children }: { children: ReactNode }) {
  const me = useMe()
  if (me.isPending) return <Loading text="Loading your account..." />
  if (me.isError) {
    return (
      <div className="mx-auto mt-20 max-w-md">
        <ErrorBox error={me.error} />
      </div>
    )
  }
  return <CurrentUserContext.Provider value={me.data}>{children}</CurrentUserContext.Provider>
}

export function AuthGate({ children }: { children: ReactNode }) {
  if (!isLoggedIn()) return <Navigate to="/login" replace />
  return <Gate>{children}</Gate>
}
