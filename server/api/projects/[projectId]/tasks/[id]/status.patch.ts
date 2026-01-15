import { getDb, schema } from '../../../../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const UpdateStatusSchema = v.object({
  status: v.picklist(['todo', 'in_progress', 'completed', 'cancelled']),
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
  const result = v.safeParse(UpdateStatusSchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const { status } = result.output
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

  const now = new Date().toISOString()
  let completionDate = existing.completionDate
  let section = existing.section

  // Handle completion date
  if (status === 'completed' && existing.status !== 'completed') {
    completionDate = now
  } else if (status !== 'completed' && existing.status === 'completed') {
    completionDate = null
  }

  // Auto-move to past section if completed/cancelled (can be overridden by user setting later)
  if ((status === 'completed' || status === 'cancelled') && section !== 'past') {
    section = 'past'
  }

  // When reopening a task from completed/cancelled, move to current
  if (status !== 'completed' && status !== 'cancelled' &&
      (existing.status === 'completed' || existing.status === 'cancelled')) {
    section = 'current'
  }

  await db
    .update(schema.tasks)
    .set({
      status,
      completionDate,
      section,
      updatedAt: now,
    })
    .where(eq(schema.tasks.id, id))

  const task = await db.query.tasks.findFirst({
    where: eq(schema.tasks.id, id),
  })

  return { data: task }
})
