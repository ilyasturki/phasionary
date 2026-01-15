import { getDb, schema } from '../../../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const UpdateTaskSchema = v.object({
  title: v.optional(v.pipe(v.string(), v.minLength(1), v.maxLength(200))),
  description: v.optional(v.nullable(v.string())),
  deadline: v.optional(v.nullable(v.string())),
  timeEstimateValue: v.optional(v.nullable(v.number())),
  timeEstimateUnit: v.optional(v.nullable(v.picklist(['minutes', 'hours', 'days']))),
  categoryId: v.optional(v.string()),
  priority: v.optional(v.nullable(v.picklist(['high', 'medium', 'low']))),
  notes: v.optional(v.nullable(v.string())),
  status: v.optional(v.picklist(['todo', 'in_progress', 'completed', 'cancelled'])),
  section: v.optional(v.picklist(['current', 'future', 'past'])),
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
  const result = v.safeParse(UpdateTaskSchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const data = result.output
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

  // If category is being changed, verify it exists in project
  if (data.categoryId && data.categoryId !== existing.categoryId) {
    const category = await db.query.categories.findFirst({
      where: and(
        eq(schema.categories.id, data.categoryId),
        eq(schema.categories.projectId, projectId)
      ),
    })

    if (!category) {
      throw createError({ statusCode: 400, message: 'Category not found' })
    }
  }

  const now = new Date().toISOString()
  let completionDate = existing.completionDate

  // Handle status changes for completion date
  if (data.status !== undefined) {
    if (data.status === 'completed' && existing.status !== 'completed') {
      completionDate = now
    } else if (data.status !== 'completed' && existing.status === 'completed') {
      completionDate = null
    }
  }

  // Build update object
  const updates: Record<string, unknown> = { updatedAt: now }

  if (data.title !== undefined) updates.title = data.title
  if (data.description !== undefined) updates.description = data.description
  if (data.deadline !== undefined) updates.deadline = data.deadline
  if (data.timeEstimateValue !== undefined) updates.timeEstimateValue = data.timeEstimateValue
  if (data.timeEstimateUnit !== undefined) updates.timeEstimateUnit = data.timeEstimateUnit
  if (data.categoryId !== undefined) updates.categoryId = data.categoryId
  if (data.priority !== undefined) updates.priority = data.priority
  if (data.notes !== undefined) updates.notes = data.notes
  if (data.status !== undefined) updates.status = data.status
  if (data.section !== undefined) updates.section = data.section
  updates.completionDate = completionDate

  await db
    .update(schema.tasks)
    .set(updates)
    .where(eq(schema.tasks.id, id))

  const task = await db.query.tasks.findFirst({
    where: eq(schema.tasks.id, id),
  })

  return { data: task }
})
