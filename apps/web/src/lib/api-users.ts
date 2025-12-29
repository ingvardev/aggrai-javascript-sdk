// API Users REST client
// These endpoints use REST API instead of GraphQL

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL?.replace('/graphql', '') || 'http://localhost:8080'

// Types
export interface APIUser {
  id: string
  tenant_id: string
  name: string
  description: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface APIKey {
  id: string
  user_id: string
  key_prefix: string
  name: string
  scopes: string[]
  active: boolean
  expires_at?: string
  last_used_at?: string
  usage_count: number
  created_at: string
  revoked_at?: string
}

export interface APIKeyWithRawKey extends APIKey {
  key: string // Raw key - ONLY SHOWN ONCE
}

interface CreateUserRequest {
  name: string
  description?: string
}

interface CreateKeyRequest {
  user_id: string
  name: string
  scopes?: string[]
}

// API Client class
export class APIUsersClient {
  private baseUrl: string
  private sessionToken: string | null

  constructor(sessionToken: string | null) {
    this.baseUrl = API_BASE_URL
    this.sessionToken = sessionToken
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    if (this.sessionToken) {
      headers['Authorization'] = `Bearer ${this.sessionToken}`
    }

    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.message || error.error || 'Request failed')
    }
    return response.json()
  }

  // Users
  async listUsers(): Promise<APIUser[]> {
    const response = await fetch(`${this.baseUrl}/api/admin/users`, {
      method: 'GET',
      headers: this.getHeaders(),
    })
    return this.handleResponse<APIUser[]>(response)
  }

  async createUser(name: string, description?: string): Promise<APIUser> {
    const body: CreateUserRequest = { name, description }
    const response = await fetch(`${this.baseUrl}/api/admin/users`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(body),
    })
    return this.handleResponse<APIUser>(response)
  }

  // API Keys
  async listKeys(userId: string): Promise<APIKey[]> {
    const response = await fetch(`${this.baseUrl}/api/admin/users/${userId}/api-keys`, {
      method: 'GET',
      headers: this.getHeaders(),
    })
    return this.handleResponse<APIKey[]>(response)
  }

  async createKey(userId: string, name: string, scopes?: string[]): Promise<APIKeyWithRawKey> {
    const body: CreateKeyRequest = { user_id: userId, name, scopes }
    const response = await fetch(`${this.baseUrl}/api/admin/api-keys`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(body),
    })
    return this.handleResponse<APIKeyWithRawKey>(response)
  }

  async revokeKey(keyId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/admin/api-keys/${keyId}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    })
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.message || error.error || 'Failed to revoke key')
    }
  }
}

// Factory function to create client with current session
export function createAPIUsersClient(sessionToken: string | null): APIUsersClient {
  return new APIUsersClient(sessionToken)
}
