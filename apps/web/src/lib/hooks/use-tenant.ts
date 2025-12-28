'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { graphqlClient, Tenant, ME_QUERY, UPDATE_TENANT_MUTATION, TenantSettings, NotificationSettings } from '@/lib/api'

interface MeResponse {
  me: Tenant
}

interface UpdateTenantResponse {
  updateTenant: Tenant
}

interface UpdateTenantInput {
  name?: string
  defaultProvider?: string
  settings?: {
    darkMode?: boolean
    notifications?: Partial<NotificationSettings>
  }
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

export function useUpdateTenant() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: UpdateTenantInput) => {
      const data = await graphqlClient.request<UpdateTenantResponse>(
        UPDATE_TENANT_MUTATION,
        { input }
      )
      return data.updateTenant
    },
    onSuccess: (updatedTenant) => {
      // Update the tenant cache
      queryClient.setQueryData(['tenant', 'me'], updatedTenant)
    },
  })
}
