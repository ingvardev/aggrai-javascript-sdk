'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Eye, EyeOff, Copy, RefreshCw, Trash2, Info } from 'lucide-react'
import { useProviders } from '@/lib/hooks'
import { toast } from 'sonner'

interface ApiKey {
  id: string
  name: string
  key: string
  lastUsed: string
  createdAt: string
}

// Current API key from environment
const CURRENT_API_KEY = 'dev-api-key-12345'

const providerKeys = [
  { id: 'openai', name: 'OpenAI API Key', placeholder: 'sk-...', envVar: 'OPENAI_API_KEY' },
  { id: 'anthropic', name: 'Anthropic API Key', placeholder: 'sk-ant-...', envVar: 'ANTHROPIC_API_KEY' },
  { id: 'ollama', name: 'Ollama URL', placeholder: 'http://localhost:11434', envVar: 'OLLAMA_URL' },
]

export function ApiKeysSettings() {
  const { data: providers, isLoading: providersLoading } = useProviders()
  const [showKeys, setShowKeys] = useState<Record<string, boolean>>({})
  const [showCurrentKey, setShowCurrentKey] = useState(false)

  const toggleShowKey = (id: string) => {
    setShowKeys((prev) => ({ ...prev, [id]: !prev[id] }))
  }

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
    toast.success('Copied to clipboard')
  }

  return (
    <div className="space-y-6">
      {/* Current API Key */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Current API Key</CardTitle>
              <CardDescription>
                Your API key for accessing the AI Aggregator API
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              API key management will be available in a future update. Currently using a development key.
            </AlertDescription>
          </Alert>

          <div className="mt-4 flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <p className="font-medium">Development Key</p>
                <Badge variant="outline" className="text-xs">
                  Active
                </Badge>
              </div>
              <div className="flex items-center gap-2">
                <code className="text-sm text-muted-foreground font-mono">
                  {showCurrentKey
                    ? CURRENT_API_KEY
                    : CURRENT_API_KEY.slice(0, 8) + '•'.repeat(12)}
                </code>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={() => setShowCurrentKey(!showCurrentKey)}
                >
                  {showCurrentKey ? (
                    <EyeOff className="h-3 w-3" />
                  ) : (
                    <Eye className="h-3 w-3" />
                  )}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={() => copyKey(CURRENT_API_KEY)}
                >
                  <Copy className="h-3 w-3" />
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Use header: X-API-Key: {showCurrentKey ? CURRENT_API_KEY : '••••••••'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Provider Status */}
      <Card>
        <CardHeader>
          <CardTitle>Provider Status</CardTitle>
          <CardDescription>
            Current status of configured AI providers
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {providersLoading ? (
              <p className="text-sm text-muted-foreground">Loading providers...</p>
            ) : providers && providers.length > 0 ? (
              providers.map((provider) => (
                <div key={provider.name} className="flex items-center justify-between rounded-lg border p-3">
                  <div className="flex items-center gap-3">
                    <div className={`h-2 w-2 rounded-full ${provider.enabled ? 'bg-green-500' : 'bg-red-500'}`} />
                    <div>
                      <p className="font-medium">{provider.name}</p>
                      <p className="text-xs text-muted-foreground capitalize">{provider.type}</p>
                    </div>
                  </div>
                  <Badge variant={provider.enabled ? 'default' : 'secondary'}>
                    {provider.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">No providers configured</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Provider API Keys Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Provider API Keys</CardTitle>
          <CardDescription>
            API keys are configured via environment variables on the server
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {providerKeys.map((provider) => (
            <div key={provider.id} className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="font-medium">{provider.name}</p>
                <p className="text-xs text-muted-foreground font-mono">{provider.envVar}</p>
              </div>
              <Badge variant="secondary">Server Config</Badge>
            </div>
          ))}
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              Provider API keys are managed through environment variables for security.
              See the project documentation for configuration details.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    </div>
  )
}
