const TOKEN_KEY = 'minijira.token'
const EXPIRES_KEY = 'minijira.expires_at'

export function getToken(): string | null {
  const token = localStorage.getItem(TOKEN_KEY)
  const expiresAt = localStorage.getItem(EXPIRES_KEY)
  if (!token) return null
  if (expiresAt && new Date(expiresAt).getTime() <= Date.now()) {
    clearToken()
    return null
  }
  return token
}

export function setToken(token: string, expiresAt: string): void {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(EXPIRES_KEY, expiresAt)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(EXPIRES_KEY)
}

export function isLoggedIn(): boolean {
  return getToken() !== null
}
