import type { APIResponse } from '@myfusionhelper/types'
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@/lib/auth-client'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.myfusionhelper.ai'

class APIError extends Error {
  constructor(
    public statusCode: number,
    public code: string,
    message: string
  ) {
    super(message)
    this.name = 'APIError'
  }
}

// snake_case → camelCase transformation for API responses
function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, c) => c.toUpperCase())
}

// camelCase → snake_case transformation for API requests
function camelToSnake(str: string): string {
  return str.replace(/[A-Z]/g, (c) => `_${c.toLowerCase()}`)
}

function transformKeys(obj: unknown, transformFn: (key: string) => string): unknown {
  if (obj === null || obj === undefined) return obj
  if (Array.isArray(obj)) return obj.map((item) => transformKeys(item, transformFn))
  if (typeof obj === 'object' && obj !== null) {
    const result: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(obj)) {
      result[transformFn(key)] = transformKeys(value, transformFn)
    }
    return result
  }
  return obj
}

function toCamelCase<T>(data: unknown): T {
  return transformKeys(data, snakeToCamel) as T
}

function toSnakeCase(data: unknown): unknown {
  return transformKeys(data, camelToSnake)
}

// Token refresh state to prevent concurrent refresh attempts
let isRefreshing = false
let refreshPromise: Promise<boolean> | null = null

async function attemptTokenRefresh(): Promise<boolean> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const refreshToken = getRefreshToken()
      if (!refreshToken) return false

      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!response.ok) return false

      const data = await response.json()
      if (data.data?.token) {
        setTokens(data.data.token, data.data.refresh_token)
        return true
      }
      return false
    } catch {
      return false
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  isRetry = false
): Promise<APIResponse<T>> {
  const token = getAccessToken()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  })

  // On 401, try to refresh the token and retry once
  if (response.status === 401 && !isRetry && !path.includes('/auth/')) {
    const refreshed = await attemptTokenRefresh()
    if (refreshed) {
      return request<T>(path, options, true)
    }
    // Refresh failed — clear auth and let the error propagate
    clearTokens()
  }

  if (!response.ok) {
    let errorData: { error?: { code?: string; message?: string } } = {}
    try {
      errorData = await response.json()
    } catch {
      // response may not be JSON
    }
    throw new APIError(
      response.status,
      errorData.error?.code || 'UNKNOWN_ERROR',
      errorData.error?.message || `Request failed (${response.status})`
    )
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return { success: true, data: undefined as T }
  }

  const raw = await response.json()

  // Backend returns snake_case — transform to camelCase for frontend
  const data = toCamelCase<APIResponse<T>>(raw)

  return data
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: 'GET' }),

  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(toSnakeCase(body)) : undefined,
    }),

  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(toSnakeCase(body)) : undefined,
    }),

  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'PATCH',
      body: body ? JSON.stringify(toSnakeCase(body)) : undefined,
    }),

  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}

export { APIError }
