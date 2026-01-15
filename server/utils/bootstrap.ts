import { getDb, schema } from '../db'

const DEFAULT_CATEGORIES = [
  'Feature',
  'Fix',
  'Ergonomy',
  'Documentation',
  'Research',
]

export async function bootstrapUser(userId: string) {
  const db = getDb()
  const now = new Date().toISOString()

  // Check if user already has projects
  const existingProjects = await db.query.projects.findMany({
    where: (projects, { eq }) => eq(projects.userId, userId),
    limit: 1,
  })

  if (existingProjects.length > 0) {
    return // User already bootstrapped
  }

  // Create default project
  const projectId = crypto.randomUUID()
  await db.insert(schema.projects).values({
    id: projectId,
    name: 'My Project',
    description: 'Your first project',
    userId,
    createdAt: now,
    updatedAt: now,
  })

  // Create default categories
  for (const categoryName of DEFAULT_CATEGORIES) {
    await db.insert(schema.categories).values({
      id: crypto.randomUUID(),
      name: categoryName,
      projectId,
      createdAt: now,
    })
  }

  return projectId
}
