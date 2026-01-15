<script setup lang="ts">
const categoriesStore = useCategoriesStore()
const { categories, isLoading } = storeToRefs(categoriesStore)

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)

const newCategoryName = ref('')
const editingCategory = ref<{ id: string; name: string } | null>(null)
const deletingCategory = ref<{ id: string; name: string; taskCount: number } | null>(null)
const reassignTo = ref('')
const error = ref('')

async function createCategory() {
  if (!newCategoryName.value.trim()) return
  error.value = ''

  try {
    await categoriesStore.createCategory(newCategoryName.value.trim())
    newCategoryName.value = ''
    showCreateModal.value = false
  } catch (e: any) {
    error.value = e.data?.message || 'Failed to create category'
  }
}

function openEditModal(category: { id: string; name: string }) {
  editingCategory.value = { ...category }
  error.value = ''
  showEditModal.value = true
}

async function updateCategory() {
  if (!editingCategory.value || !editingCategory.value.name.trim()) return
  error.value = ''

  try {
    await categoriesStore.updateCategory(
      editingCategory.value.id,
      editingCategory.value.name.trim()
    )
    showEditModal.value = false
    editingCategory.value = null
  } catch (e: any) {
    error.value = e.data?.message || 'Failed to update category'
  }
}

function openDeleteModal(category: { id: string; name: string; taskCount?: number }) {
  deletingCategory.value = {
    id: category.id,
    name: category.name,
    taskCount: category.taskCount || 0,
  }
  reassignTo.value = ''
  error.value = ''
  showDeleteModal.value = true
}

async function deleteCategory() {
  if (!deletingCategory.value) return
  error.value = ''

  try {
    await categoriesStore.deleteCategory(
      deletingCategory.value.id,
      reassignTo.value || undefined
    )
    showDeleteModal.value = false
    deletingCategory.value = null
  } catch (e: any) {
    error.value = e.data?.message || 'Failed to delete category'
  }
}

const otherCategories = computed(() =>
  categories.value.filter((c) => c.id !== deletingCategory.value?.id)
)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-semibold text-text-primary">Categories</h2>
      <button
        @click="showCreateModal = true"
        class="flex items-center gap-1 text-sm text-accent hover:text-accent-hover"
      >
        <Icon name="lucide:plus" class="h-4 w-4" />
        Add
      </button>
    </div>

    <!-- Category list -->
    <div v-if="isLoading" class="py-4 text-center text-text-muted">
      Loading...
    </div>
    <div v-else-if="categories.length === 0" class="py-4 text-center text-text-muted">
      No categories
    </div>
    <ul v-else class="space-y-1">
      <li
        v-for="category in categories"
        :key="category.id"
        class="group flex items-center justify-between border border-border bg-bg-surface px-3 py-2"
      >
        <span class="text-sm text-text-primary">{{ category.name }}</span>
        <div class="flex items-center gap-2">
          <span class="text-xs text-text-muted">{{ category.taskCount || 0 }}</span>
          <div class="flex gap-1 opacity-0 group-hover:opacity-100">
            <button
              @click="openEditModal(category)"
              class="p-1 text-text-muted hover:text-text-primary"
              title="Edit"
            >
              <Icon name="lucide:pencil" class="h-3 w-3" />
            </button>
            <button
              @click="openDeleteModal(category)"
              class="p-1 text-text-muted hover:text-error"
              title="Delete"
            >
              <Icon name="lucide:trash-2" class="h-3 w-3" />
            </button>
          </div>
        </div>
      </li>
    </ul>

    <!-- Create Modal -->
    <Teleport to="body">
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="showCreateModal = false"
      >
        <div class="w-full max-w-md border border-border bg-bg-elevated p-6">
          <h2 class="mb-4 text-lg font-semibold text-text-primary">
            New Category
          </h2>

          <form @submit.prevent="createCategory">
            <div v-if="error" class="mb-4 border border-error bg-error/10 p-3 text-sm text-error">
              {{ error }}
            </div>

            <div class="mb-4">
              <label for="categoryName" class="mb-1 block text-sm text-text-secondary">
                Category name
              </label>
              <input
                id="categoryName"
                v-model="newCategoryName"
                type="text"
                required
                maxlength="100"
                class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
                placeholder="Category name"
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

    <!-- Edit Modal -->
    <Teleport to="body">
      <div
        v-if="showEditModal && editingCategory"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="showEditModal = false"
      >
        <div class="w-full max-w-md border border-border bg-bg-elevated p-6">
          <h2 class="mb-4 text-lg font-semibold text-text-primary">
            Edit Category
          </h2>

          <form @submit.prevent="updateCategory">
            <div v-if="error" class="mb-4 border border-error bg-error/10 p-3 text-sm text-error">
              {{ error }}
            </div>

            <div class="mb-4">
              <label for="editCategoryName" class="mb-1 block text-sm text-text-secondary">
                Category name
              </label>
              <input
                id="editCategoryName"
                v-model="editingCategory.name"
                type="text"
                required
                maxlength="100"
                class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
                autofocus
              />
            </div>

            <div class="flex justify-end gap-2">
              <button
                type="button"
                @click="showEditModal = false"
                class="border border-border px-4 py-2 text-sm text-text-secondary hover:bg-bg-muted"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover"
              >
                Save
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Modal -->
    <Teleport to="body">
      <div
        v-if="showDeleteModal && deletingCategory"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="showDeleteModal = false"
      >
        <div class="w-full max-w-md border border-border bg-bg-elevated p-6">
          <h2 class="mb-4 text-lg font-semibold text-text-primary">
            Delete Category
          </h2>

          <div v-if="error" class="mb-4 border border-error bg-error/10 p-3 text-sm text-error">
            {{ error }}
          </div>

          <p class="mb-4 text-sm text-text-secondary">
            Are you sure you want to delete "<span class="text-text-primary">{{ deletingCategory.name }}</span>"?
          </p>

          <div v-if="deletingCategory.taskCount > 0" class="mb-4">
            <p class="mb-2 text-sm text-warning">
              This category has {{ deletingCategory.taskCount }} task(s). Choose a category to reassign them:
            </p>
            <select
              v-model="reassignTo"
              required
              class="w-full border border-border bg-bg-surface px-3 py-2 text-text-primary focus:border-border-focus focus:outline-none"
            >
              <option value="" disabled>Select category</option>
              <option
                v-for="cat in otherCategories"
                :key="cat.id"
                :value="cat.id"
              >
                {{ cat.name }}
              </option>
            </select>
          </div>

          <div class="flex justify-end gap-2">
            <button
              type="button"
              @click="showDeleteModal = false"
              class="border border-border px-4 py-2 text-sm text-text-secondary hover:bg-bg-muted"
            >
              Cancel
            </button>
            <button
              @click="deleteCategory"
              class="bg-error px-4 py-2 text-sm font-medium text-white hover:bg-error/80"
              :disabled="deletingCategory.taskCount > 0 && !reassignTo"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
