import { getDb, schema } from '../../../../db'
import { eq, and } from 'drizzle-orm'
import * as v from 'valibot'

const CreateTaskSchema = v.object({
  title: v.pipe(v.string(), v.minLength(1), v.maxLength(200)),
  description: v.optional(v.nullable(v.string())),
  deadline: v.optional(v.nullable(v.string())),
  timeEstimateValue: v.optional(v.nullable(v.number())),
  timeEstimateUnit: v.optional(v.nullable(v.picklist(['minutes', 'hours', 'days']))),
  categoryId: v.string(),
  priority: v.optional(v.nullable(v.picklist(['high', 'medium', 'low']))),
  notes: v.optional(v.nullable(v.string())),
  section: v.optional(v.picklist(['current', 'future', 'past'])),
})

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const projectId = getRouterParam(event, 'projectId')
  if (!projectId) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

  const body = await readBody(event)
  const result = v.safeParse(CreateTaskSchema, body)

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

  // Verify category exists and belongs to project
  const category = await db.query.categories.findFirst({
    where: and(
      eq(schema.categories.id, data.categoryId),
      eq(schema.categories.projectId, projectId)
    ),
  })

  if (!category) {
    throw createError({ statusCode: 400, message: 'Category not found' })
  }

  const now = new Date().toISOString()
  const taskId = crypto.randomUUID()

  await db.insert(schema.tasks).values({
    id: taskId,
    title: data.title,
    description: data.description || null,
    deadline: data.deadline || null,
    timeEstimateValue: data.timeEstimateValue || null,
    timeEstimateUnit: data.timeEstimateUnit || null,
    status: 'todo',
    section: data.section || 'current',
    priority: data.priority || null,
    notes: data.notes || null,
    completionDate: null,
    projectId,
    categoryId: data.categoryId,
    createdAt: now,
    updatedAt: now,
  })

  const task = await db.query.tasks.findFirst({
    where: eq(schema.tasks.id, taskId),
  })

  return { data: task }
})
