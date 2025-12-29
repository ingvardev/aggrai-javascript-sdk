'use client'

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getGraphQLClient,
  TenantOwner,
  Tenant,
  AuthPayload,
  LOGIN_MUTATION,
  LOGOUT_MUTATION,
  CURRENT_OWNER_QUERY,
} from '@/lib/api'

// Session token storage
const SESSION_TOKEN_KEY = 'session_token'

function getStoredToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(SESSION_TOKEN_KEY)
}

function setStoredToken(token: string | null): void {
  if (typeof window === 'undefined') return
  if (token) {
    localStorage.setItem(SESSION_TOKEN_KEY, token)
  } else {
    localStorage.removeItem(SESSION_TOKEN_KEY)
  }
}

interface AuthContextType {
  owner: TenantOwner | null
  tenant: Tenant | null
  sessionToken: string | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>
  logout: () => Promise<void>
  refetch: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const [sessionToken, setSessionToken] = useState<string | null>(null)
  const [isInitialized, setIsInitialized] = useState(false)

  // Initialize token from storage
  useEffect(() => {
    const token = getStoredToken()
    setSessionToken(token)
    setIsInitialized(true)
  }, [])

  // Fetch current owner
  const {
    data: ownerData,
    isLoading: isLoadingOwner,
    refetch,
  } = useQuery({
    queryKey: ['currentOwner', sessionToken],
    queryFn: async () => {
      if (!sessionToken) return null
      const client = getGraphQLClient(sessionToken)
      const result = await client.request<{ currentOwner: TenantOwner | null }>(
        CURRENT_OWNER_QUERY
      )
      return result.currentOwner
    },
    enabled: isInitialized && !!sessionToken,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: async ({ email, password }: { email: string; password: string }) => {
      const client = getGraphQLClient()
      const result = await client.request<{ login: AuthPayload }>(LOGIN_MUTATION, {
        input: { email, password },
      })
      return result.login
    },
    onSuccess: (data) => {
      if (data.success && data.sessionToken) {
        setSessionToken(data.sessionToken)
        setStoredToken(data.sessionToken)
        queryClient.invalidateQueries({ queryKey: ['currentOwner'] })
      }
    },
  })

  // Logout mutation
  const logoutMutation = useMutation({
    mutationFn: async () => {
      if (!sessionToken) return
      const client = getGraphQLClient(sessionToken)
      await client.request(LOGOUT_MUTATION)
    },
    onSettled: () => {
      setSessionToken(null)
      setStoredToken(null)
      queryClient.clear()
    },
  })

  const login = useCallback(
    async (email: string, password: string) => {
      try {
        const result = await loginMutation.mutateAsync({ email, password })
        if (result.success) {
          return { success: true }
        }
        return { success: false, error: result.error || 'Login failed' }
      } catch (error) {
        return { success: false, error: 'An error occurred during login' }
      }
    },
    [loginMutation]
  )

  const logout = useCallback(async () => {
    await logoutMutation.mutateAsync()
  }, [logoutMutation])

  const value: AuthContextType = {
    owner: ownerData ?? null,
    tenant: null, // TODO: fetch tenant data if needed
    sessionToken,
    isLoading: !isInitialized || isLoadingOwner,
    isAuthenticated: !!ownerData,
    login,
    logout,
    refetch,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
