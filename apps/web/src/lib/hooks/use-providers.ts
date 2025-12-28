'use client'

import { useQuery } from '@tanstack/react-query'
import { graphqlClient, Provider, PROVIDERS_QUERY } from '@/lib/api'

interface ProvidersResponse {
  providers: Provider[]
}

export function useProviders() {
  return useQuery({
    queryKey: ['providers'],
    queryFn: async () => {
      const data = await graphqlClient.request<ProvidersResponse>(PROVIDERS_QUERY)
      return data.providers
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}
