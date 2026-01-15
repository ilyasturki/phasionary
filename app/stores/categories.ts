import { defineStore } from 'pinia'

export interface Category {
  id: string
  name: string
  projectId: string
  createdAt: string
  taskCount?: number
}

export const useCategoriesStore = defineStore('categories', () => {
  const categories = ref<Category[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const projectsStore = useProjectsStore()

  async function fetchCategories() {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Category[] }>(
        `/api/projects/${projectId}/categories`
      )
      categories.value = response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch categories'
    } finally {
      isLoading.value = false
    }
  }

  async function createCategory(name: string) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Category }>(
        `/api/projects/${projectId}/categories`,
        {
          method: 'POST',
          body: { name },
        }
      )
      categories.value.push(response.data)
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to create category'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function updateCategory(id: string, name: string) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Category }>(
        `/api/projects/${projectId}/categories/${id}`,
        {
          method: 'PUT',
          body: { name },
        }
      )
      const index = categories.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        categories.value[index] = response.data
      }
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to update category'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function deleteCategory(id: string, reassignTo?: string) {
    const projectId = projectsStore.activeProjectId
    if (!projectId) return

    isLoading.value = true
    error.value = null
    try {
      await $fetch(`/api/projects/${projectId}/categories/${id}`, {
        method: 'DELETE',
        body: reassignTo ? { reassignTo } : undefined,
      })
      const index = categories.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        categories.value.splice(index, 1)
      }
    } catch (e: any) {
      error.value = e.message || 'Failed to delete category'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  function clearCategories() {
    categories.value = []
  }

  return {
    categories,
    isLoading,
    error,
    fetchCategories,
    createCategory,
    updateCategory,
    deleteCategory,
    clearCategories,
  }
})
