import { GraphQLClient } from 'graphql-request'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/graphql'

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
