<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: ['auth'],
})

const auth = useAuth()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true

  try {
    const result = await auth.signUp.email({
      name: name.value,
      email: email.value,
      password: password.value,
    })

    if (result.error) {
      if (result.error.message?.includes('exist')) {
        error.value = 'An account with this email already exists'
      } else if (result.error.message?.includes('password')) {
        error.value = 'Password must be at least 8 characters'
      } else {
        error.value = result.error.message || 'Failed to create account'
      }
    } else {
      await router.push('/dashboard')
    }
  } catch (e) {
    error.value = 'Unable to connect. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="mb-6 text-xl font-semibold text-text-primary">Create account</h1>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div v-if="error" class="border border-error bg-error/10 p-3 text-sm text-error">
        {{ error }}
      </div>

      <div>
        <label for="name" class="mb-1 block text-sm text-text-secondary">
          Name
        </label>
        <input
          id="name"
          v-model="name"
          type="text"
          required
          autocomplete="name"
          class="w-full border border-border bg-bg-elevated px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
          placeholder="Your name"
        />
      </div>

      <div>
        <label for="email" class="mb-1 block text-sm text-text-secondary">
          Email
        </label>
        <input
          id="email"
          v-model="email"
          type="email"
          required
          autocomplete="email"
          class="w-full border border-border bg-bg-elevated px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
          placeholder="you@example.com"
        />
      </div>

      <div>
        <label for="password" class="mb-1 block text-sm text-text-secondary">
          Password
        </label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          minlength="8"
          autocomplete="new-password"
          class="w-full border border-border bg-bg-elevated px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
          placeholder="At least 8 characters"
        />
      </div>

      <button
        type="submit"
        :disabled="isLoading"
        class="w-full bg-accent py-2 font-medium text-white hover:bg-accent-hover disabled:opacity-50"
      >
        {{ isLoading ? 'Creating account...' : 'Create account' }}
      </button>
    </form>

    <p class="mt-4 text-center text-sm text-text-secondary">
      Already have an account?
      <NuxtLink to="/login" class="text-accent hover:text-accent-hover">
        Log in
      </NuxtLink>
    </p>
  </div>
</template>
