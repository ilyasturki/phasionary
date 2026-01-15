import { getDb, schema } from '../../../../db'
import { eq, and, ne } from 'drizzle-orm'
import * as v from 'valibot'

const UpdateCategorySchema = v.object({
  name: v.pipe(v.string(), v.minLength(1), v.maxLength(100)),
})

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  const id = getRouterParam(event, 'id')
  if (!projectId || !id) {
    throw createError({ statusCode: 400, message: 'Project ID and Category ID required' })
  }

  const body = await readBody(event)
  const result = v.safeParse(UpdateCategorySchema, body)

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

  // Verify category exists and belongs to project
  const existing = await db.query.categories.findFirst({
    where: and(
      eq(schema.categories.id, id),
      eq(schema.categories.projectId, projectId)
    ),
  })

  if (!existing) {
    throw createError({ statusCode: 404, message: 'Category not found' })
  }

  // Check for duplicate name (case-insensitive)
  if (name !== existing.name) {
    const duplicate = await db.query.categories.findFirst({
      where: and(
        eq(schema.categories.projectId, projectId),
        eq(schema.categories.name, name),
        ne(schema.categories.id, id)
      ),
    })

    if (duplicate) {
      throw createError({
        statusCode: 400,
        message: 'A category with this name already exists',
      })
    }
  }

  await db
    .update(schema.categories)
    .set({ name })
    .where(eq(schema.categories.id, id))

  const category = await db.query.categories.findFirst({
    where: eq(schema.categories.id, id),
  })

  return { data: category }
})
