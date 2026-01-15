<script setup lang="ts">
import type { TaskStatus, TaskPriority } from '~/stores/tasks'

const emit = defineEmits<{
  'update:category': [value: string | null]
  'update:status': [value: TaskStatus | null]
  'update:priority': [value: TaskPriority]
  'update:overdue': [value: boolean]
}>()

const categoriesStore = useCategoriesStore()
const { categories } = storeToRefs(categoriesStore)

const selectedCategory = ref<string | null>(null)
const selectedStatus = ref<TaskStatus | null>(null)
const selectedPriority = ref<TaskPriority>(null)
const showOverdueOnly = ref(false)

const statusOptions: { value: TaskStatus | null; label: string }[] = [
  { value: null, label: 'All Status' },
  { value: 'todo', label: 'To Do' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const priorityOptions: { value: TaskPriority; label: string }[] = [
  { value: null, label: 'All Priority' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
]

watch(selectedCategory, (val) => emit('update:category', val))
watch(selectedStatus, (val) => emit('update:status', val))
watch(selectedPriority, (val) => emit('update:priority', val))
watch(showOverdueOnly, (val) => emit('update:overdue', val))

function clearFilters() {
  selectedCategory.value = null
  selectedStatus.value = null
  selectedPriority.value = null
  showOverdueOnly.value = false
}

const hasActiveFilters = computed(() =>
  selectedCategory.value !== null ||
  selectedStatus.value !== null ||
  selectedPriority.value !== null ||
  showOverdueOnly.value
)
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <!-- Category filter -->
    <select
      v-model="selectedCategory"
      class="border border-border bg-bg-surface px-2 py-1 text-xs text-text-primary focus:border-border-focus focus:outline-none"
    >
      <option :value="null">All Categories</option>
      <option
        v-for="cat in categories"
        :key="cat.id"
        :value="cat.id"
      >
        {{ cat.name }}
      </option>
    </select>

    <!-- Status filter -->
    <select
      v-model="selectedStatus"
      class="border border-border bg-bg-surface px-2 py-1 text-xs text-text-primary focus:border-border-focus focus:outline-none"
    >
      <option
        v-for="opt in statusOptions"
        :key="opt.value ?? 'all'"
        :value="opt.value"
      >
        {{ opt.label }}
      </option>
    </select>

    <!-- Priority filter -->
    <select
      v-model="selectedPriority"
      class="border border-border bg-bg-surface px-2 py-1 text-xs text-text-primary focus:border-border-focus focus:outline-none"
    >
      <option
        v-for="opt in priorityOptions"
        :key="opt.value ?? 'all'"
        :value="opt.value"
      >
        {{ opt.label }}
      </option>
    </select>

    <!-- Overdue toggle -->
    <label class="flex items-center gap-1 text-xs text-text-secondary">
      <input
        v-model="showOverdueOnly"
        type="checkbox"
        class="h-3 w-3 accent-accent"
      />
      Overdue only
    </label>

    <!-- Clear filters -->
    <button
      v-if="hasActiveFilters"
      @click="clearFilters"
      class="text-xs text-text-muted hover:text-text-primary"
    >
      Clear
    </button>
  </div>
</template>
