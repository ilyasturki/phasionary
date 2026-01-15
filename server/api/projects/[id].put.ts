import { getDb, schema } from '../../db'
import { eq, and, ne } from 'drizzle-orm'
import * as v from 'valibot'

const UpdateProjectSchema = v.object({
  name: v.optional(v.pipe(v.string(), v.minLength(1), v.maxLength(100))),
  description: v.optional(v.nullable(v.string())),
})

export default defineEventHandler(async (event) => {
  const session = event.context.session
  if (!session?.user?.id) {
    throw createError({ statusCode: 401, message: 'Unauthorized' })
  }

  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, message: 'Project ID required' })
  }

  const body = await readBody(event)
  const result = v.safeParse(UpdateProjectSchema, body)

  if (!result.success) {
    throw createError({
      statusCode: 400,
      message: 'Validation error',
      data: result.issues,
    })
  }

  const { name, description } = result.output
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

  // Check for duplicate name (case-insensitive) if name is being changed
  if (name && name !== existing.name) {
    const duplicate = await db.query.projects.findFirst({
      where: and(
        eq(schema.projects.userId, session.user.id),
        eq(schema.projects.name, name),
        ne(schema.projects.id, id)
      ),
    })

    if (duplicate) {
      throw createError({
        statusCode: 400,
        message: 'A project with this name already exists',
      })
    }
  }

  const now = new Date().toISOString()
  await db
    .update(schema.projects)
    .set({
      ...(name !== undefined && { name }),
      ...(description !== undefined && { description }),
      updatedAt: now,
    })
    .where(eq(schema.projects.id, id))

  const project = await db.query.projects.findFirst({
    where: eq(schema.projects.id, id),
  })

  return { data: project }
})
