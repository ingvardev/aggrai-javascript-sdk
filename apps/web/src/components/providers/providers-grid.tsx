'use client'

import { useState } from 'react'
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
  XCircle,
  Clock,
} from 'lucide-react'

interface Provider {
  id: string
  name: string
  type: 'openai' | 'claude' | 'ollama' | 'local'
  description: string
  enabled: boolean
  status: 'online' | 'offline' | 'checking'
  latency: string
  models: string[]
  totalJobs: number
  successRate: number
}

const providers: Provider[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    type: 'openai',
    description: 'GPT-4, GPT-3.5-turbo, and DALL-E models',
    enabled: true,
    status: 'online',
    latency: '145ms',
    models: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo'],
    totalJobs: 856,
    successRate: 98.2,
  },
  {
    id: 'claude',
    name: 'Anthropic Claude',
    type: 'claude',
    description: 'Claude 3 Opus, Sonnet, and Haiku models',
    enabled: true,
    status: 'online',
    latency: '198ms',
    models: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
    totalJobs: 324,
    successRate: 99.1,
  },
  {
    id: 'ollama',
    name: 'Ollama',
    type: 'ollama',
    description: 'Local LLM runtime with open-source models',
    enabled: false,
    status: 'offline',
    latency: '-',
    models: ['llama3', 'mistral', 'codellama'],
    totalJobs: 42,
    successRate: 95.2,
  },
  {
    id: 'local',
    name: 'Stub Provider',
    type: 'local',
    description: 'Built-in test provider for development',
    enabled: true,
    status: 'online',
    latency: '12ms',
    models: ['stub-model'],
    totalJobs: 1234,
    successRate: 100,
  },
]

const providerIcons = {
  openai: Cloud,
  claude: Cloud,
  ollama: Cpu,
  local: Server,
}

const statusConfig = {
  online: {
    variant: 'success' as const,
    icon: CheckCircle2,
    label: 'Online',
  },
  offline: {
    variant: 'secondary' as const,
    icon: XCircle,
    label: 'Offline',
  },
  checking: {
    variant: 'warning' as const,
    icon: Clock,
    label: 'Checking',
  },
}

export function ProvidersGrid() {
  const [providerStates, setProviderStates] = useState(
    providers.reduce((acc, p) => ({ ...acc, [p.id]: p.enabled }), {} as Record<string, boolean>)
  )

  const toggleProvider = (id: string) => {
    setProviderStates((prev) => ({ ...prev, [id]: !prev[id] }))
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {providers.map((provider) => {
        const Icon = providerIcons[provider.type]
        const statusInfo = statusConfig[provider.status]
        const StatusIcon = statusInfo.icon
        const isEnabled = providerStates[provider.id]

        return (
          <Card key={provider.id} className={!isEnabled ? 'opacity-60' : ''}>
            <CardHeader className="flex flex-row items-start justify-between space-y-0">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
                  <Icon className="h-6 w-6" />
                </div>
                <div>
                  <CardTitle className="text-lg">{provider.name}</CardTitle>
                  <CardDescription className="text-sm">
                    {provider.description}
                  </CardDescription>
                </div>
              </div>
              <Switch
                checked={isEnabled}
                onCheckedChange={() => toggleProvider(provider.id)}
              />
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Status and Latency */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <StatusIcon className="h-4 w-4" />
                  <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
                </div>
                {provider.status === 'online' && (
                  <span className="text-sm text-muted-foreground">
                    Latency: {provider.latency}
                  </span>
                )}
              </div>

              {/* Models */}
              <div>
                <p className="text-xs font-medium text-muted-foreground mb-2">
                  Available Models
                </p>
                <div className="flex flex-wrap gap-1">
                  {provider.models.map((model) => (
                    <Badge key={model} variant="outline" className="text-xs">
                      {model}
                    </Badge>
                  ))}
                </div>
              </div>

              {/* Stats */}
              <div className="flex items-center justify-between border-t pt-4">
                <div>
                  <p className="text-xl font-semibold">{provider.totalJobs}</p>
                  <p className="text-xs text-muted-foreground">Total Jobs</p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-semibold">{provider.successRate}%</p>
                  <p className="text-xs text-muted-foreground">Success Rate</p>
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
