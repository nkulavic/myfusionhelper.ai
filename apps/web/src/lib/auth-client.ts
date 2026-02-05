import { AUTH_CONFIG } from './auth'

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(AUTH_CONFIG.tokenKey)
}

export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(AUTH_CONFIG.refreshTokenKey)
}

export function setTokens(accessToken: string, refreshToken?: string) {
  if (typeof window === 'undefined') return
  localStorage.setItem(AUTH_CONFIG.tokenKey, accessToken)
  if (refreshToken) {
    localStorage.setItem(AUTH_CONFIG.refreshTokenKey, refreshToken)
  }
  // Set a simple cookie so middleware can detect auth state
  document.cookie = 'mfh_authenticated=1; path=/; max-age=2592000; SameSite=Lax'
}

export function clearTokens() {
  if (typeof window === 'undefined') return
  localStorage.removeItem(AUTH_CONFIG.tokenKey)
  localStorage.removeItem(AUTH_CONFIG.refreshTokenKey)
  // Clear the auth indicator cookie
  document.cookie = 'mfh_authenticated=; path=/; max-age=0'
}
