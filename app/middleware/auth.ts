export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuth()
  const { data: session } = await auth.useSession(useFetch)

  // Public routes that don't require authentication
  const publicRoutes = ['/', '/login', '/signup']

  if (!session.value && !publicRoutes.includes(to.path)) {
    return navigateTo('/login')
  }

  // Redirect authenticated users away from auth pages
  if (session.value && (to.path === '/login' || to.path === '/signup')) {
    return navigateTo('/dashboard')
  }
})
