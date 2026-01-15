<script setup lang="ts">
import type { Task, TaskSection, TaskStatus } from '~/stores/tasks'

const tasksStore = useTasksStore()
const categoriesStore = useCategoriesStore()
const { tasksByCategory, isLoading, filters } = storeToRefs(tasksStore)
const { categories } = storeToRefs(categoriesStore)

const activeSection = ref<TaskSection>('current')
const showTaskForm = ref(false)
const editingTask = ref<Task | null>(null)
const deletingTask = ref<Task | null>(null)

const sections: { value: TaskSection; label: string }[] = [
  { value: 'current', label: 'Current' },
  { value: 'future', label: 'Future' },
  { value: 'past', label: 'Past' },
]

// Fetch tasks when section changes
watch(
  activeSection,
  async (section) => {
    tasksStore.setFilters({ section })
    await tasksStore.fetchTasks()
  },
  { immediate: true }
)

function getCategoryName(categoryId: string): string {
  const cat = categories.value.find((c) => c.id === categoryId)
  return cat?.name || 'Unknown'
}

function openCreateForm() {
  editingTask.value = null
  showTaskForm.value = true
}

function openEditForm(task: Task) {
  editingTask.value = task
  showTaskForm.value = true
}

function openDeleteConfirm(task: Task) {
  deletingTask.value = task
}

async function handleSaveTask(data: Partial<Task>) {
  try {
    if (editingTask.value) {
      await tasksStore.updateTask(editingTask.value.id, data)
    } else {
      await tasksStore.createTask(data)
    }
    showTaskForm.value = false
    editingTask.value = null
    // Refresh tasks and categories (for task counts)
    await tasksStore.fetchTasks()
    await categoriesStore.fetchCategories()
  } catch (e) {
    console.error('Failed to save task:', e)
  }
}

async function handleStatusChange(taskId: string, status: TaskStatus) {
  try {
    await tasksStore.updateTaskStatus(taskId, status)
    // Refresh tasks and categories after status change (may move to different section)
    await tasksStore.fetchTasks()
    await categoriesStore.fetchCategories()
  } catch (e) {
    console.error('Failed to update status:', e)
  }
}

async function handleDeleteTask() {
  if (!deletingTask.value) return
  try {
    await tasksStore.deleteTask(deletingTask.value.id)
    deletingTask.value = null
    // Refresh categories for updated counts
    await categoriesStore.fetchCategories()
  } catch (e) {
    console.error('Failed to delete task:', e)
  }
}

// Get sorted category IDs that have tasks
const categoriesWithTasks = computed(() => {
  const ids: string[] = []
  for (const [categoryId] of tasksByCategory.value) {
    ids.push(categoryId)
  }
  // Sort by category order
  return ids.sort((a, b) => {
    const aIndex = categories.value.findIndex((c) => c.id === a)
    const bIndex = categories.value.findIndex((c) => c.id === b)
    return aIndex - bIndex
  })
})
</script>

<template>
  <div>
    <!-- Section tabs and create button -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex gap-1">
        <button
          v-for="sec in sections"
          :key="sec.value"
          @click="activeSection = sec.value"
          class="px-3 py-1.5 text-sm"
          :class="
            activeSection === sec.value
              ? 'bg-accent text-white'
              : 'bg-bg-surface text-text-secondary hover:bg-bg-muted'
          "
        >
          {{ sec.label }}
        </button>
      </div>

      <button
        v-if="activeSection !== 'past'"
        @click="openCreateForm"
        class="flex items-center gap-1 bg-accent px-3 py-1.5 text-sm font-medium text-white hover:bg-accent-hover"
      >
        <Icon name="lucide:plus" class="h-4 w-4" />
        New Task
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="py-8 text-center text-text-muted">
      Loading tasks...
    </div>

    <!-- Task list grouped by category -->
    <div v-else-if="categoriesWithTasks.length > 0" class="space-y-6">
      <div
        v-for="categoryId in categoriesWithTasks"
        :key="categoryId"
      >
        <h3 class="mb-2 text-sm font-medium text-text-secondary">
          {{ getCategoryName(categoryId) }}
          <span class="text-text-muted">({{ tasksByCategory.get(categoryId)?.length || 0 }})</span>
        </h3>
        <div class="space-y-2">
          <TaskCard
            v-for="task in tasksByCategory.get(categoryId)"
            :key="task.id"
            :task="task"
            @edit="openEditForm"
            @delete="openDeleteConfirm"
            @status-change="handleStatusChange"
          />
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="py-8 text-center">
      <p class="text-text-muted">
        <template v-if="activeSection === 'current'">
          No tasks in current. Create one to get started!
        </template>
        <template v-else-if="activeSection === 'future'">
          No tasks planned for later.
        </template>
        <template v-else>
          No completed or cancelled tasks yet.
        </template>
      </p>
    </div>

    <!-- Task form modal -->
    <TaskForm
      :is-open="showTaskForm"
      :task="editingTask"
      @close="showTaskForm = false"
      @save="handleSaveTask"
    />

    <!-- Delete confirmation modal -->
    <Teleport to="body">
      <div
        v-if="deletingTask"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="deletingTask = null"
      >
        <div class="w-full max-w-md border border-border bg-bg-elevated p-6">
          <h2 class="mb-4 text-lg font-semibold text-text-primary">
            Delete Task
          </h2>
          <p class="mb-4 text-sm text-text-secondary">
            Are you sure you want to delete "<span class="text-text-primary">{{ deletingTask.title }}</span>"?
            This action cannot be undone.
          </p>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              @click="deletingTask = null"
              class="border border-border px-4 py-2 text-sm text-text-secondary hover:bg-bg-muted"
            >
              Cancel
            </button>
            <button
              @click="handleDeleteTask"
              class="bg-error px-4 py-2 text-sm font-medium text-white hover:bg-error/80"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
