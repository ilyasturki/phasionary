<script setup lang="ts">
definePageMeta({
  middleware: ['auth'],
})

const auth = useAuth()
const { data: session } = await auth.useSession(useFetch)
const router = useRouter()

const projectsStore = useProjectsStore()
const categoriesStore = useCategoriesStore()
const tasksStore = useTasksStore()

const { activeProject, isLoading: projectsLoading } = storeToRefs(projectsStore)

// Fetch projects on mount
await projectsStore.fetchProjects()

// Watch for project changes and fetch categories
watch(
  () => projectsStore.activeProjectId,
  async (projectId) => {
    if (projectId) {
      await categoriesStore.fetchCategories()
      await tasksStore.fetchTasks()
    } else {
      categoriesStore.clearCategories()
      tasksStore.clearTasks()
    }
  },
  { immediate: true }
)

async function handleSignOut() {
  await auth.signOut()
  await router.push('/login')
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <h1 class="text-2xl font-semibold text-text-primary">Dashboard</h1>
        <ProjectSelector />
      </div>
      <button
        @click="handleSignOut"
        class="border border-border px-3 py-1.5 text-sm text-text-secondary hover:bg-bg-muted hover:text-text-primary"
      >
        Sign out
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="projectsLoading" class="py-8 text-center text-text-muted">
      Loading projects...
    </div>

    <!-- Main content -->
    <div v-else-if="activeProject" class="grid grid-cols-1 gap-6 lg:grid-cols-4">
      <!-- Sidebar -->
      <aside class="lg:col-span-1">
        <CategoryManager />
      </aside>

      <!-- Main area -->
      <main class="lg:col-span-3">
        <div class="border border-border bg-bg-surface p-4">
          <h2 class="mb-2 text-lg font-semibold text-text-primary">
            {{ activeProject.name }}
          </h2>
          <p v-if="activeProject.description" class="mb-4 text-sm text-text-secondary">
            {{ activeProject.description }}
          </p>
          <p class="text-sm text-text-muted">
            Tasks will appear here. Coming in Iteration 1c.
          </p>
        </div>
      </main>
    </div>

    <!-- No projects -->
    <div v-else class="py-8 text-center">
      <p class="text-text-muted">No projects found. Create one to get started!</p>
    </div>
  </div>
</template>
