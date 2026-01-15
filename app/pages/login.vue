<script setup lang="ts">
definePageMeta({
  layout: 'auth',
  middleware: ['auth'],
})

const auth = useAuth()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true

  try {
    const result = await auth.signIn.email({
      email: email.value,
      password: password.value,
    })

    if (result.error) {
      error.value = result.error.message || 'Invalid email or password'
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
    <h1 class="mb-6 text-xl font-semibold text-text-primary">Log in</h1>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div v-if="error" class="border border-error bg-error/10 p-3 text-sm text-error">
        {{ error }}
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
          autocomplete="current-password"
          class="w-full border border-border bg-bg-elevated px-3 py-2 text-text-primary placeholder:text-text-muted focus:border-border-focus focus:outline-none"
          placeholder="Enter your password"
        />
      </div>

      <button
        type="submit"
        :disabled="isLoading"
        class="w-full bg-accent py-2 font-medium text-white hover:bg-accent-hover disabled:opacity-50"
      >
        {{ isLoading ? 'Logging in...' : 'Log in' }}
      </button>
    </form>

    <p class="mt-4 text-center text-sm text-text-secondary">
      Don't have an account?
      <NuxtLink to="/signup" class="text-accent hover:text-accent-hover">
        Sign up
      </NuxtLink>
    </p>
  </div>
</template>
