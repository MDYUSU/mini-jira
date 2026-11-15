import { clearToken, getToken } from './auth'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type Method = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

type Query = Record<string, string | number | boolean | undefined | null>

interface RequestOptions {
  body?: unknown
  query?: Query
  skipAuthRedirect?: boolean
}

function buildUrl(path: string, query?: Query): string {
  const url = `/api${path}`
  if (!query) return url
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '' || value === false) continue
    params.set(key, String(value))
  }
  const qs = params.toString()
  return qs ? `${url}?${qs}` : url
}

export async function request<T>(method: Method, path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {}
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`
  if (options.body !== undefined) headers['Content-Type'] = 'application/json'

  let res: Response
  try {
    res = await fetch(buildUrl(path, options.query), {
      method,
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
    })
  } catch {
    throw new ApiError(0, 'Cannot reach the server. Is the backend running on port 8080?')
  }

  if (res.status === 401 && !options.skipAuthRedirect) {
    clearToken()
    if (window.location.pathname !== '/login') window.location.assign('/login')
    throw new ApiError(401, 'Your session has expired. Please log in again.')
  }

  if (res.status === 204) return undefined as T

  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = null
    }
  }

  if (!res.ok) {
    const message =
      data && typeof data === 'object' && 'error' in data && typeof data.error === 'string'
        ? data.error
        : res.status === 502 || res.status === 504
          ? 'Cannot reach the server. Is the backend running on port 8080?'
          : `Request failed with status ${res.status}`
    throw new ApiError(res.status, message)
  }

  return data as T
}

export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message
  return 'Something went wrong'
}
