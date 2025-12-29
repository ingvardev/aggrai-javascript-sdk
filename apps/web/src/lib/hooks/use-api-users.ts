'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from './use-auth'
import {
  createAPIUsersClient,
  APIUser,
  APIKey,
  APIKeyWithRawKey,
  ActivityEntry
} from '@/lib/api-users'

// Query keys
export const apiUsersKeys = {
  all: ['apiUsers'] as const,
  lists: () => [...apiUsersKeys.all, 'list'] as const,
  list: () => [...apiUsersKeys.lists()] as const,
  detail: (id: string) => [...apiUsersKeys.all, 'detail', id] as const,
  keys: (userId: string) => [...apiUsersKeys.all, 'keys', userId] as const,
  activity: (userId: string) => [...apiUsersKeys.all, 'activity', userId] as const,
}

// Hooks

export function useAPIUsers() {
  const { sessionToken } = useAuth()

  return useQuery({
    queryKey: apiUsersKeys.list(),
    queryFn: async () => {
      const client = createAPIUsersClient(sessionToken)
      return client.listUsers()
    },
    enabled: !!sessionToken,
  })
}

export function useAPIUser(userId: string) {
  const { data: users, isLoading, error } = useAPIUsers()

  const user = users?.find(u => u.id === userId)

  return {
    data: user,
    isLoading,
    error,
  }
}

export function useCreateAPIUser() {
  const { sessionToken } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ name, description }: { name: string; description?: string }) => {
      const client = createAPIUsersClient(sessionToken)
      return client.createUser(name, description)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiUsersKeys.lists() })
    },
  })
}

// API Keys hooks

export function useAPIKeys(userId: string) {
  const { sessionToken } = useAuth()

  return useQuery({
    queryKey: apiUsersKeys.keys(userId),
    queryFn: async () => {
      const client = createAPIUsersClient(sessionToken)
      return client.listKeys(userId)
    },
    enabled: !!sessionToken && !!userId,
  })
}

export function useCreateAPIKey() {
  const { sessionToken } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      userId,
      name,
      scopes
    }: {
      userId: string
      name: string
      scopes?: string[]
    }): Promise<APIKeyWithRawKey> => {
      const client = createAPIUsersClient(sessionToken)
      return client.createKey(userId, name, scopes)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: apiUsersKeys.keys(variables.userId) })
    },
  })
}

export function useRevokeAPIKey() {
  const { sessionToken } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ keyId, userId }: { keyId: string; userId: string }) => {
      const client = createAPIUsersClient(sessionToken)
      await client.revokeKey(keyId)
      return { keyId, userId }
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: apiUsersKeys.keys(variables.userId) })
    },
  })
}

export function useUserActivity(userId: string) {
  const { sessionToken } = useAuth()

  return useQuery({
    queryKey: apiUsersKeys.activity(userId),
    queryFn: async () => {
      const client = createAPIUsersClient(sessionToken)
      return client.getUserActivity(userId)
    },
    enabled: !!sessionToken && !!userId,
  })
}
