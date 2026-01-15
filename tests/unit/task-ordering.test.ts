import { describe, it, expect } from 'vitest'

// Task ordering logic extracted for testing
type Task = {
  id: string
  title: string
  priority: 'high' | 'medium' | 'low' | null
  deadline: string | null
  timeEstimateValue: number | null
  timeEstimateUnit: 'minutes' | 'hours' | 'days' | null
}

function toMinutes(value: number, unit: 'minutes' | 'hours' | 'days' | null): number {
  switch (unit) {
    case 'minutes':
      return value
    case 'hours':
      return value * 60
    case 'days':
      return value * 60 * 24
    default:
      return value
  }
}

function sortTasks(tasks: Task[]): Task[] {
  const priorityOrder: Record<string, number> = {
    high: 0,
    medium: 1,
    low: 2,
  }

  return [...tasks].sort((a, b) => {
    // Priority comparison (high -> medium -> low -> null)
    const aPriority = a.priority ? priorityOrder[a.priority] : 3
    const bPriority = b.priority ? priorityOrder[b.priority] : 3
    if (aPriority !== bPriority) return aPriority - bPriority

    // Deadline comparison (earliest first, null last)
    if (a.deadline && !b.deadline) return -1
    if (!a.deadline && b.deadline) return 1
    if (a.deadline && b.deadline) {
      const dateCompare = new Date(a.deadline).getTime() - new Date(b.deadline).getTime()
      if (dateCompare !== 0) return dateCompare
    }

    // Time estimate comparison (shortest first, null last)
    const aMinutes = a.timeEstimateValue ? toMinutes(a.timeEstimateValue, a.timeEstimateUnit) : Infinity
    const bMinutes = b.timeEstimateValue ? toMinutes(b.timeEstimateValue, b.timeEstimateUnit) : Infinity
    if (aMinutes !== bMinutes) return aMinutes - bMinutes

    // Title comparison (alphabetical, case-insensitive)
    return a.title.toLowerCase().localeCompare(b.title.toLowerCase())
  })
}

describe('Task Ordering', () => {
  it('sorts by priority (high -> medium -> low -> null)', () => {
    const tasks: Task[] = [
      { id: '1', title: 'A', priority: null, deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '2', title: 'B', priority: 'low', deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '3', title: 'C', priority: 'high', deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '4', title: 'D', priority: 'medium', deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
    ]

    const sorted = sortTasks(tasks)
    expect(sorted.map(t => t.priority)).toEqual(['high', 'medium', 'low', null])
  })

  it('sorts by deadline (earliest first, null last) within same priority', () => {
    const tasks: Task[] = [
      { id: '1', title: 'A', priority: 'high', deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '2', title: 'B', priority: 'high', deadline: '2024-01-15', timeEstimateValue: null, timeEstimateUnit: null },
      { id: '3', title: 'C', priority: 'high', deadline: '2024-01-10', timeEstimateValue: null, timeEstimateUnit: null },
    ]

    const sorted = sortTasks(tasks)
    expect(sorted.map(t => t.title)).toEqual(['C', 'B', 'A'])
  })

  it('sorts by time estimate (shortest first, null last) within same priority and deadline', () => {
    const tasks: Task[] = [
      { id: '1', title: 'A', priority: 'high', deadline: '2024-01-15', timeEstimateValue: null, timeEstimateUnit: null },
      { id: '2', title: 'B', priority: 'high', deadline: '2024-01-15', timeEstimateValue: 2, timeEstimateUnit: 'hours' },
      { id: '3', title: 'C', priority: 'high', deadline: '2024-01-15', timeEstimateValue: 30, timeEstimateUnit: 'minutes' },
    ]

    const sorted = sortTasks(tasks)
    expect(sorted.map(t => t.title)).toEqual(['C', 'B', 'A'])
  })

  it('normalizes time estimates to minutes for comparison', () => {
    const tasks: Task[] = [
      { id: '1', title: 'A', priority: null, deadline: null, timeEstimateValue: 2, timeEstimateUnit: 'days' },
      { id: '2', title: 'B', priority: null, deadline: null, timeEstimateValue: 3, timeEstimateUnit: 'hours' },
      { id: '3', title: 'C', priority: null, deadline: null, timeEstimateValue: 90, timeEstimateUnit: 'minutes' },
    ]

    const sorted = sortTasks(tasks)
    // 90 min < 3 hours (180 min) < 2 days (2880 min)
    expect(sorted.map(t => t.title)).toEqual(['C', 'B', 'A'])
  })

  it('sorts alphabetically by title as final tiebreaker (case-insensitive)', () => {
    const tasks: Task[] = [
      { id: '1', title: 'Zebra', priority: null, deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '2', title: 'apple', priority: null, deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
      { id: '3', title: 'Banana', priority: null, deadline: null, timeEstimateValue: null, timeEstimateUnit: null },
    ]

    const sorted = sortTasks(tasks)
    expect(sorted.map(t => t.title)).toEqual(['apple', 'Banana', 'Zebra'])
  })
})
