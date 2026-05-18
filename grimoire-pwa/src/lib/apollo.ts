import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { getIdToken } from 'firebase/auth'
import { auth } from './firebase'

const httpLink = createHttpLink({
  uri: (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080/query',
})

const authLink = setContext(async (_, { headers }: { headers?: Record<string, string> }) => {
  const token = auth.currentUser ? await getIdToken(auth.currentUser) : null
  return {
    headers: {
      ...headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  }
})

export const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache(),
})
