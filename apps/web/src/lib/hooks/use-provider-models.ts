'use client'

import { useQuery } from '@tanstack/react-query'
import { graphqlClient, ModelInfo, PROVIDER_MODELS_QUERY } from '@/lib/api'

interface ProviderModelsResponse {
  providerModels: ModelInfo[]
}

export function useProviderModels(provider: string | null) {
  return useQuery({
    queryKey: ['providerModels', provider],
    queryFn: async () => {
      if (!provider) return []
      const data = await graphqlClient.request<ProviderModelsResponse>(
        PROVIDER_MODELS_QUERY,
        { provider }
      )
      return data.providerModels
    },
    enabled: !!provider,
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  })
}
