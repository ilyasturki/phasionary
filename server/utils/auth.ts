import { betterAuth } from 'better-auth'
import { drizzleAdapter } from 'better-auth/adapters/drizzle'
import Database from 'better-sqlite3'
import { drizzle } from 'drizzle-orm/better-sqlite3'
import * as schema from '../db/schema'
import { bootstrapUser } from './bootstrap'

let _auth: ReturnType<typeof betterAuth> | null = null

export function getAuth() {
  if (!_auth) {
    const config = useRuntimeConfig()
    const sqlite = new Database(config.databasePath)
    sqlite.pragma('journal_mode = WAL')
    const db = drizzle(sqlite, { schema })

    _auth = betterAuth({
      database: drizzleAdapter(db, {
        provider: 'sqlite',
        schema: {
          user: schema.users,
          session: schema.sessions,
          account: schema.accounts,
          verification: schema.verifications,
        },
      }),
      emailAndPassword: {
        enabled: true,
        minPasswordLength: 8,
      },
      session: {
        cookieCache: {
          enabled: true,
          maxAge: 5 * 60, // 5 minutes
        },
      },
      trustedOrigins: config.public.appUrl ? [config.public.appUrl] : [],
      user: {
        additionalFields: {},
      },
      databaseHooks: {
        user: {
          create: {
            after: async (user) => {
              // Bootstrap user with default project and categories
              await bootstrapUser(user.id)
            },
          },
        },
      },
    })
  }
  return _auth
}

export type Auth = ReturnType<typeof getAuth>
