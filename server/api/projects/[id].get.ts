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
  const project = await db.query.projects.findFirst({
    where: and(
      eq(schema.projects.id, id),
      eq(schema.projects.userId, session.user.id)
    ),
  })

  if (!project) {
    throw createError({ statusCode: 404, message: 'Project not found' })
  }

  return { data: project }
})
