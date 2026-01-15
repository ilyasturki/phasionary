import { getAuth } from '../utils/auth'

export default defineEventHandler(async (event) => {
  // Skip auth for non-API routes and the auth endpoint itself
  const path = getRequestURL(event).pathname
  if (!path.startsWith('/api/') || path.startsWith('/api/auth/')) {
    return
  }

  const auth = getAuth()
  const session = await auth.api.getSession({
    headers: getHeaders(event),
  })

  if (!session) {
    throw createError({
      statusCode: 401,
      message: 'Unauthorized',
    })
  }

  // Attach session to event context for use in API routes
  event.context.session = session
})
