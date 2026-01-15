<script setup lang="ts">
definePageMeta({
  middleware: ['auth'],
})

const auth = useAuth()
const { data: session } = await auth.useSession(useFetch)
</script>

<template>
  <div class="py-12 text-center">
    <h1 class="mb-4 text-3xl font-semibold text-text-primary">
      Task-based Project Planning
    </h1>
    <p class="mb-8 text-text-secondary">
      A focused way to organize, track, and complete work items across projects.
    </p>

    <div v-if="session">
      <NuxtLink
        to="/dashboard"
        class="inline-block bg-accent px-6 py-2 font-medium text-white hover:bg-accent-hover"
      >
        Go to Dashboard
      </NuxtLink>
    </div>
    <div v-else class="flex items-center justify-center gap-4">
      <NuxtLink
        to="/signup"
        class="inline-block bg-accent px-6 py-2 font-medium text-white hover:bg-accent-hover"
      >
        Get Started
      </NuxtLink>
      <NuxtLink
        to="/login"
        class="inline-block border border-border px-6 py-2 font-medium text-text-primary hover:bg-bg-muted"
      >
        Log in
      </NuxtLink>
    </div>
  </div>
</template>
