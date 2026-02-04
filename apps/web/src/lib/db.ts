import { sql } from '@vercel/postgres'
import { drizzle } from 'drizzle-orm/vercel-postgres'

// For Vercel Postgres
export const db = drizzle(sql)

// Database schema will be defined here
// When ready for AWS migration, we can add sync hooks
