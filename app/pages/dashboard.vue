<script setup lang="ts">
definePageMeta({
  middleware: ['auth'],
})

const auth = useAuth()
const { data: session } = await auth.useSession(useFetch)
const router = useRouter()

async function handleSignOut() {
  await auth.signOut()
  await router.push('/login')
}
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-semibold text-text-primary">Dashboard</h1>
      <button
        @click="handleSignOut"
        class="border border-border px-3 py-1.5 text-sm text-text-secondary hover:bg-bg-muted hover:text-text-primary"
      >
        Sign out
      </button>
    </div>

    <div v-if="session" class="border border-border bg-bg-surface p-4">
      <p class="text-text-secondary">
        Welcome, <span class="text-text-primary">{{ session.user.name }}</span>!
      </p>
      <p class="mt-2 text-sm text-text-muted">
        Your projects will appear here.
      </p>
    </div>
  </div>
</template>
