'use client'

import { useQuery } from '@tanstack/react-query'
import { graphqlClient, Tenant, ME_QUERY } from '@/lib/api'

interface MeResponse {
  me: Tenant
}

export function useTenant() {
  return useQuery({
    queryKey: ['tenant', 'me'],
    queryFn: async () => {
      const data = await graphqlClient.request<MeResponse>(ME_QUERY)
      return data.me
    },
    staleTime: 60000, // Cache for 1 minute
  })
}
