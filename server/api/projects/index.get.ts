import { getDb, schema } from '../../db'
import { eq } from 'drizzle-orm'

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const db = getDb()
  const projects = await db
    .select()
    .from(schema.projects)
    .where(eq(schema.projects.userId, session.user.id))
    .orderBy(schema.projects.createdAt)

  return { data: projects }
})
