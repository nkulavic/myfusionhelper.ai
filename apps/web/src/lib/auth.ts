import { betterAuth } from 'better-auth'

export const auth = betterAuth({
  // Use in-memory for prototype until database is configured
  // Swap to drizzle + Neon Postgres when DATABASE_URL is available
  ...(process.env.DATABASE_URL
    ? {
        database: {
          type: 'postgres' as const,
          url: process.env.DATABASE_URL,
        },
      }
    : {}),
  emailAndPassword: {
    enabled: true,
    autoSignIn: true,
  },
  socialProviders: {
    ...(process.env.GOOGLE_CLIENT_ID && process.env.GOOGLE_CLIENT_SECRET
      ? {
          google: {
            clientId: process.env.GOOGLE_CLIENT_ID,
            clientSecret: process.env.GOOGLE_CLIENT_SECRET,
          },
        }
      : {}),
  },
  session: {
    expiresIn: 60 * 60 * 24 * 7, // 7 days
    updateAge: 60 * 60 * 24, // 1 day
  },
})
