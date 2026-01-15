import { getDb, schema } from '../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const CreateProjectSchema = v.object({
  name: v.pipe(v.string(), v.minLength(1), v.maxLength(100)),
  description: v.optional(v.string()),
})

const DEFAULT_CATEGORIES = [
  'Feature',
  'Fix',
  'Ergonomy',
  'Documentation',
  'Research',
]

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const body = await readBody(event)
  const result = v.safeParse(CreateProjectSchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const { name, description } = result.output
  const db = getDb()

  // Check for duplicate name (case-insensitive)
  const existing = await db.query.projects.findFirst({
    where: and(
      eq(schema.projects.userId, session.user.id),
      eq(schema.projects.name, name)
    ),
  })

  if (existing) {
    throw createError({
      statusCode: 400,
      message: 'A project with this name already exists',
    })
  }

  const now = new Date().toISOString()
  const projectId = crypto.randomUUID()

  // Create project
  await db.insert(schema.projects).values({
    id: projectId,
    name,
    description: description || null,
    userId: session.user.id,
    createdAt: now,
    updatedAt: now,
  })

  // Create default categories
  for (const categoryName of DEFAULT_CATEGORIES) {
    await db.insert(schema.categories).values({
      id: crypto.randomUUID(),
      name: categoryName,
      projectId,
      createdAt: now,
    })
  }

  const project = await db.query.projects.findFirst({
    where: eq(schema.projects.id, projectId),
  })

  return { data: project }
})
