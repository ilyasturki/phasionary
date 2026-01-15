import { getDb, schema } from '../../../../db'
import { eq, and } from 'drizzle-orm'

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  const id = getRouterParam(event, 'id')
  if (!projectId || !id) {
    throw createError({ statusCode: 400, message: 'Project ID and Task ID required' })
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

  // Verify task exists and belongs to project
  const existing = await db.query.tasks.findFirst({
    where: and(
      eq(schema.tasks.id, id),
      eq(schema.tasks.projectId, projectId)
    ),
  })

  if (!existing) {
    throw createError({ statusCode: 404, message: 'Task not found' })
  }

  // Delete task
  await db.delete(schema.tasks).where(eq(schema.tasks.id, id))

  return { data: { success: true } }
})
