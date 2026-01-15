import { getDb, schema } from '../../../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const CreateCategorySchema = v.object({
  name: v.pipe(v.string(), v.minLength(1), v.maxLength(100)),
})

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  if (!projectId) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

  const body = await readBody(event)
  const result = v.safeParse(CreateCategorySchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const { name } = result.output
  const db = getDb()

  // Verify project exists and belongs to user
  const project = await db.query.projects.findFirst({
    where: and(
      eq(schema.projects.id, projectId),
      eq(schema.projects.userId, session.user.id)
    ),
  })

  if (!project) {
    throw createError({ statusCode: 404, message: 'Project not found' })
  }

  // Check for duplicate name (case-insensitive)
  const existing = await db.query.categories.findFirst({
    where: and(
      eq(schema.categories.projectId, projectId),
      eq(schema.categories.name, name)
    ),
  })

  if (existing) {
    throw createError({
      statusCode: 400,
      message: 'A category with this name already exists',
    })
  }

  const now = new Date().toISOString()
  const categoryId = crypto.randomUUID()

  await db.insert(schema.categories).values({
    id: categoryId,
    name,
    projectId,
    createdAt: now,
  })

  const category = await db.query.categories.findFirst({
    where: eq(schema.categories.id, categoryId),
  })

  return { data: category }
})
