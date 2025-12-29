import { GraphQLClient } from 'graphql-request'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/graphql'

// Create a function to get the client with current auth token
export function getGraphQLClient(sessionToken?: string | null) {
  const headers: Record<string, string> = {}

  if (sessionToken) {
    headers['Authorization'] = `Bearer ${sessionToken}`
  } else {
    // Fallback to API key for backward compatibility
    headers['X-API-Key'] = 'dev-api-key-12345'
  }

  return new GraphQLClient(API_URL, { headers })
}

// Default client for non-auth requests
export const graphqlClient = new GraphQLClient(API_URL, {
  headers: {
    'X-API-Key': 'dev-api-key-12345',
  },
})

// Job types
export interface Job {
  id: string
  tenantId: string
  type: 'TEXT' | 'IMAGE'
  input: string
  status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'
  result?: string
  error?: string
  provider?: string
  tokensIn: number
  tokensOut: number
  cost: number
  createdAt: string
  updatedAt: string
  startedAt?: string
  finishedAt?: string
}

export interface Provider {
  id: string
  name: string
  type: 'OPENAI' | 'CLAUDE' | 'LOCAL' | 'OLLAMA'
  enabled: boolean
  priority: number
}

export interface ModelInfo {
  id: string
  name: string
  description?: string
  maxTokens?: number
}

export interface UsageSummary {
  provider: string
  totalTokensIn: number
  totalTokensOut: number
  totalCost: number
  jobCount: number
}

// GraphQL queries
export const JOBS_QUERY = `
  query Jobs($filter: JobsFilter, $pagination: PaginationInput) {
    jobs(filter: $filter, pagination: $pagination) {
      edges {
        node {
          id
          type
          input
          status
          result
          error
          provider
          tokensIn
          tokensOut
          cost
          createdAt
          updatedAt
        }
      }
      pageInfo {
        totalCount
        hasNextPage
        hasPreviousPage
      }
    }
  }
`

export const JOB_QUERY = `
  query Job($id: ID!) {
    job(id: $id) {
      id
      type
      input
      status
      result
      error
      provider
      tokensIn
      tokensOut
      cost
      createdAt
      updatedAt
      startedAt
      finishedAt
    }
  }
`

export const PROVIDERS_QUERY = `
  query Providers {
    providers {
      id
      name
      type
      enabled
      priority
    }
  }
`

export const PROVIDER_MODELS_QUERY = `
  query ProviderModels($provider: String!) {
    providerModels(provider: $provider) {
      id
      name
      description
      maxTokens
    }
  }
`

export const USAGE_SUMMARY_QUERY = `
  query UsageSummary {
    usageSummary {
      provider
      totalTokensIn
      totalTokensOut
      totalCost
      jobCount
    }
  }
`

export const USAGE_UPDATED_SUBSCRIPTION = `
  subscription UsageUpdated {
    usageUpdated {
      provider
      totalTokensIn
      totalTokensOut
      totalCost
      jobCount
    }
  }
`

export const CREATE_JOB_MUTATION = `
  mutation CreateJob($input: CreateJobInput!) {
    createJob(input: $input) {
      id
      type
      input
      status
      createdAt
    }
  }
`

// Tenant types
export interface NotificationSettings {
  jobCompleted: boolean
  jobFailed: boolean
  providerOffline: boolean
  usageThreshold: boolean
  weeklySummary: boolean
  marketingEmails: boolean
}

export interface TenantSettings {
  darkMode: boolean
  notifications: NotificationSettings
}

export interface Tenant {
  id: string
  name: string
  active: boolean
  defaultProvider?: string
  settings?: TenantSettings
  createdAt: string
  updatedAt: string
}

export const ME_QUERY = `
  query Me {
    me {
      id
      name
      active
      defaultProvider
      settings {
        darkMode
        notifications {
          jobCompleted
          jobFailed
          providerOffline
          usageThreshold
          weeklySummary
          marketingEmails
        }
      }
      createdAt
      updatedAt
    }
  }
`

export const UPDATE_TENANT_MUTATION = `
  mutation UpdateTenant($input: UpdateTenantInput!) {
    updateTenant(input: $input) {
      id
      name
      active
      defaultProvider
      settings {
        darkMode
        notifications {
          jobCompleted
          jobFailed
          providerOffline
          usageThreshold
          weeklySummary
          marketingEmails
        }
      }
      createdAt
      updatedAt
    }
  }
`

// Pricing types
export interface ProviderPricing {
  id: string
  provider: string
  model: string
  inputPricePerMillion: number
  outputPricePerMillion: number
  imagePrice: number | null
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export const PRICING_LIST_QUERY = `
  query PricingList {
    pricingList {
      id
      provider
      model
      inputPricePerMillion
      outputPricePerMillion
      imagePrice
      isDefault
      createdAt
      updatedAt
    }
  }
`

export const PRICING_BY_PROVIDER_QUERY = `
  query PricingByProvider($provider: String!) {
    pricingByProvider(provider: $provider) {
      id
      provider
      model
      inputPricePerMillion
      outputPricePerMillion
      imagePrice
      isDefault
      createdAt
      updatedAt
    }
  }
`

export const CREATE_PRICING_MUTATION = `
  mutation CreatePricing($input: CreatePricingInput!) {
    createPricing(input: $input) {
      id
      provider
      model
      inputPricePerMillion
      outputPricePerMillion
      imagePrice
      isDefault
      createdAt
      updatedAt
    }
  }
`

export const UPDATE_PRICING_MUTATION = `
  mutation UpdatePricing($id: ID!, $input: UpdatePricingInput!) {
    updatePricing(id: $id, input: $input) {
      id
      provider
      model
      inputPricePerMillion
      outputPricePerMillion
      imagePrice
      isDefault
      createdAt
      updatedAt
    }
  }
`

export const DELETE_PRICING_MUTATION = `
  mutation DeletePricing($id: ID!) {
    deletePricing(id: $id)
  }
`

// ============================================
// Authentication Types & Queries
// ============================================

export type OwnerRole = 'OWNER' | 'ADMIN' | 'MEMBER'

export interface TenantOwner {
  id: string
  tenantId: string
  email: string
  name: string
  role: OwnerRole
  active: boolean
  emailVerified: boolean
  lastLoginAt?: string
  createdAt: string
  updatedAt: string
}

export interface AuthPayload {
  success: boolean
  sessionToken?: string
  owner?: TenantOwner
  tenant?: Tenant
  error?: string
}

export interface Session {
  id: string
  userAgent?: string
  ipAddress?: string
  expiresAt: string
  createdAt: string
}

export const LOGIN_MUTATION = `
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      success
      sessionToken
      owner {
        id
        tenantId
        email
        name
        role
        active
        emailVerified
        lastLoginAt
        createdAt
        updatedAt
      }
      tenant {
        id
        name
        active
      }
      error
    }
  }
`

export const LOGOUT_MUTATION = `
  mutation Logout {
    logout
  }
`

export const LOGOUT_ALL_MUTATION = `
  mutation LogoutAll {
    logoutAll
  }
`

export const REGISTER_MUTATION = `
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      success
      sessionToken
      owner {
        id
        tenantId
        email
        name
        role
        active
      }
      tenant {
        id
        name
        active
      }
      error
    }
  }
`

export const CURRENT_OWNER_QUERY = `
  query CurrentOwner {
    currentOwner {
      id
      tenantId
      email
      name
      role
      active
      emailVerified
      lastLoginAt
      createdAt
      updatedAt
    }
  }
`

export const MY_SESSIONS_QUERY = `
  query MySessions {
    mySessions {
      id
      userAgent
      ipAddress
      expiresAt
      createdAt
    }
  }
`

export const TENANT_OWNERS_QUERY = `
  query TenantOwners {
    tenantOwners {
      id
      tenantId
      email
      name
      role
      active
      emailVerified
      lastLoginAt
      createdAt
      updatedAt
    }
  }
`

export const CREATE_OWNER_MUTATION = `
  mutation CreateOwner($input: CreateOwnerInput!) {
    createOwner(input: $input) {
      id
      tenantId
      email
      name
      role
      active
      createdAt
    }
  }
`

export const UPDATE_OWNER_MUTATION = `
  mutation UpdateOwner($id: ID!, $input: UpdateOwnerInput!) {
    updateOwner(id: $id, input: $input) {
      id
      tenantId
      email
      name
      role
      active
      updatedAt
    }
  }
`

export const DELETE_OWNER_MUTATION = `
  mutation DeleteOwner($id: ID!) {
    deleteOwner(id: $id)
  }
`

export const CHANGE_PASSWORD_MUTATION = `
  mutation ChangePassword($input: ChangePasswordInput!) {
    changePassword(input: $input)
  }
`
