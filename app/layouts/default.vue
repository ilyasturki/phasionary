<script setup lang="ts">
const auth = useAuth()
const { data: session, isPending } = await auth.useSession(useFetch)
</script>

<template>
  <div class="min-h-screen bg-bg-base font-mono text-text-primary">
    <!-- Header -->
    <header class="sticky top-0 z-40 border-b border-border bg-bg-surface">
      <div class="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6">
        <NuxtLink to="/" class="text-lg font-semibold tracking-tight">
          Phasionary
        </NuxtLink>

        <nav v-if="session" class="flex items-center gap-2 sm:gap-4">
          <NuxtLink
            to="/dashboard"
            class="text-sm text-text-secondary hover:text-text-primary"
          >
            Dashboard
          </NuxtLink>
          <span
            v-if="session?.user?.email"
            class="hidden text-sm text-text-muted sm:inline"
          >
            {{ session.user.email }}
          </span>
        </nav>

        <nav v-else-if="!isPending" class="flex items-center gap-2 sm:gap-4">
          <NuxtLink
            to="/login"
            class="text-sm text-text-secondary hover:text-text-primary"
          >
            Log in
          </NuxtLink>
          <NuxtLink
            to="/signup"
            class="bg-accent px-3 py-1.5 text-sm font-medium text-white hover:bg-accent-hover"
          >
            Sign up
          </NuxtLink>
        </nav>
      </div>
    </header>

    <!-- Main content -->
    <main class="mx-auto max-w-7xl px-4 py-4 sm:px-6 sm:py-6">
      <slot />
    </main>
  </div>
</template>
