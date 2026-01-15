import { getDb, schema } from '../../../../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const UpdateSectionSchema = v.object({
  section: v.picklist(['current', 'future', 'past']),
})

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

  const body = await readBody(event)
  const result = v.safeParse(UpdateSectionSchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const { section } = result.output
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

  // Enforce: tasks can only be in Past if completed/cancelled
  if (section === 'past' && existing.status !== 'completed' && existing.status !== 'cancelled') {
    throw createError({
      statusCode: 400,
      message: 'Only completed or cancelled tasks can be moved to Past',
    })
  }

  const now = new Date().toISOString()

  await db
    .update(schema.tasks)
    .set({
      section,
      updatedAt: now,
    })
    .where(eq(schema.tasks.id, id))

  const task = await db.query.tasks.findFirst({
    where: eq(schema.tasks.id, id),
  })

  return { data: task }
})
