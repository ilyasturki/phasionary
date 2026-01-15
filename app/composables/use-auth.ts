import { createAuthClient } from 'better-auth/vue'

let _authClient: ReturnType<typeof createAuthClient> | null = null

export function useAuth() {
  if (!_authClient) {
    const config = useRuntimeConfig()
    _authClient = createAuthClient({
      baseURL: config.public.appUrl || undefined,
    })
  }
  return _authClient
}
