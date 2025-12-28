'use client'

import { createClient, Client } from 'graphql-ws'

let wsClient: Client | null = null

export function getWsClient(): Client {
  if (!wsClient) {
    const apiKey = process.env.NEXT_PUBLIC_API_KEY || 'dev-api-key-12345'
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/graphql'

    wsClient = createClient({
      url: wsUrl,
      connectionParams: {
        'X-API-Key': apiKey,
      },
      retryAttempts: 5,
      shouldRetry: () => true,
      on: {
        connected: () => {
          console.log('[WS] Connected to GraphQL subscriptions')
        },
        closed: () => {
          console.log('[WS] Disconnected from GraphQL subscriptions')
        },
        error: (error) => {
          console.error('[WS] Error:', error)
        },
      },
    })
  }

  return wsClient
}

export function closeWsClient() {
  if (wsClient) {
    wsClient.dispose()
    wsClient = null
  }
}
