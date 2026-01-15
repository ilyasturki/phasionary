import { sqliteTable, text, integer } from 'drizzle-orm/sqlite-core'

// Users table - managed by better-auth but we define it here for relationships
export const users = sqliteTable('users', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  email: text('email').notNull().unique(),
  emailVerified: integer('email_verified', { mode: 'boolean' }).notNull().default(false),
  image: text('image'),
  createdAt: text('created_at').notNull(),
  updatedAt: text('updated_at').notNull(),
})

// Sessions table - managed by better-auth
export const sessions = sqliteTable('sessions', {
  id: text('id').primaryKey(),
  expiresAt: text('expires_at').notNull(),
  token: text('token').notNull().unique(),
  createdAt: text('created_at').notNull(),
  updatedAt: text('updated_at').notNull(),
  ipAddress: text('ip_address'),
  userAgent: text('user_agent'),
  userId: text('user_id')
    .notNull()
    .references(() => users.id, { onDelete: 'cascade' }),
})

// Accounts table - managed by better-auth
export const accounts = sqliteTable('accounts', {
  id: text('id').primaryKey(),
  accountId: text('account_id').notNull(),
  providerId: text('provider_id').notNull(),
  userId: text('user_id')
    .notNull()
    .references(() => users.id, { onDelete: 'cascade' }),
  accessToken: text('access_token'),
  refreshToken: text('refresh_token'),
  idToken: text('id_token'),
  accessTokenExpiresAt: text('access_token_expires_at'),
  refreshTokenExpiresAt: text('refresh_token_expires_at'),
  scope: text('scope'),
  password: text('password'),
  createdAt: text('created_at').notNull(),
  updatedAt: text('updated_at').notNull(),
})

// Verification table - managed by better-auth
export const verifications = sqliteTable('verifications', {
  id: text('id').primaryKey(),
  identifier: text('identifier').notNull(),
  value: text('value').notNull(),
  expiresAt: text('expires_at').notNull(),
  createdAt: text('created_at'),
  updatedAt: text('updated_at'),
})

// Projects table
export const projects = sqliteTable('projects', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  description: text('description'),
  userId: text('user_id')
    .notNull()
    .references(() => users.id, { onDelete: 'cascade' }),
  createdAt: text('created_at').notNull(),
  updatedAt: text('updated_at').notNull(),
})

// Categories table
export const categories = sqliteTable('categories', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  projectId: text('project_id')
    .notNull()
    .references(() => projects.id, { onDelete: 'cascade' }),
  createdAt: text('created_at').notNull(),
})

// Tasks table
export const tasks = sqliteTable('tasks', {
  id: text('id').primaryKey(),
  title: text('title').notNull(),
  description: text('description'),
  deadline: text('deadline'),
  timeEstimateValue: integer('time_estimate_value'),
  timeEstimateUnit: text('time_estimate_unit'), // minutes, hours, days
  status: text('status').notNull().default('todo'), // todo, in_progress, completed, cancelled
  section: text('section').notNull().default('current'), // current, future, past
  priority: text('priority'), // high, medium, low, or null
  notes: text('notes'),
  completionDate: text('completion_date'),
  projectId: text('project_id')
    .notNull()
    .references(() => projects.id, { onDelete: 'cascade' }),
  categoryId: text('category_id')
    .notNull()
    .references(() => categories.id),
  createdAt: text('created_at').notNull(),
  updatedAt: text('updated_at').notNull(),
})

// Type exports
export type User = typeof users.$inferSelect
export type Session = typeof sessions.$inferSelect
export type Account = typeof accounts.$inferSelect
export type Project = typeof projects.$inferSelect
export type NewProject = typeof projects.$inferInsert
export type Category = typeof categories.$inferSelect
export type NewCategory = typeof categories.$inferInsert
export type Task = typeof tasks.$inferSelect
export type NewTask = typeof tasks.$inferInsert
