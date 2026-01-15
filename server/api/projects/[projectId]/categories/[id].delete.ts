import { getDb, schema } from '../../../../db'
import { eq, and, count } from 'drizzle-orm'
import * as v from 'valibot'

const DeleteCategorySchema = v.object({
  reassignTo: v.optional(v.string()),
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

  const body = await readBody(event).catch(() => ({}))
  const result = v.safeParse(DeleteCategorySchema, body)
  const reassignTo = result.success ? result.output.reassignTo : undefined

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

  // Count categories in project
  const [categoryCount] = await db
    .select({ count: count() })
    .from(schema.categories)
    .where(eq(schema.categories.projectId, projectId))

  if ((categoryCount?.count ?? 0) <= 1) {
    throw createError({
      statusCode: 400,
      message: 'Cannot delete the last category. Each project must have at least one category.',
    })
  }

  // Check if category has tasks
  const [taskCount] = await db
    .select({ count: count() })
    .from(schema.tasks)
    .where(eq(schema.tasks.categoryId, id))

  if ((taskCount?.count ?? 0) > 0) {
    if (!reassignTo) {
      throw createError({
        statusCode: 400,
        message: 'Category has tasks. Provide a reassignTo category ID.',
      })
    }

    // Verify reassignment category exists
    const reassignCategory = await db.query.categories.findFirst({
      where: and(
        eq(schema.categories.id, reassignTo),
        eq(schema.categories.projectId, projectId)
      ),
    })

    if (!reassignCategory) {
      throw createError({
        statusCode: 400,
        message: 'Reassignment category not found',
      })
    }

    // Reassign tasks
    await db
      .update(schema.tasks)
      .set({ categoryId: reassignTo })
      .where(eq(schema.tasks.categoryId, id))
  }

  // Delete category
  await db.delete(schema.categories).where(eq(schema.categories.id, id))

  return { data: { success: true } }
})
