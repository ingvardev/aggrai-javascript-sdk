'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  graphqlClient,
  ProviderPricing,
  PRICING_LIST_QUERY,
  PRICING_BY_PROVIDER_QUERY,
  CREATE_PRICING_MUTATION,
  UPDATE_PRICING_MUTATION,
  DELETE_PRICING_MUTATION,
} from '@/lib/api'

interface PricingListResponse {
  pricingList: ProviderPricing[]
}

interface PricingByProviderResponse {
  pricingByProvider: ProviderPricing[]
}

interface CreatePricingResponse {
  createPricing: ProviderPricing
}

interface UpdatePricingResponse {
  updatePricing: ProviderPricing
}

interface DeletePricingResponse {
  deletePricing: boolean
}

interface CreatePricingInput {
  provider: string
  model: string
  inputPricePerMillion: number
  outputPricePerMillion: number
  imagePrice?: number | null
  isDefault?: boolean
}

interface UpdatePricingInput {
  inputPricePerMillion?: number
  outputPricePerMillion?: number
  imagePrice?: number | null
  isDefault?: boolean
}

export function usePricingList() {
  return useQuery({
    queryKey: ['pricing', 'list'],
    queryFn: async () => {
      const data = await graphqlClient.request<PricingListResponse>(PRICING_LIST_QUERY)
      return data.pricingList
    },
    staleTime: 60000, // Cache for 1 minute
  })
}

export function usePricingByProvider(provider: string) {
  return useQuery({
    queryKey: ['pricing', 'provider', provider],
    queryFn: async () => {
      const data = await graphqlClient.request<PricingByProviderResponse>(
        PRICING_BY_PROVIDER_QUERY,
        { provider }
      )
      return data.pricingByProvider
    },
    enabled: !!provider,
    staleTime: 60000,
  })
}

export function useCreatePricing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: CreatePricingInput) => {
      const data = await graphqlClient.request<CreatePricingResponse>(
        CREATE_PRICING_MUTATION,
        { input }
      )
      return data.createPricing
    },
    onSuccess: () => {
      // Invalidate pricing queries to refetch
      queryClient.invalidateQueries({ queryKey: ['pricing'] })
    },
  })
}

export function useUpdatePricing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, ...input }: { id: string } & UpdatePricingInput) => {
      const data = await graphqlClient.request<UpdatePricingResponse>(
        UPDATE_PRICING_MUTATION,
        { id, input }
      )
      return data.updatePricing
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pricing'] })
    },
  })
}

export function useDeletePricing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const data = await graphqlClient.request<DeletePricingResponse>(
        DELETE_PRICING_MUTATION,
        { id }
      )
      return data.deletePricing
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pricing'] })
    },
  })
}
