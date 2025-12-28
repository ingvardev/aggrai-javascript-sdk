'use client'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import {
  Cloud,
  Server,
  Cpu,
  Settings,
  CheckCircle2,
  Loader2,
} from 'lucide-react'
import { useProviders, useUsageSummary } from '@/lib/hooks'

const providerIcons: Record<string, typeof Cloud> = {
  OPENAI: Cloud,
  CLAUDE: Cloud,
  OLLAMA: Cpu,
  LOCAL: Server,
}

const providerDescriptions: Record<string, string> = {
  OPENAI: 'GPT-4, GPT-3.5-turbo, and DALL-E models',
  CLAUDE: 'Claude 3 Opus, Sonnet, and Haiku models',
  OLLAMA: 'Local LLM runtime with open-source models',
  LOCAL: 'Built-in test provider for development',
}

export function ProvidersGrid() {
  const { data: providers, isLoading, error } = useProviders()
  const { data: usageSummary } = useUsageSummary()

  // Get usage stats per provider
  const getProviderStats = (providerName: string) => {
    const usage = usageSummary?.find(
      (u) => u.provider.toLowerCase() === providerName.toLowerCase()
    )
    return {
      totalJobs: usage?.jobCount || 0,
      totalCost: usage?.totalCost || 0,
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        Failed to load providers. Is the API running?
      </div>
    )
  }

  if (!providers?.length) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        No providers configured
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {providers.map((provider) => {
        const Icon = providerIcons[provider.type] || Server
        const stats = getProviderStats(provider.name)

        return (
          <Card key={provider.id} className={!provider.enabled ? 'opacity-60' : ''}>
            <CardHeader className="flex flex-row items-start justify-between space-y-0">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
                  <Icon className="h-6 w-6" />
                </div>
                <div>
                  <CardTitle className="text-lg">{provider.name}</CardTitle>
                  <CardDescription className="text-sm">
                    {providerDescriptions[provider.type] || `${provider.type} provider`}
                  </CardDescription>
                </div>
              </div>
              <Switch checked={provider.enabled} disabled />
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Status */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  <Badge variant={provider.enabled ? 'success' : 'secondary'}>
                    {provider.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <span className="text-sm text-muted-foreground">
                  Priority: {provider.priority}
                </span>
              </div>

              {/* Stats */}
              <div className="flex items-center justify-between border-t pt-4">
                <div>
                  <p className="text-xl font-semibold">{stats.totalJobs}</p>
                  <p className="text-xs text-muted-foreground">Total Jobs</p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-semibold">
                    ${stats.totalCost.toFixed(4)}
                  </p>
                  <p className="text-xs text-muted-foreground">Total Cost</p>
                </div>
                <Button variant="ghost" size="sm">
                  <Settings className="h-4 w-4 mr-2" />
                  Configure
                </Button>
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
