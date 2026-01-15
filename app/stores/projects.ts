import { defineStore } from 'pinia'

export interface Project {
  id: string
  name: string
  description: string | null
  userId: string
  createdAt: string
  updatedAt: string
}

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<Project[]>([])
  const activeProjectId = ref<string | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const activeProject = computed(() =>
    projects.value.find((p) => p.id === activeProjectId.value) || null
  )

  async function fetchProjects() {
    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Project[] }>('/api/projects')
      projects.value = response.data
      // Set active project to first one if none selected
      if (!activeProjectId.value && projects.value.length > 0) {
        activeProjectId.value = projects.value[0].id
      }
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch projects'
    } finally {
      isLoading.value = false
    }
  }

  async function createProject(data: { name: string; description?: string }) {
    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Project }>('/api/projects', {
        method: 'POST',
        body: data,
      })
      projects.value.push(response.data)
      activeProjectId.value = response.data.id
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to create project'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function updateProject(
    id: string,
    data: { name?: string; description?: string }
  ) {
    isLoading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data: Project }>(`/api/projects/${id}`, {
        method: 'PUT',
        body: data,
      })
      const index = projects.value.findIndex((p) => p.id === id)
      if (index !== -1) {
        projects.value[index] = response.data
      }
      return response.data
    } catch (e: any) {
      error.value = e.message || 'Failed to update project'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function deleteProject(id: string) {
    isLoading.value = true
    error.value = null
    try {
      await $fetch(`/api/projects/${id}`, { method: 'DELETE' })
      const index = projects.value.findIndex((p) => p.id === id)
      if (index !== -1) {
        projects.value.splice(index, 1)
      }
      // If deleted project was active, switch to another
      if (activeProjectId.value === id) {
        activeProjectId.value = projects.value[0]?.id || null
      }
    } catch (e: any) {
      error.value = e.message || 'Failed to delete project'
      throw e
    } finally {
      isLoading.value = false
    }
  }

  function setActiveProject(id: string) {
    activeProjectId.value = id
  }

  return {
    projects,
    activeProjectId,
    activeProject,
    isLoading,
    error,
    fetchProjects,
    createProject,
    updateProject,
    deleteProject,
    setActiveProject,
  }
})
