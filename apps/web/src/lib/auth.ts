// Auth configuration
// Authentication is handled entirely by the Go backend which talks to Cognito.
// The frontend just calls backend auth endpoints and manages tokens client-side.

export const AUTH_CONFIG = {
  tokenKey: 'mfh_access_token',
  refreshTokenKey: 'mfh_refresh_token',
}
