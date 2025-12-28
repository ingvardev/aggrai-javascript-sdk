'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Cpu, Cloud, Server, Loader2 } from 'lucide-react'
import { useProviders } from '@/lib/hooks'

const providerIcons: Record<string, typeof Cloud> = {
  OPENAI: Cloud,
  CLAUDE: Cloud,
  OLLAMA: Cpu,
  LOCAL: Server,
}

export function ProviderStatus() {
  const { data: providers, isLoading, error } = useProviders()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Provider Status</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Failed to load providers
          </div>
        ) : providers?.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No providers configured
          </div>
        ) : (
          <div className="space-y-4">
            {providers?.map((provider) => {
              const Icon = providerIcons[provider.type] || Server

              return (
                <div
                  key={provider.id}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                      <Icon className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-sm font-medium">{provider.name}</p>
                      <p className="text-xs text-muted-foreground">
                        Priority: {provider.priority}
                      </p>
                    </div>
                  </div>
                  <Badge variant={provider.enabled ? 'success' : 'secondary'}>
                    {provider.enabled ? 'enabled' : 'disabled'}
                  </Badge>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
