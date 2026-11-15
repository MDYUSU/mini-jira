import { useEffect, useState } from 'react'
import { getRememberedOrgs, type RememberedOrg } from '../lib/orgs'

export function useRememberedOrgs(userId: number): RememberedOrg[] {
  const [orgs, setOrgs] = useState(() => getRememberedOrgs(userId))

  useEffect(() => {
    const update = () => setOrgs(getRememberedOrgs(userId))
    update()
    window.addEventListener('minijira:orgs', update)
    window.addEventListener('storage', update)
    return () => {
      window.removeEventListener('minijira:orgs', update)
      window.removeEventListener('storage', update)
    }
  }, [userId])

  return orgs
}
