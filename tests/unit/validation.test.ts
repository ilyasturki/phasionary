import { describe, it, expect } from 'vitest'
import * as v from 'valibot'

// Recreate schemas for testing
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

const CreateProjectSchema = v.object({
  name: v.pipe(v.string(), v.minLength(1), v.maxLength(100)),
  description: v.optional(v.string()),
})

describe('Task Validation', () => {
  it('validates a minimal valid task', () => {
    const result = v.safeParse(CreateTaskSchema, {
      title: 'Test task',
      categoryId: 'cat-123',
    })
    expect(result.success).toBe(true)
  })

  it('rejects empty title', () => {
    const result = v.safeParse(CreateTaskSchema, {
      title: '',
      categoryId: 'cat-123',
    })
    expect(result.success).toBe(false)
  })

  it('rejects title over 200 characters', () => {
    const result = v.safeParse(CreateTaskSchema, {
      title: 'a'.repeat(201),
      categoryId: 'cat-123',
    })
    expect(result.success).toBe(false)
  })

  it('accepts valid priority values', () => {
    for (const priority of ['high', 'medium', 'low', null]) {
      const result = v.safeParse(CreateTaskSchema, {
        title: 'Test',
        categoryId: 'cat-123',
        priority,
      })
      expect(result.success).toBe(true)
    }
  })

  it('rejects invalid priority values', () => {
    const result = v.safeParse(CreateTaskSchema, {
      title: 'Test',
      categoryId: 'cat-123',
      priority: 'urgent',
    })
    expect(result.success).toBe(false)
  })

  it('accepts valid time estimate units', () => {
    for (const unit of ['minutes', 'hours', 'days']) {
      const result = v.safeParse(CreateTaskSchema, {
        title: 'Test',
        categoryId: 'cat-123',
        timeEstimateValue: 5,
        timeEstimateUnit: unit,
      })
      expect(result.success).toBe(true)
    }
  })

  it('rejects invalid time estimate unit', () => {
    const result = v.safeParse(CreateTaskSchema, {
      title: 'Test',
      categoryId: 'cat-123',
      timeEstimateValue: 5,
      timeEstimateUnit: 'weeks',
    })
    expect(result.success).toBe(false)
  })

  it('accepts valid section values', () => {
    for (const section of ['current', 'future', 'past']) {
      const result = v.safeParse(CreateTaskSchema, {
        title: 'Test',
        categoryId: 'cat-123',
        section,
      })
      expect(result.success).toBe(true)
    }
  })
})

describe('Project Validation', () => {
  it('validates a valid project', () => {
    const result = v.safeParse(CreateProjectSchema, {
      name: 'My Project',
    })
    expect(result.success).toBe(true)
  })

  it('rejects empty name', () => {
    const result = v.safeParse(CreateProjectSchema, {
      name: '',
    })
    expect(result.success).toBe(false)
  })

  it('rejects name over 100 characters', () => {
    const result = v.safeParse(CreateProjectSchema, {
      name: 'a'.repeat(101),
    })
    expect(result.success).toBe(false)
  })

  it('accepts optional description', () => {
    const result = v.safeParse(CreateProjectSchema, {
      name: 'My Project',
      description: 'A great project',
    })
    expect(result.success).toBe(true)
  })
})
