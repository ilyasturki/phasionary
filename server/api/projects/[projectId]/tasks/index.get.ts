import { getDb, schema } from '../../../../db'
import { eq, and } from 'drizzle-orm'

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  if (!projectId) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

  const query = getQuery(event)
  const section = query.section as string | undefined
  const status = query.status as string | undefined
  const category = query.category as string | undefined
  const priority = query.priority as string | undefined

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

  // Build query with filters
  let tasks = await db.query.tasks.findMany({
    where: eq(schema.tasks.projectId, projectId),
  })

  // Apply filters
  if (section) {
    tasks = tasks.filter((t) => t.section === section)
  }
  if (status) {
    tasks = tasks.filter((t) => t.status === status)
  }
  if (category) {
    tasks = tasks.filter((t) => t.categoryId === category)
  }
  if (priority) {
    tasks = tasks.filter((t) => t.priority === priority)
  }

  return { data: tasks }
})
