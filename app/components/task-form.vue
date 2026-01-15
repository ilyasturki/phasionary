<script setup lang="ts">
import type { Task, TaskPriority, TaskSection, TimeEstimateUnit } from '~/stores/tasks'

const props = defineProps<{
  task?: Task | null
  isOpen: boolean
}>()

const emit = defineEmits<{
  close: []
  save: [data: Partial<Task>]
}>()

const categoriesStore = useCategoriesStore()
const { categories } = storeToRefs(categoriesStore)

const title = ref('')
const description = ref('')
const categoryId = ref('')
const deadline = ref('')
const timeEstimateValue = ref<number | null>(null)
const timeEstimateUnit = ref<TimeEstimateUnit | null>(null)
const priority = ref<TaskPriority>(null)
const notes = ref('')
const section = ref<TaskSection>('current')

const isEditing = computed(() => !!props.task)

// Reset form when opening/closing or when task changes
watch(
  () => [props.isOpen, props.task],
  () => {
    if (props.isOpen) {
      if (props.task) {
        title.value = props.task.title
        description.value = props.task.description || ''
        categoryId.value = props.task.categoryId
        deadline.value = props.task.deadline ? props.task.deadline.split('T')[0] || '' : ''
        timeEstimateValue.value = props.task.timeEstimateValue
        timeEstimateUnit.value = props.task.timeEstimateUnit
        priority.value = props.task.priority
        notes.value = props.task.notes || ''
        section.value = props.task.section
      } else {
        title.value = ''
        description.value = ''
        categoryId.value = categories.value[0]?.id || ''
        deadline.value = ''
        timeEstimateValue.value = null
        timeEstimateUnit.value = null
        priority.value = null
        notes.value = ''
        section.value = 'current'
      }
    }
  },
  { immediate: true }
)

function handleSubmit() {
  const data: Partial<Task> = {
    title: title.value.trim(),
    description: description.value.trim() || null,
    categoryId: categoryId.value,
    deadline: deadline.value ? new Date(deadline.value).toISOString() : null,
    timeEstimateValue: timeEstimateValue.value,
    timeEstimateUnit: timeEstimateValue.value ? timeEstimateUnit.value : null,
    priority: priority.value,
    notes: notes.value.trim() || null,
    section: section.value,
  }

  emit('save', data)
}

const priorityOptions: { value: TaskPriority; label: string }[] = [
  { value: null, label: 'None' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
]

const unitOptions: { value: TimeEstimateUnit; label: string }[] = [
  { value: 'minutes', label: 'Minutes' },
  { value: 'hours', label: 'Hours' },
  { value: 'days', label: 'Days' },
]

const sectionOptions: { value: TaskSection; label: string }[] = [
  { value: 'current', label: 'Current' },
  { value: 'future', label: 'Future' },
]
</script>

<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="emit('close')"
    >
      <div class="w-full max-w-lg max-h-[90vh] overflow-y-auto border border-border bg-bg-elevated p-6">
        <h2 class="mb-4 text-lg font-semibold text-text-primary">
          {{ isEditing ? 'Edit Task' : 'New Task' }}
        </h2>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <!-- Title -->
          <div>
            <label for="title" class="mb-1 block text-sm text-text-secondary">
              Title *
            </label>
            <input
              id="title"
              v-model="title"
              type="text"
              required
              maxlength="200"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
              placeholder="Task title"
              autofocus
            />
          </div>

          <!-- Description -->
          <div>
            <label for="description" class="mb-1 block text-sm text-text-secondary">
              Description
            </label>
            <textarea
              id="description"
              v-model="description"
              rows="3"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
              placeholder="Task description (optional)"
            />
          </div>

          <!-- Category -->
          <div>
            <label for="category" class="mb-1 block text-sm text-text-secondary">
              Category *
            </label>
            <select
              id="category"
              v-model="categoryId"
              required
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
            >
              <option value="" disabled>Select category</option>
              <option
                v-for="cat in categories"
                :key="cat.id"
                :value="cat.id"
              >
                {{ cat.name }}
              </option>
            </select>
          </div>

          <!-- Section (for new tasks) -->
          <div v-if="!isEditing">
            <label for="section" class="mb-1 block text-sm text-text-secondary">
              Section
            </label>
            <select
              id="section"
              v-model="section"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
            >
              <option
                v-for="opt in sectionOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- Priority -->
          <div>
            <label for="priority" class="mb-1 block text-sm text-text-secondary">
              Priority
            </label>
            <select
              id="priority"
              v-model="priority"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
            >
              <option
                v-for="opt in priorityOptions"
                :key="opt.value ?? 'none'"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- Deadline -->
          <div>
            <label for="deadline" class="mb-1 block text-sm text-text-secondary">
              Deadline
            </label>
            <input
              id="deadline"
              v-model="deadline"
              type="date"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
            />
          </div>

          <!-- Time Estimate -->
          <div>
            <label class="mb-1 block text-sm text-text-secondary">
              Time Estimate
            </label>
            <div class="flex gap-2">
              <input
                v-model.number="timeEstimateValue"
                type="number"
                min="1"
                class="w-24 border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
                placeholder="Value"
              />
              <select
                v-model="timeEstimateUnit"
                class="flex-1 border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
              >
                <option :value="null">Select unit</option>
                <option
                  v-for="opt in unitOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Notes -->
          <div>
            <label for="notes" class="mb-1 block text-sm text-text-secondary">
              Notes
            </label>
            <textarea
              id="notes"
              v-model="notes"
              rows="2"
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
              placeholder="Additional notes (optional)"
            />
          </div>

          <!-- Actions -->
          <div class="flex justify-end gap-2 pt-2">
            <button
              type="button"
              @click="emit('close')"
              class="border border-border px-4 py-2 text-sm text-text-secondary hover:bg-bg-muted"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover"
            >
              {{ isEditing ? 'Save' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>
