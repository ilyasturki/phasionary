<script setup lang="ts">
import type { Task, TaskStatus } from '~/stores/tasks'

const props = defineProps<{
  task: Task
}>()

const emit = defineEmits<{
  edit: [task: Task]
  delete: [task: Task]
  statusChange: [taskId: string, status: TaskStatus]
}>()

const categoriesStore = useCategoriesStore()
const { categories } = storeToRefs(categoriesStore)

const categoryName = computed(() => {
  const cat = categories.value.find((c) => c.id === props.task.categoryId)
  return cat?.name || 'Unknown'
})

const statusOptions: { value: TaskStatus; label: string }[] = [
  { value: 'todo', label: 'To Do' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const priorityColors: Record<string, string> = {
  high: 'text-priority-high',
  medium: 'text-priority-medium',
  low: 'text-priority-low',
}

const statusColors: Record<string, string> = {
  todo: 'bg-text-muted',
  in_progress: 'bg-info',
  completed: 'bg-success',
  cancelled: 'bg-text-muted',
}

const isOverdue = computed(() => {
  if (!props.task.deadline) return false
  if (props.task.status === 'completed' || props.task.status === 'cancelled') return false
  return new Date(props.task.deadline) < new Date()
})

function formatDeadline(deadline: string | null): string {
  if (!deadline) return ''
  const date = new Date(deadline)
  const today = new Date()
  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)

  // Check if it's today
  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  }
  // Check if it's tomorrow
  if (date.toDateString() === tomorrow.toDateString()) {
    return 'Tomorrow'
  }
  // Check if it's in the past
  if (date < today) {
    const daysAgo = Math.floor((today.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
    return `${daysAgo}d overdue`
  }

  return date.toLocaleDateString()
}

function formatEstimate(value: number | null, unit: string | null): string {
  if (!value || !unit) return ''
  return `${value} ${unit}`
}

function onStatusChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('statusChange', props.task.id, target.value as TaskStatus)
}
</script>

<template>
  <div
    class="group border border-border bg-bg-surface p-3 transition-colors hover:bg-bg-muted"
    :class="{ 'opacity-60 line-through': task.status === 'cancelled' }"
  >
    <div class="flex items-start justify-between gap-2">
      <!-- Title and metadata -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <!-- Priority indicator -->
          <span
            v-if="task.priority"
            :class="priorityColors[task.priority]"
            class="text-xs font-medium uppercase"
          >
            {{ task.priority }}
          </span>
          <h3 class="text-sm font-medium text-text-primary truncate">
            {{ task.title }}
          </h3>
        </div>

        <!-- Description preview -->
        <p
          v-if="task.description"
          class="mt-1 text-xs text-text-secondary truncate"
        >
          {{ task.description }}
        </p>

        <!-- Metadata row -->
        <div class="mt-2 flex flex-wrap items-center gap-3 text-xs text-text-muted">
          <!-- Deadline -->
          <span
            v-if="task.deadline"
            class="flex items-center gap-1"
            :class="{ 'text-error font-medium': isOverdue }"
          >
            <Icon name="lucide:calendar" class="h-3 w-3" />
            {{ formatDeadline(task.deadline) }}
          </span>

          <!-- Time estimate -->
          <span v-if="task.timeEstimateValue" class="flex items-center gap-1">
            <Icon name="lucide:clock" class="h-3 w-3" />
            {{ formatEstimate(task.timeEstimateValue, task.timeEstimateUnit) }}
          </span>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center gap-2">
        <!-- Status dropdown -->
        <select
          :value="task.status"
          @change="onStatusChange"
          class="border border-border bg-bg-elevated px-2 py-1 text-xs text-text-primary focus:border-border-focus focus:outline-none"
        >
          <option
            v-for="opt in statusOptions"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </option>
        </select>

        <!-- Edit/Delete buttons -->
        <div class="flex gap-1 opacity-0 group-hover:opacity-100">
          <button
            @click="emit('edit', task)"
            class="p-1 text-text-muted hover:text-text-primary"
            title="Edit"
          >
            <Icon name="lucide:pencil" class="h-4 w-4" />
          </button>
          <button
            @click="emit('delete', task)"
            class="p-1 text-text-muted hover:text-error"
            title="Delete"
          >
            <Icon name="lucide:trash-2" class="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
