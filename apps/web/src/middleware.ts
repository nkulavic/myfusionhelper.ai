import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Routes that don't require authentication
const publicRoutes = ['/', '/login', '/register', '/forgot-password', '/reset-password', '/terms', '/privacy', '/eula']

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Check if the route is public
  const isPublicRoute = publicRoutes.some(
    (route) => pathname === route || pathname.startsWith(route + '/')
  )

  // Check for auth indicator cookie (set by auth-client when tokens are stored)
  const isAuthenticated = request.cookies.has('mfh_authenticated')

  // Authenticated users: rewrite / to dashboard (URL stays /), redirect away from login/register
  if (isAuthenticated) {
    if (pathname === '/') {
      return NextResponse.rewrite(new URL('/dashboard', request.url))
    }
    if (pathname === '/login' || pathname === '/register') {
      return NextResponse.redirect(new URL('/', request.url))
    }
  }

  // Allow public routes
  if (isPublicRoute) {
    return NextResponse.next()
  }

  // If not authenticated and trying to access protected route, redirect to login
  if (!isAuthenticated) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('callbackUrl', pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|.*\\..*).*)'],
}
