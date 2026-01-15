import { getAuth } from '../../utils/auth'

export default defineEventHandler(async (event) => {
  const auth = getAuth()
  return auth.handler(toWebRequest(event))
})
