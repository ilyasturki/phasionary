import { getDb, schema } from '../../db'
import { eq, and } from 'drizzle-orm'

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

  const db = getDb()

  // Verify project exists and belongs to user
  const existing = await db.query.projects.findFirst({
    where: and(
      eq(schema.projects.id, id),
      eq(schema.projects.userId, session.user.id)
    ),
  })

  if (!existing) {
    throw createError({ statusCode: 404, message: 'Project not found' })
  }

  // Delete project (cascades to categories and tasks)
  await db.delete(schema.projects).where(eq(schema.projects.id, id))

  return { data: { success: true } }
})
