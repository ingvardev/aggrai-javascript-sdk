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
import { Eye, EyeOff, Copy, RefreshCw, Trash2 } from 'lucide-react'

interface ApiKey {
  id: string
  name: string
  key: string
  lastUsed: string
  createdAt: string
}

const mockApiKeys: ApiKey[] = [
  {
    id: '1',
    name: 'Development Key',
    key: 'dev-api-key-12345',
    lastUsed: '2 hours ago',
    createdAt: 'Jan 15, 2024',
  },
  {
    id: '2',
    name: 'Production Key',
    key: 'prod-api-key-67890',
    lastUsed: '5 minutes ago',
    createdAt: 'Dec 01, 2023',
  },
]

const providerKeys = [
  { id: 'openai', name: 'OpenAI API Key', placeholder: 'sk-...' },
  { id: 'anthropic', name: 'Anthropic API Key', placeholder: 'sk-ant-...' },
  { id: 'ollama', name: 'Ollama URL', placeholder: 'http://localhost:11434' },
]

export function ApiKeysSettings() {
  const [showKeys, setShowKeys] = useState<Record<string, boolean>>({})
  const [apiKeys] = useState(mockApiKeys)

  const toggleShowKey = (id: string) => {
    setShowKeys((prev) => ({ ...prev, [id]: !prev[id] }))
  }

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
  }

  return (
    <div className="space-y-6">
      {/* Your API Keys */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Your API Keys</CardTitle>
              <CardDescription>
                Manage your API keys for accessing the AI Aggregator
              </CardDescription>
            </div>
            <Button size="sm">Generate New Key</Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {apiKeys.map((apiKey) => (
              <div
                key={apiKey.id}
                className="flex items-center justify-between rounded-lg border p-4"
              >
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{apiKey.name}</p>
                    <Badge variant="outline" className="text-xs">
                      Active
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <code className="text-sm text-muted-foreground font-mono">
                      {showKeys[apiKey.id]
                        ? apiKey.key
                        : apiKey.key.slice(0, 8) + '•'.repeat(12)}
                    </code>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => toggleShowKey(apiKey.id)}
                    >
                      {showKeys[apiKey.id] ? (
                        <EyeOff className="h-3 w-3" />
                      ) : (
                        <Eye className="h-3 w-3" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => copyKey(apiKey.key)}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Last used: {apiKey.lastUsed} · Created: {apiKey.createdAt}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button variant="ghost" size="icon">
                    <RefreshCw className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" className="text-destructive">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Provider API Keys */}
      <Card>
        <CardHeader>
          <CardTitle>Provider API Keys</CardTitle>
          <CardDescription>
            Configure API keys for external AI providers
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {providerKeys.map((provider) => (
            <div key={provider.id} className="space-y-2">
              <Label htmlFor={provider.id}>{provider.name}</Label>
              <div className="flex gap-2">
                <Input
                  id={provider.id}
                  type="password"
                  placeholder={provider.placeholder}
                  className="font-mono"
                />
                <Button variant="outline">Test</Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button>Save Changes</Button>
      </div>
    </div>
  )
}
