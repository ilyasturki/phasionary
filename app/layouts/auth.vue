<template>
  <div class="flex min-h-screen items-center justify-center bg-bg-base font-mono">
    <div class="w-full max-w-md">
      <div class="mb-8 text-center">
        <NuxtLink to="/" class="text-2xl font-semibold tracking-tight text-text-primary">
          Phasionary
        </NuxtLink>
      </div>
      <div class="border border-border bg-bg-surface p-6">
        <slot />
      </div>
    </div>
  </div>
</template>
