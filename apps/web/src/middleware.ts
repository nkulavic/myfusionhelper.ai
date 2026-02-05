import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Routes that don't require authentication
const publicRoutes = ['/', '/login', '/register']

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Check if the route is public
  const isPublicRoute = publicRoutes.some(
    (route) => pathname === route || pathname.startsWith(route + '/')
  )

  // Allow public routes
  if (isPublicRoute) {
    return NextResponse.next()
  }

  // Check for Cognito auth cookies (Amplify v6 stores tokens as CognitoIdentityServiceProvider.* cookies)
  const hasCognitoCookie = request.cookies.getAll().some(
    (cookie) => cookie.name.startsWith('CognitoIdentityServiceProvider.')
  )

  // If no Cognito session and trying to access protected route, redirect to login
  if (!hasCognitoCookie) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('callbackUrl', pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|.*\\..*).*)'],
}
