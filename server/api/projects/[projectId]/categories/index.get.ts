import { getDb, schema } from '../../../../db'
import { eq, and, sql } from 'drizzle-orm'

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  if (!projectId) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

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

  // Get categories with task counts
  const categories = await db
    .select({
      id: schema.categories.id,
      name: schema.categories.name,
      projectId: schema.categories.projectId,
      createdAt: schema.categories.createdAt,
      taskCount: sql<number>`(
        SELECT COUNT(*) FROM ${schema.tasks}
        WHERE ${schema.tasks.categoryId} = ${schema.categories.id}
      )`.as('task_count'),
    })
    .from(schema.categories)
    .where(eq(schema.categories.projectId, projectId))
    .orderBy(schema.categories.createdAt)

  return { data: categories }
})
