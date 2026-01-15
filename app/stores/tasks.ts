import { defineStore } from 'pinia'

export type TaskStatus = 'todo' | 'in_progress' | 'completed' | 'cancelled'
export type TaskSection = 'current' | 'future' | 'past'
export type TaskPriority = 'high' | 'medium' | 'low' | null
export type TimeEstimateUnit = 'minutes' | 'hours' | 'days'

export interface Task {
  id: string
  title: string
  description: string | null
  deadline: string | null
  timeEstimateValue: number | null
  timeEstimateUnit: TimeEstimateUnit | null
  status: TaskStatus
  section: TaskSection
  priority: TaskPriority
  notes: string | null
  completionDate: string | null
  projectId: string
  categoryId: string
  createdAt: string
  updatedAt: string
}

export interface TaskFilters {
  section?: TaskSection
  status?: TaskStatus
  categoryId?: string
  priority?: TaskPriority
}

export const useTasksStore = defineStore('tasks', () => {
  const tasks = ref<Task[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<TaskFilters>({ section: 'current' })

  const projectsStore = useProjectsStore()

  // Sort tasks by priority -> deadline -> time estimate -> title
  const sortedTasks = computed(() => {
    const priorityOrder: Record<string, number> = {
      high: 0,
      medium: 1,
      low: 2,
    }

    return [...tasks.value].sort((a, b) => {
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
  })

  // Group tasks by category
  const tasksByCategory = computed(() => {
    const grouped = new Map<string, Task[]>()
    for (const task of sortedTasks.value) {
      const list = grouped.get(task.categoryId) || []
      list.push(task)
      grouped.set(task.categoryId, list)
    }
    return grouped
  })

  function toMinutes(value: number, unit: TimeEstimateUnit | null): number {
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

  async function fetchTasks(filterOverrides?: TaskFilters) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const queryParams = new URLSearchParams()
      const activeFilters = { ...filters.value, ...filterOverrides }

      if (activeFilters.section) queryParams.set('section', activeFilters.section)
      if (activeFilters.status) queryParams.set('status', activeFilters.status)
      if (activeFilters.categoryId) queryParams.set('category', activeFilters.categoryId)
      if (activeFilters.priority) queryParams.set('priority', activeFilters.priority)

      const url = `/api/projects/${projectId}/tasks?${queryParams.toString()}`
      const response = await $fetch<{ data: Task[] }>(url)
      tasks.value = response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch tasks'
    } finally {
      isLoading.value = false
    }
  }

  async function createTask(data: Partial<Task>) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Task }>(
        `/api/projects/${projectId}/tasks`,
        {
          method: 'POST',
          body: data,
        }
      )
      tasks.value.push(response.data)
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to create task'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function updateTask(id: string, data: Partial<Task>) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Task }>(
        `/api/projects/${projectId}/tasks/${id}`,
        {
          method: 'PUT',
          body: data,
        }
      )
      const index = tasks.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tasks.value[index] = response.data
      }
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to update task'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function updateTaskStatus(id: string, status: TaskStatus) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    try {
      const response = await $fetch<{ data: Task }>(
        `/api/projects/${projectId}/tasks/${id}/status`,
        {
          method: 'PATCH',
          body: { status },
        }
      )
      const index = tasks.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tasks.value[index] = response.data
      }
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to update task status'
      throw e
    }
  }

  async function updateTaskSection(id: string, section: TaskSection) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    try {
      const response = await $fetch<{ data: Task }>(
        `/api/projects/${projectId}/tasks/${id}/section`,
        {
          method: 'PATCH',
          body: { section },
        }
      )
      const index = tasks.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tasks.value[index] = response.data
      }
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to update task section'
      throw e
    }
  }

  async function deleteTask(id: string) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      await $fetch(`/api/projects/${projectId}/tasks/${id}`, { method: 'DELETE' })
      const index = tasks.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tasks.value.splice(index, 1)
      }
    } catch (e: any) {
      error.value = e.message || 'Failed to delete task'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  function setFilters(newFilters: TaskFilters) {
    filters.value = { ...filters.value, ...newFilters }
  }

  function clearTasks() {
    tasks.value = []
  }

  return {
    tasks,
    sortedTasks,
    tasksByCategory,
    isLoading,
    error,
    filters,
    fetchTasks,
    createTask,
    updateTask,
    updateTaskStatus,
    updateTaskSection,
    deleteTask,
    setFilters,
    clearTasks,
  }
})
