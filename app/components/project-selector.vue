<script setup lang="ts">
const projectsStore = useProjectsStore()
const { projects, activeProject, isLoading } = storeToRefs(projectsStore)

const isOpen = ref(false)
const showCreateModal = ref(false)
const newProjectName = ref('')
const createError = ref('')

async function selectProject(id: string) {
  projectsStore.setActiveProject(id)
  isOpen.value = false
}

async function createProject() {
  if (!newProjectName.value.trim()) return
  createError.value = ''

  try {
    await projectsStore.createProject({ name: newProjectName.value.trim() })
    newProjectName.value = ''
    showCreateModal.value = false
  } catch (e: any) {
    createError.value = e.data?.message || 'Failed to create project'
  }
}

function openCreateModal() {
  isOpen.value = false
  showCreateModal.value = true
}
</script>

<template>
  <div class="relative">
    <!-- Trigger button -->
    <button
      @click="isOpen = !isOpen"
      class="flex items-center gap-2 border border-border bg-bg-surface px-3 py-2 text-sm hover:bg-bg-muted"
    >
      <span v-if="activeProject" class="text-text-primary">
        {{ activeProject.name }}
      </span>
      <span v-else class="text-text-muted">Select project</span>
      <Icon name="lucide:chevron-down" class="h-4 w-4 text-text-muted" />
    </button>

    <!-- Dropdown -->
    <div
      v-if="isOpen"
      class="absolute left-0 top-full z-50 mt-1 min-w-[200px] border border-border bg-bg-elevated shadow-lg"
    >
      <div class="max-h-[300px] overflow-y-auto">
        <button
          v-for="project in projects"
          :key="project.id"
          @click="selectProject(project.id)"
          class="flex w-full items-center px-3 py-2 text-left text-sm hover:bg-bg-muted"
          :class="{ 'bg-bg-muted': project.id === activeProject?.id }"
        >
          <span class="text-text-primary">{{ project.name }}</span>
        </button>
      </div>
      <div class="border-t border-border">
        <button
          @click="openCreateModal"
          class="flex w-full items-center gap-2 px-3 py-2 text-sm text-accent hover:bg-bg-muted"
        >
          <Icon name="lucide:plus" class="h-4 w-4" />
          New Project
        </button>
      </div>
    </div>

    <!-- Click outside to close -->
    <div
      v-if="isOpen"
      class="fixed inset-0 z-40"
      @click="isOpen = false"
    />

    <!-- Create Modal -->
    <Teleport to="body">
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="showCreateModal = false"
      >
        <div class="w-full max-w-md border border-border bg-bg-elevated p-6">
          <h2 class="mb-4 text-lg font-semibold text-text-primary">
            New Project
          </h2>

          <form @submit.prevent="createProject">
            <div v-if="createError" class="mb-4 border border-error bg-error/10 p-3 text-sm text-error">
              {{ createError }}
            </div>

            <div class="mb-4">
              <label for="projectName" class="mb-1 block text-sm text-text-secondary">
                Project name
              </label>
              <input
                id="projectName"
                v-model="newProjectName"
                type="text"
                required
                maxlength="100"
                class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
                placeholder="My New Project"
                autofocus
              />
            </div>

            <div class="flex justify-end gap-2">
              <button
                type="button"
                @click="showCreateModal = false"
                class="border border-border px-4 py-2 text-sm text-text-secondary hover:bg-bg-muted"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover"
              >
                Create
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
