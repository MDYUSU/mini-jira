const KEY = 'minijira.orgs'

export interface RememberedOrg {
  id: number
  name: string
}

function storageKey(userId: number): string {
  return `${KEY}.${userId}`
}

export function getRememberedOrgs(userId: number): RememberedOrg[] {
  try {
    const raw = localStorage.getItem(storageKey(userId))
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function rememberOrg(userId: number, org: RememberedOrg): RememberedOrg[] {
  const list = getRememberedOrgs(userId).filter((o) => o.id !== org.id)
  const next = [...list, org].sort((a, b) => a.id - b.id)
  localStorage.setItem(storageKey(userId), JSON.stringify(next))
  window.dispatchEvent(new Event('minijira:orgs'))
  return next
}

export function forgetOrg(userId: number, orgId: number): RememberedOrg[] {
  const next = getRememberedOrgs(userId).filter((o) => o.id !== orgId)
  localStorage.setItem(storageKey(userId), JSON.stringify(next))
  window.dispatchEvent(new Event('minijira:orgs'))
  return next
}
